package pkg

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestCreatePod(t *testing.T) {
	// Set up a fake clientset for simulating a Kubernetes cluster
	clientset := fake.NewSimpleClientset()

	// Define test input parameters
	namespace := "test-namespace"
	podName := "test-pod"
	containerName := "test-container"
	serviceAccountName := "test-service-account"

	// Test createPod function
	createdPod, err := createPod(clientset.CoreV1(), namespace, podName, containerName, serviceAccountName)
	if err == nil {
		t.Logf("Created pod: %v\n", createdPod.Name)
		t.Logf("  namespace: %v\n", createdPod.Namespace)
		t.Logf("  ServiceAccountName: %v\n", createdPod.Spec.ServiceAccountName)
	}
	fmt.Println(createdPod)
	assert.NoError(t, err, "createPod should not return an error")

	// Check if the pod was created correctly
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, podName, pod.ObjectMeta.Name, "Pod name should match the input")
	assert.Equal(t, namespace, pod.ObjectMeta.Namespace, "Pod namespace should match the input")
	assert.Equal(t, serviceAccountName, pod.Spec.ServiceAccountName, "ServiceAccountName should match the input")
	assert.Equal(t, corev1.RestartPolicyNever, pod.Spec.RestartPolicy, "RestartPolicy should be 'Never'")

	// Check the container in the pod
	assert.Len(t, pod.Spec.Containers, 1, "There should be exactly 1 container")
	container := pod.Spec.Containers[0]
	assert.Equal(t, containerName, container.Name, "Container name should match the input")
	assert.Equal(t, "amazon/aws-cli:latest", container.Image, "Container image should be 'amazon/aws-cli:latest'")
	assert.Equal(t, []string{"sleep", "3600"}, container.Command, "Container command should be 'sleep 3600'")
}

func TestWaitForPodDeletion(t *testing.T) {
	// Set up a fake clientset for simulating a Kubernetes cluster
	clientset := fake.NewSimpleClientset()

	// Define test input parameters
	namespace := "test-namespace"
	podName := "test-pod"
	timeout := int64(5)

	// Create a pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
	}
	_, err := clientset.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Delete the pod asynchronously after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		err := clientset.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
		assert.NoError(t, err)
	}()

	// Test waitForPodDeletion function
	err = waitForPodDeletion(clientset.CoreV1(), namespace, podName, &timeout)
	assert.NoError(t, err, "waitForPodDeletion should not return an error if the pod is deleted within the timeout")

	// Create a new pod
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
	}
	_, err = clientset.CoreV1().Pods(namespace).Create(context.Background(), pod2, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Set a shorter timeout for the test to fail
	shortTimeout := int64(1)

	// Test waitForPodDeletion function with shorter timeout
	err = waitForPodDeletion(clientset.CoreV1(), namespace, podName, &shortTimeout)
	assert.Error(t, err, "waitForPodDeletion should return an error if the pod is not deleted within the timeout")
}

//func TestExecCommandsInPod(t *testing.T) {
//	// Set up a fake clientset for simulating a Kubernetes cluster
//	clientset := fake.NewSimpleClientset()
//
//	// Define test input parameters
//	namespace := "test-namespace"
//	podName := "test-pod"
//	containerName := "test-container"
//	commands := []string{"echo 'Hello, world!'"}
//	outputFile := "test_output.txt"
//
//	// Execute the function with the fake clientset and test input parameters
//	err := execCommandsInPod(clientset.CoreV1()), namespace, podName, containerName, commands, outputFile)
//	assert.NoError(t, err)
//
//	// Check the output file for correct content
//	content, err := ioutil.ReadFile(outputFile)
//	assert.NoError(t, err)
//
//	expectedContent := []byte("Hello, world!\n")
//	assert.Equal(t, expectedContent, content)
//
//	// Clean up the output file
//	err = os.Remove(outputFile)
//	assert.NoError(t, err)
//}
