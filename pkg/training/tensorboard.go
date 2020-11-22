package training

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/config"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tensorboardURL(name, namespace string) (url string, err error) {

	var (
		port int32
	)
	clientset := config.GetArenaConfiger().GetClientSet()
	// 1. Get port
	serviceList, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,role=tensorboard", name),
	})
	if err != nil {
		// if errors.IsNotFound(err) {
		// 	log.Debugf("The tensorboard service doesn't exist")
		// 	return "", nil
		// }else{
		// 	return "", err
		// }
		return "", err
	}

	if len(serviceList.Items) == 0 {
		log.Debugf("Failed to find the tensorboard service due to service"+
			"List is empty when selector is release=%s,role=tensorboard.", name)
		return "", nil
	}

	service := serviceList.Items[0]
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
	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	node := v1.Node{}
	findReadyNode := false

	for _, item := range nodeList.Items {
		if isNodeReady(item) {
			node = item
			findReadyNode = true
			break
		}
	}

	if !findReadyNode {
		return "", fmt.Errorf("Failed to find the ready node for exporting tensorboard.")
	}
	url = fmt.Sprintf("http://%s:%d", getNodeInternalAddress(node), port)

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
