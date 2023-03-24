package main

import (
	"context"
	"fmt"
	"log"
	"os"

	//	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func main() {
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
			log.Fatalf("Failed to load Kubernetes configuration: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Failed to load in-cluster configuration: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	listPods(clientset)
}

func listPods(clientset *kubernetes.Clientset) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}

	fmt.Println("Listing all pods in the cluster:")
	for _, pod := range pods.Items {
		fmt.Printf("Namespace: %s, Name: %s, Status: %s\n", pod.Namespace, pod.Name, pod.Status.Phase)
	}
}
