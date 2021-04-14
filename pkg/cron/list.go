package cron

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func ListCronTask(namespace string, allNamespaces bool) ([]*types.CronInfo, error) {
	config := config.GetArenaConfiger().GetRestConfig()

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	gvr := schema.GroupVersionResource{
		Group:    "machinelearning.seldon.io",
		Version:  "v1",
		Resource: "seldondeployments",
	}

	list, err := dynamicClient.Resource(gvr).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	fmt.Println(list.GetKind())

	return nil, nil
}

func DisplayAllCronTasks(tasks []*types.CronInfo, allNamespace bool, format types.FormatStyle) {
	//TODO
}
