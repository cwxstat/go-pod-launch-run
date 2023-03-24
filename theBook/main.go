package main

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"


	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func main() {
	fmt.Println("here")
	// Set up the Kubernetes client configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Define the Pod and container where the command will be executed
	namespace := "default"
	podName := "dev"
	containerName := "dev"

	// Define the command to execute
	cmd := []string{"sh", "-c", "echo 'Hello, world!' > /tmp/hello.txt"}

	fmt.Println("here")
	// Create a request to execute the command
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName).
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	// Create an SPDY executor
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		panic(err)
	}

	// Execute the command in the container
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Command executed successfully")
}
