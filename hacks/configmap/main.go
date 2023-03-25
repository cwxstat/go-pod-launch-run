package main

import (
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	createConfigMap(clientset)
}

func createConfigMap(clientset *kubernetes.Clientset) {
	namespace := "argo"
	configMapName := "spike"
	configMapData := map[string]string{
		"example-key": "example-value",
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: configMapData,
	}

	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create ConfigMap: %v", err)
	}

	fmt.Printf("Successfully created ConfigMap %q in namespace %q\n", configMapName, namespace)
}
