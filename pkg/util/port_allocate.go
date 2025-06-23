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
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var k8sClusterUsedPorts = []int{}

const AUTO_SELECT_PORT_MIN = 20000
const AUTO_SELECT_PORT_MAX = 30000

// If default port is available, use it
// If not set defaultPort, select port automatically
func SelectAvailablePortWithDefault(client *kubernetes.Clientset, port int) (int, error) {
	// if set port, return the port
	if port != 0 {
		return port, nil
	}
	return SelectAvailablePort(client)
}

// Select a available port in range (AUTO_SELECT_PORT_MIN ~ AUTO_SELECT_PORT_MAX), and exclude used ports in k8s
// if 20000 is selected this time, make sure next time it will select 20001
func SelectAvailablePort(client *kubernetes.Clientset) (int, error) {
	if _, err := initK8sClusterUsedPort(client); err != nil {
		return 0, err
	}
	port := AUTO_SELECT_PORT_MIN
	for port < AUTO_SELECT_PORT_MAX {
		if !isPortInUsed(port) {
			setUsedPort(port)
			return port, nil
		}
		port++
	}
	return 0, fmt.Errorf("failed to select a available port")
}

func initK8sClusterUsedPort(client *kubernetes.Clientset) ([]int, error) {
	maxUsedPort := AUTO_SELECT_PORT_MAX - AUTO_SELECT_PORT_MIN
	if len(k8sClusterUsedPorts) == 0 || len(k8sClusterUsedPorts) >= maxUsedPort {
		var err error
		k8sClusterUsedPorts, err = getClusterUsedNodePorts(client)
		if err != nil {
			return k8sClusterUsedPorts, err
		}
	}
	return k8sClusterUsedPorts, nil
}

// Gather used node ports for k8s cluster
// 1. HostNetwork pod's HostPort
// 2. NodePort / Loadbalancer Service's NodePort
func getClusterUsedNodePorts(client *kubernetes.Clientset) ([]int, error) {
	k8sClusterUsedPorts = []int{}
	pods, err := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, pod := range pods.Items {
		// filter pod
		if excludeInactivePod(&pod) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				usedHostPort := port.HostPort
				if pod.Spec.HostNetwork {
					usedHostPort = port.ContainerPort
				}

				if int(usedHostPort) >= AUTO_SELECT_PORT_MIN && int(usedHostPort) < AUTO_SELECT_PORT_MAX {
					k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(usedHostPort))
				}
			}
		}
	}

	services, err := client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, service := range services.Items {
		if service.Spec.Type == corev1.ServiceTypeNodePort || service.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, port := range service.Spec.Ports {
				if int(port.NodePort) >= AUTO_SELECT_PORT_MIN && int(port.NodePort) < AUTO_SELECT_PORT_MAX {
					k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(port.NodePort))
				}
			}
		}
	}
	log.Debugf("Get K8S used ports, %++v", k8sClusterUsedPorts)
	return k8sClusterUsedPorts, nil
}

// exclude Inactive pod when compute ports
func excludeInactivePod(pod *corev1.Pod) bool {
	// pod not assigned
	if len(pod.Spec.NodeName) == 0 {
		return true
	}
	// pod is Successed or failed
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return true
	}
	return false
}

func isPortInUsed(port int) bool {
	for _, usedPort := range k8sClusterUsedPorts {
		if port == usedPort {
			return true
		}
	}
	return false
}

func setUsedPort(port int) {
	k8sClusterUsedPorts = append(k8sClusterUsedPorts, port)
}
