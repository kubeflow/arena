package training

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/config"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tensorboardURL(name, namespace string, services []*v1.Service, nodes []*v1.Node) (url string, err error) {

	var (
		port int32
	)
	var service *v1.Service
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
	if service.Spec.Type == v1.ServiceTypeLoadBalancer {
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			return fmt.Sprintf("http://%s:%d",
				service.Status.LoadBalancer.Ingress[0].IP,
				service.Spec.Ports[0].Port), nil
		}
	}

	port = portList[0].NodePort

	// 2. Get address
	var node *v1.Node

	for _, n := range nodes {
		if isNodeReady(*n) {
			node = n
			break
		}
	}

	if node == nil {
		return "", fmt.Errorf("Failed to find the ready node for exporting tensorboard.")
	}
	url = fmt.Sprintf("http://%s:%d", getNodeInternalAddress(*node), port)

	return url, nil
}

func isNodeReady(node v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func getNodeInternalAddress(node v1.Node) string {
	address := "unknown"
	if len(node.Status.Addresses) > 0 {
		//address = nodeInfo.node.Status.Addresses[0].Address
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				address = addr.Address
			}
		}
	}
	return address
}

func PrepareServicesAndNodesForTensorboard(jobs []TrainingJob, allNamespaces bool) ([]*v1.Service, []*v1.Node) {
	clientset := config.GetArenaConfiger().GetClientSet()
	services := []*v1.Service{}
	nodes := []*v1.Node{}
	var err error
	if len(jobs) == 0 {
		return services, nodes
	}
	labels := map[string]string{
		"release": "",
		"role":    "tensorboard",
	}
	if len(jobs) == 1 {
		labels["release"] = jobs[0].Name()
	}
	namespace := jobs[0].Namespace()
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}
	services, err = listServices(clientset, namespace, labels)
	if err != nil {
		log.Errorf("failed to list k8s services when query dashboard url,reason: %v", err)
		services = []*v1.Service{}
	}
	nodes, err = listNodes(clientset, map[string]string{})
	if err != nil {
		log.Errorf("failed to list nodes when query dashboard url,reason: %v", err)
		nodes = []*v1.Node{}
	}
	return services, nodes
}
