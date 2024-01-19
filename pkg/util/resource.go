package util

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	allServices = map[string][]v1.Service{}
	allPods     = map[string][]v1.Pod{}
)

func AcquireAllPods(namespace string, client *kubernetes.Clientset) ([]v1.Pod, error) {
	if podsCache, ok := allPods[namespace]; ok {
		return podsCache, nil
	}
	pods := []v1.Pod{}
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return pods, err
	}
	pods = append(pods, podList.Items...)
	allPods[namespace] = pods
	log.Debugf("Pods in %s: %++v", namespace, allPods[namespace])
	return pods, nil
}

func AcquireServingServices(namespace string, client *kubernetes.Clientset) ([]v1.Service, error) {
	if serviceCache, ok := allServices[namespace]; ok {
		return serviceCache, nil
	}
	serviceList, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "servingName",
	})
	if err != nil {
		log.Errorf("Failed to list services due to %v", err)
		return []v1.Service{}, err
	}
	allServices[namespace] = serviceList.Items
	log.Debugf("Services in %s: %++v", namespace, allServices[namespace])
	return allServices[namespace], nil
}
