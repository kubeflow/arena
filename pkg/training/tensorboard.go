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

package training

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/k8saccesser"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tensorboardURL(name, namespace string, services []*corev1.Service, nodes []*corev1.Node) (url string, err error) {

	var (
		port int32
	)
	var service *corev1.Service
	for _, svc := range services {
		if svc.Labels["release"] == name && namespace == svc.Namespace {
			service = svc
			break
		}
	}
	if service == nil {
		log.Debugf("Failed to find the tensorboard service due to service"+
			"List is empty when selector is release=%s,role=tensorboard.", name)
		return "", nil
	}
	portList := service.Spec.Ports
	if len(portList) == 0 {
		log.Debugf("Failed to find the tensorboard service due to ports list is empty.")
		return "", nil
	}

	// Get Address for loadbalancer
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			return fmt.Sprintf("http://%s:%d",
				service.Status.LoadBalancer.Ingress[0].IP,
				service.Spec.Ports[0].Port), nil
		}
	}

	port = portList[0].NodePort

	// 2. Get address
	var node *corev1.Node

	for _, n := range nodes {
		if isNodeReady(*n) {
			node = n
			break
		}
	}

	if node == nil {
		return "", fmt.Errorf("failed to find the ready node for exporting tensorboard")
	}
	url = fmt.Sprintf("http://%s:%d", getNodeInternalAddress(*node), port)

	return url, nil
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func getNodeInternalAddress(node corev1.Node) string {
	address := "unknown"
	if len(node.Status.Addresses) > 0 {
		//address = nodeInfo.node.Status.Addresses[0].Address
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				address = addr.Address
			}
		}
	}
	return address
}

func PrepareServicesAndNodesForTensorboard(jobs []TrainingJob, allNamespaces bool) ([]*corev1.Service, []*corev1.Node) {
	services := []*corev1.Service{}
	nodes := []*corev1.Node{}
	var err error
	if len(jobs) == 0 {
		return services, nodes
	}
	labelSelector := "role=tensorboard,release"
	if len(jobs) == 1 {
		labelSelector = fmt.Sprintf("role=tensorboard,release=%v", jobs[0].Name())
	}
	namespace := jobs[0].Namespace()
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}
	services, err = k8saccesser.GetK8sResourceAccesser().ListServices(namespace, labelSelector)
	if err != nil {
		log.Errorf("failed to list k8s services when query dashboard url,reason: %v", err)
		services = []*corev1.Service{}
	}
	nodes, err = k8saccesser.GetK8sResourceAccesser().ListNodes("")
	if err != nil {
		log.Errorf("failed to list nodes when query dashboard url,reason: %v", err)
		nodes = []*corev1.Node{}
	}
	return services, nodes
}
