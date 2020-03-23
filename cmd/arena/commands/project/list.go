package project

import (
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/command"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"os"
	"strconv"
	"text/tabwriter"
)

var (
	queueResource = schema.GroupVersionResource{
		Group:    "scheduling.incubator.k8s.io",
		Version:  "v1alpha1",
		Resource: "queues",
	}
)

type ProjectInfo struct {
	name         string
	deservedGPUs string
}

type Queue struct {
	Spec struct {
		DeservedGpus int `mapstructure:"deservedGpus,omitempty"`
	} `mapstructure:"spec,omitempty"`
	Metadata struct {
		Name string `mapstructure:"name,omitempty"`
	} `mapstructure:"metadata,omitempty"`
}

func runListCommand(cmd *cobra.Command, args []string) error {
	clientset, err := client.GetClientSet()
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

	projects := make(map[string]*ProjectInfo)

	for _, namespace := range namespaceList.Items {
		if namespace.Labels == nil {
			continue
		}

		runaiQueue := namespace.Labels["runai/queue"]

		if runaiQueue != "" {
			projects[runaiQueue] = &ProjectInfo{
				name: namespace.Name,
			}
		}
	}

	clientConfig, err := client.GetClientConfig()
	if err != nil {
		return err
	}
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	queueList, err := dynamicClient.Resource(queueResource).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, queueItem := range queueList.Items {
		var queue Queue
		if err := mapstructure.Decode(queueItem.Object, &queue); err != nil {
			return err
		}

		if project, found := projects[queue.Metadata.Name]; found {
			project.deservedGPUs = strconv.Itoa(queue.Spec.DeservedGpus)
		}
	}

	printProjects(projects)

	return nil
}

func printProjects(infos map[string]*ProjectInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	util.PrintLine(w, "NAME", "DESERVED GPUs")

	for _, info := range infos {
		deservedInfo := "deleted"

		if info.deservedGPUs != "" {
			deservedInfo = info.deservedGPUs
		}

		util.PrintLine(w, info.name, deservedInfo)
	}

	_ = w.Flush()
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
