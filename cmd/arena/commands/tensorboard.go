// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"path"
	"strings"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type submitTensorboardArgs struct {
	UseTensorboard   bool   `yaml:"useTensorboard"`   // --tensorboard
	TensorboardImage string `yaml:"tensorboardImage"` // --tensorboardImage
	TrainingLogdir   string `yaml:"trainingLogdir"`   // --logdir
	HostLogPath      string `yaml:"hostLogPath"`
	IsLocalLogging   bool   `yaml:"isLocalLogging"`
}

func (submitArgs *submitTensorboardArgs) processTensorboard(dataMap map[string]string) {
	if submitArgs.UseTensorboard {
		log.Debugf("dataMap %v", dataMap)
		if path.IsAbs(submitArgs.TrainingLogdir) && !submitArgs.isLoggingInPVC(dataMap) {
			// Need to consider pvc
			submitArgs.HostLogPath = fmt.Sprintf("/arena_logs/training%s", util.RandomInt32())
			submitArgs.IsLocalLogging = true
		} else {
			// doing nothing for hdfs path
			log.Debugf("Doing nothing for logging Path %s", submitArgs.TrainingLogdir)
			submitArgs.IsLocalLogging = false
		}
	}
}

// check if the path in the pvc
func (submitArgs *submitTensorboardArgs) isLoggingInPVC(dataMap map[string]string) (inPVC bool) {
	for pvc, path := range dataMap {
		if strings.HasPrefix(submitArgs.TrainingLogdir, path) {
			log.Debugf("Log path %s is contained by pvc %s's path %s", submitArgs.TrainingLogdir, pvc, path)
			inPVC = true
			break
		} else {
			log.Debugf("Log path %s is not contained by pvc %s's path %s", submitArgs.TrainingLogdir, pvc, path)
		}
	}

	return inPVC
}

func tensorboardURL(name, namespace string, clientset kubernetes.Interface) (url string, err error) {

	var (
		port int32
	)

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
