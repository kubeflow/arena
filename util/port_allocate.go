package util

import (
	"k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
	log "github.com/sirupsen/logrus"
	"fmt"
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
		if ! isPortInUsed(port) {
			setUsedPort(port)
			return port, nil
		}
		port ++
	}
	return 0, fmt.Errorf("failed to select a available port")
}

func initK8sClusterUsedPort(client *kubernetes.Clientset) ([]int, error) {
	if len(k8sClusterUsedPorts) == 0 {
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
	pods, err := client.CoreV1().Pods("").List(meta_v1.ListOptions{})
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, pod := range pods.Items {
		// fileter pod
		if excludeInactivePod(&pod) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				usedHostPort := port.HostPort
				if pod.Spec.HostNetwork {
					usedHostPort = port.ContainerPort
				}

				k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(usedHostPort))
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

// exclude Inactive pod when compute ports
func excludeInactivePod(pod *v1.Pod) bool {
	// pod not assigned
	if len(pod.Spec.NodeName) == 0 {
		return true
	}
	// pod is Successed or failed
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
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
