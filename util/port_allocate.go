package util

import (
	"k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
	log "github.com/sirupsen/logrus"
	"fmt"
)

var k8sClusterUsedPorts = []int{}

const AUTO_SELECT_PORT_MIN = 30000
const AUTO_SELECT_PORT_MAX = 50000

// If default value is available, use it
// else select port automatically
func SelectAvailablePortWithDefault(client *kubernetes.Clientset, defaultPort int) (int, error) {
	if len(k8sClusterUsedPorts) == 0 {
		var err error
		k8sClusterUsedPorts, err = GetK8sUsedNodePorts(client)
		if err != nil {
			return 0, err
		}
	}
	if !isPortInUse(defaultPort) {
		return defaultPort, nil
	}
	return SelectAvailablePort(client)
}

// Select a available port in range (AUTO_SELECT_PORT_MIN ~ AUTO_SELECT_PORT_MAX), and exclude used ports in k8s
// if 30000 is selected this time, make sure next time it will select 30001
func SelectAvailablePort(client *kubernetes.Clientset) (int, error) {
	var err error
	if len(k8sClusterUsedPorts) == 0 {
		k8sClusterUsedPorts, err = GetK8sUsedNodePorts(client)
	}

	if err != nil {
		return 0, err
	}
	port := AUTO_SELECT_PORT_MIN
	for port < AUTO_SELECT_PORT_MAX {
		if ! isPortInUse(port) {
			setUsedPort(port)
			return port, nil
		}
		port ++
	}
	return 0, fmt.Errorf("failed to select a available port")
}

// Gather used node ports for k8s cluster
// 1. HostNetwork pod's HostPort
// 2. NodePort / Loadbalancer Service's NodePort

func GetK8sUsedNodePorts(client *kubernetes.Clientset) ([]int, error) {
	pods, err := client.CoreV1().Pods("").List(meta_v1.ListOptions{})
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, pod := range pods.Items {
		if pod.Spec.HostNetwork {
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(port.HostPort))
				}
			}
		}
	}

	services, err := client.CoreV1().Services("").List(meta_v1.ListOptions{})
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, service := range services.Items {
		if service.Spec.Type == v1.ServiceTypeNodePort || service.Spec.Type == v1.ServiceTypeLoadBalancer {
			for _, port := range service.Spec.Ports {
				k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(port.NodePort))
			}
		}
	}
	log.Debug("Get K8S used ports, %++v", k8sClusterUsedPorts)
	return k8sClusterUsedPorts, nil
}

func isPortInUse(port int) bool {
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