package pkg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/cwxstat/go-pod-launch-run/pkg/vscode"
	"github.com/emicklei/go-restful/v3/log"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1Inter "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
	blog "log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// go get k8s.io/client-go@v0.26.3
var timeout int64 = 60

type SPDYExecutorFactory interface {
	NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error)
}

type defaultSPDYExecutorFactory struct{}

func (f *defaultSPDYExecutorFactory) NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	//return spdy.NewSPDYExecutor(config, method, url)
	return remotecommand.NewSPDYExecutor(config, method, url)
}

func B(i SPDYExecutorFactory) {
	fmt.Println(i)
}

func do() {
	B(&defaultSPDYExecutorFactory{})
}

func Run(podName,
	namespace,
	containerName,
	serviceAccountName string,
	vscodeDebug bool,
	commands []string, output string) error {

	clientset, err := getClientset()
	if err != nil {
		panic(err)
	}

	if vscodeDebug {
		log.Printf("vscodeDebug: %v", vscodeDebug)
		commands = vscode.CommandsVscode()
	} else if len(commands) == 0 {
		commands = []string{"aws configure list", "aws sts get-caller-identity"}
	}

	//podName := "aws-cli-pod"
	//namespace := "default"
	//containerName := "aws-cli"

	// Launch the Pod
	pod, err := createPod(clientset.CoreV1(), namespace, podName, containerName, serviceAccountName)
	if err != nil {
		//fmt.Println("Failed to create Pod: ", err.Error())
		if strings.Contains(err.Error(), "already exists") {
			if promptAndConfirm(fmt.Sprintf("Pod %s already exists. Do you want to delete it?\n", podName)) {
				err = deletePod(clientset.CoreV1(), namespace, podName)
				return err
			}
			return err
		}
		return err
	}
	fmt.Println("Pod created successfully.", pod.Status.Phase)

	var wg sync.WaitGroup
	wg.Add(1)

	rc, err := NewRestConfig()

	// Run commands in separate goroutine
	go func() {
		defer wg.Done()

		// Wait for Pod to be running
		err = waitForPodRunning(clientset.CoreV1(), namespace, podName)
		if err != nil {
			panic(err)
		}

		// Execute the commands and write the output to a file

		err = rc.execCommandsInPod(clientset.CoreV1(),
			&defaultSPDYExecutorFactory{},
			namespace, podName, containerName, commands, output)
		if err != nil {
			log.Printf("Failed to execute commands in Pod: %v", err)
		} else {
			fmt.Println("Commands executed successfully. Output written to result.pod.")
		}

	}()

	wg.Wait()

	if vscodeDebug {
		fmt.Println("vscode debug")
		fmt.Println("kubectl exec -it", podName, "-n", namespace, "--container", containerName, "--", "bash")
		fmt.Println("kubectl port-forward", podName, "8080:8080", "-n", namespace)
		fmt.Println("code-server&")
		fmt.Println("cat ~/.config/code-server/config.yaml")
		fmt.Println("")
		fmt.Println("http://localhost:8080")
		fmt.Println("common commands:")
		fmt.Println("aws configure list")
		fmt.Println("aws sts get-caller-identity")
		fmt.Println("")
		fmt.Println("Additional installs:")
		fmt.Println("yum groupinstall -y \"Development Tools\"")
		fmt.Println("yum install -y python3-devel")
		fmt.Println("yum install -y bind-utils")
		fmt.Println("")
		fmt.Println("When you're done, run the following command to delete the pod:")
		fmt.Println("kubectl delete pod", podName, "-n", namespace, "--grace-period=0 --force")
		return nil
	}

	// Delete the Pod
	err = deletePod(clientset.CoreV1(), namespace, podName)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pod deleted successfully.")

	return nil
}

func getClientset() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	// Check if the program is running inside a Kubernetes cluster.
	if _, err = rest.InClusterConfig(); err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}

		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
		if err != nil {
			blog.Fatalf("Failed to load Kubernetes configuration: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			blog.Fatalf("Failed to load in-cluster configuration: %v", err)
		}
	}

	return kubernetes.NewForConfig(config)

}

func createPod(clientsetCoreV1 v1Inter.CoreV1Interface, namespace, podName, containerName,
	serviceAccountName string) (*v1.Pod,
	error) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: serviceAccountName,
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

	return clientsetCoreV1.Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
}

func waitForPodRunning(clientsetCoreV1 v1Inter.CoreV1Interface, namespace, podName string) error {
	for {
		pod, err := clientsetCoreV1.Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
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

func deletePod(clientsetCoreV1 v1Inter.CoreV1Interface, namespace, podName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := clientsetCoreV1.Pods(namespace).Delete(context.Background(), podName, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete pod %s in namespace %s: %v", podName, namespace, err)
	}

	err = waitForPodDeletion(clientsetCoreV1, namespace, podName, &timeout)
	if err != nil {
		return fmt.Errorf("failed to wait for pod %s in namespace %s to be deleted: %v", podName, namespace, err)
	}

	return nil
}

func waitForPodDeletion(clientsetCoreV1 v1Inter.CoreV1Interface, namespace, podName string, timeout *int64) error {

	var watchOptions = metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", podName),
		TimeoutSeconds: timeout,
		Watch:          true,
	}
	watcher, err := clientsetCoreV1.Pods(namespace).Watch(context.Background(), watchOptions)
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

type Config struct {
	restConfig *rest.Config
}

func NewRestConfig() (*Config, error) {
	// Note: You need result config to be able to connect to the cluster from outside
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		restConfig: restconfig,
	}, nil
}

func (c *Config) execCommandsInPod(clientsetCoreV1 v1Inter.CoreV1Interface,
	icmd SPDYExecutorFactory,
	namespace,
	podName,
	containerName string,
	commands []string, outputFile string) error {
	var outputBuffer bytes.Buffer
	var outputErrorBuffer bytes.Buffer

	for _, cmd := range commands {
		req := clientsetCoreV1.RESTClient().Post().
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

		//executor, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
		executor, err := icmd.NewSPDYExecutor(c.restConfig, "POST", req.URL())
		if err != nil {
			return err
		}

		var cmdOutputBuffer bytes.Buffer
		var cmdStderrBuffer bytes.Buffer
		err = executor.Stream(remotecommand.StreamOptions{
			Stdout: &cmdOutputBuffer,
			Stderr: &cmdStderrBuffer,
			Tty:    false,
		})

		if err != nil {
			fmt.Errorf("failed to execute command %s: %v", cmd, err)
			return err
		}

		outputBuffer.Write(cmdOutputBuffer.Bytes())
		outputBuffer.WriteString("\n")

		outputErrorBuffer.Write(cmdStderrBuffer.Bytes())
		outputErrorBuffer.WriteString("\n")

	}

	err := os.WriteFile(outputFile, outputBuffer.Bytes(), 0644)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s%s", outputFile, ".err"), outputErrorBuffer.Bytes(), 0644)
	return err
}

func promptAndConfirm(prompt string) bool {
	fmt.Printf("%s [y/n]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		os.Exit(1)
	}
	if strings.ToLower(strings.TrimSpace(input)) == "y" {
		return true
	}
	return false
}
