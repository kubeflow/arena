// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
