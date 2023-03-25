package main

import (
	"context"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var err error

	clientset, err := getClientset()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	err = createConfigMap(clientset.CoreV1(), "argo", "spike2", map[string]string{
		"example-key": "example-value",
	})
	if err != nil {
		log.Fatalf("Failed to create ConfigMap: %v", err)
	}
}

func getClientset() (*kubernetes.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func createConfigMap(clientsetCoreV1 v1.CoreV1Interface, namespace, configMapName string,
	configMapData map[string]string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: configMapData,
	}

	_, err := clientsetCoreV1.ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	return err
}
