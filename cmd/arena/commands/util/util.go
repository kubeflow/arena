package util

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaceFromProjectName(project string, kubeClient *client.Client) (string, error) {
	namespaceList, err := kubeClient.GetClientset().CoreV1().Namespaces().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", RUNAI_QUEUE_LABEL, project),
	})

	if err != nil {
		return "", err
	}

	if namespaceList != nil && len(namespaceList.Items) != 0 {
		return namespaceList.Items[0].Name, nil
	} else {
		return "", fmt.Errorf("project %s was not found. Please run '%s project list' to view all avaliable projects", project, config.CLIName)
	}
}
