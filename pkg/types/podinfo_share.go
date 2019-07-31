package types

import (
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"time"
)

var (
	retries   = 5
	clientset *kubernetes.Clientset
)

type podInfo struct {
	name      string
	namespace string
}

func GetActivePodsInAllNodes() ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
	})

	for i := 0; i < retries && err != nil; i++ {
		pods, err = clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
			LabelSelector: labels.Everything().String(),
		})
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return []v1.Pod{}, fmt.Errorf("failed to get Pods")
	}
	return filterActivePods(pods.Items), nil
}

func filterActivePods(pods []v1.Pod) (activePods []v1.Pod) {
	activePods = []v1.Pod{}
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}

		activePods = append(activePods, pod)
	}

	return activePods
}

func GetAllSharedGPUNode() ([]v1.Node, error) {
	nodes := []v1.Node{}
	allNodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nodes, err
	}

	for _, item := range allNodes.Items {
		if IsGPUSharingNode(item) {
			nodes = append(nodes, item)
		}
	}

	return nodes, nil
}

func gpuMemoryInPod(pod v1.Pod) int {
	var total int
	containers := pod.Spec.Containers
	for _, container := range containers {
		if val, ok := container.Resources.Limits[resourceName]; ok {
			total += int(val.Value())
		}
	}

	return total
}
