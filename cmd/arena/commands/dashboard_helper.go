package commands

import (
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
