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
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"fmt"
)

func dashboard(client kubernetes.Interface, namespace string, name string) (string, error) {
	// podList, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "ListOptions",
	// 		APIVersion: "v1",
	// 	}, LabelSelector: fmt.Sprintf("release=%s", name),
	// })

	url, err := dashboardFromLoadbalancer(client, namespace, name)
	if err != nil {
		logrus.Debugf("Failed to find the dashboard entry in the loadbalancer from %s in namespace %s due to %v",
			name,
			namespace,
			err)
	} else if len(url) > 0 {
		return url, nil
	}

	ep, err := client.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// for _, subset := range ep.Subsets{
	// 	adresses := subset.Addresses
	// }

	if len(ep.Subsets) < 1 {
		return "", fmt.Errorf("Failed to find endpoint for dashboard %s in namespace %s", name, namespace)
	}

	subset := ep.Subsets[0]
	if len(subset.Addresses) < 1 {
		return "", fmt.Errorf("Failed to find address for dashboard %s in namespace %s", name, namespace)
	}

	if len(subset.Ports) < 1 {
		return "", fmt.Errorf("Failed to find port for dashboard %s in namespace %s", name, namespace)
	}

	port := subset.Ports[0].Port
	ip := subset.Addresses[0].IP

	// return podList.Items, err
	return fmt.Sprintf("%s:%d", ip, port), nil
}

func GetJobDashboards(dashboard string, job *v1.Job, pods []corev1.Pod) []string {
	urls := []string{}
	for _, pod := range pods {
		meta := pod.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				break
			}
		}

		// Only print the job logs
		if isJob {
			spec := pod.Spec
			url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
				dashboard,
				job.Namespace,
				pod.Name,
				spec.Containers[0].Name,
				job.Namespace)

			urls = append(urls, url)
		}
	}
	return urls
}

// Get dashboard url if it's load balancer
func dashboardFromLoadbalancer(client kubernetes.Interface, namespace string, name string) (string, error) {
	svc, err := client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				address := svc.Status.LoadBalancer.Ingress[0].IP
				port := svc.Spec.Ports[0].Port
				return fmt.Sprintf("%s:%d", address, port), nil
			}
		}
	}

	return "", fmt.Errorf("Ignore non load balacner svc for %s")
}
