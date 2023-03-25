package main

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateConfigMap(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	namespace := "argo"
	configMapName := "spike"
	configMapData := map[string]string{
		"example-key": "example-value",
	}

	err := createConfigMap(clientset.CoreV1(), namespace, configMapName, configMapData)
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get ConfigMap: %v", err)
	}

	if configMap.Name != configMapName {
		t.Errorf("Expected ConfigMap name %q, but got %q", configMapName, configMap.Name)
	}

	for key, value := range configMapData {
		if configMap.Data[key] != value {
			t.Errorf("Expected ConfigMap data %q for key %q, but got %q", value, key, configMap.Data[key])
		}
	}
}
