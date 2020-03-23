package project

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/command"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// "k8s.io/client-go/dynamic"
)

// const (
// 	queueResource = schema.GroupVersionResource{
// 		Group:    "",
// 		Version:  "v1",
// 		Resource: "Queues",
// 	}
// )

func runListCommand(cmd *cobra.Command, args []string) error {
	clientset, err := util.GetClientSet()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	namespaceList, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		return err
	}

	namespaces := []string{}
	for _, namespace := range namespaceList.Items {
		if namespace.Labels != nil && namespace.Labels["runai/queue"] != "" {
			namespaces = append(namespaces, namespace.Name)
		}
	}

	fmt.Println("Namespace %v", namespaces)

	return nil
}

func newListProjectsCommand() *cobra.Command {
	commandWrapper := command.NewCommandWrapper(runListCommand)

	var command = &cobra.Command{
		Use:   "list",
		Short: "List all avaliable projects",
		Run:   commandWrapper.Run,
	}

	return command
}
