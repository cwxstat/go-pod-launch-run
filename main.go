package main

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// go get k8s.io/client-go@v0.26.3
var timeout int64 = 60

func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	podName := "aws-cli-pod"
	namespace := "default"
	containerName := "aws-cli"

	// Launch the Pod
	pod, err := createPod(clientset, namespace, podName, containerName)
	if err != nil {
		panic(err)
	}
	fmt.Println("Pod created successfully.", pod.Status)

	var wg sync.WaitGroup
	wg.Add(1)

	// Run commands in separate goroutine
	go func() {
		defer wg.Done()

		// Wait for Pod to be running
		err = waitForPodRunning(clientset, namespace, podName)
		if err != nil {
			panic(err)
		}

		// Execute the commands and write the output to a file
		err = execCommandsInPod(clientset, namespace, podName, containerName, []string{"aws configure list", "aws sts get-caller-identity"}, "result.pod")
		if err != nil {
			panic(err)
		}

		fmt.Println("Commands executed successfully. Output written to result.pod.")
	}()

	wg.Wait()

	// Delete the Pod
	err = deletePod(clientset, namespace, podName)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pod deleted successfully.")
}

func createPod(clientset *kubernetes.Clientset, namespace, podName, containerName string) (*v1.Pod, error) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "default",
			Containers: []v1.Container{
				{
					Name:    containerName,
					Image:   "amazon/aws-cli:latest",
					Command: []string{"sleep", "3600"},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	return clientset.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
}

func waitForPodRunning(clientset *kubernetes.Clientset, namespace, podName string) error {
	for {
		pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Status.Phase == v1.PodRunning {
			break
		} else if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			return fmt.Errorf("pod %s in namespace %s failed to start, current status: %v", podName, namespace, pod.Status.Phase)
		}

		time.Sleep(1 * time.Second)
	}
	return nil
}

func deletePod(clientset *kubernetes.Clientset, namespace, podName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := clientset.CoreV1().Pods(namespace).Delete(context.Background(), podName, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete pod %s in namespace %s: %v", podName, namespace, err)
	}

	err = waitForPodDeletion(clientset, namespace, podName, &timeout)
	if err != nil {
		return fmt.Errorf("failed to wait for pod %s in namespace %s to be deleted: %v", podName, namespace, err)
	}

	return nil
}

func waitForPodDeletion(clientset *kubernetes.Clientset, namespace, podName string, timeout *int64) error {

	var watchOptions = metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", podName),
		TimeoutSeconds: timeout,
		Watch:          true,
	}
	watcher, err := clientset.CoreV1().Pods(namespace).Watch(context.Background(), watchOptions)
	if err != nil {
		return err
	}

	defer watcher.Stop()

	fmt.Println("Waiting for pod deletion...")
	ch := watcher.ResultChan()
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return fmt.Errorf("watch channel closed")
			}

			switch event.Type {
			case watch.Deleted:
				return nil
			case watch.Error:
				return fmt.Errorf("watch error: %v", event.Object)
			}
		case <-time.After(time.Duration(*timeout) * time.Second):
			return fmt.Errorf("timeout waiting for pod deletion")
		}
	}
}
func execCommandsInPod(clientset *kubernetes.Clientset, namespace, podName, containerName string, commands []string, outputFile string) error {
	var outputBuffer bytes.Buffer

	// Note: You need result config to be able to connect to the cluster from outside
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return err
	}

	for _, cmd := range commands {
		req := clientset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(podName).
			Namespace(namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: containerName,
				Command:   []string{"/bin/sh", "-c", cmd},
				Stdout:    true,
				Stderr:    true,
			}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
		if err != nil {
			return err
		}

		var cmdOutputBuffer bytes.Buffer
		err = executor.Stream(remotecommand.StreamOptions{
			Stdout: &cmdOutputBuffer,
			Stderr: os.Stderr,
			Tty:    false,
		})

		if err != nil {
			return err
		}

		outputBuffer.Write(cmdOutputBuffer.Bytes())
		outputBuffer.WriteString("\n")
	}

	return os.WriteFile(outputFile, outputBuffer.Bytes(), 0644)
}
