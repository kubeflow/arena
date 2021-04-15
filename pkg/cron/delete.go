package cron

import (
	"context"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"strings"
)

func DeleteCron(name, namespace, jobType string) error {
	config := config.GetArenaConfiger().GetRestConfig()

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	err = dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	out := fmt.Sprintf("cron %s deleted", name)
	fmt.Println(out)

	configMapName := fmt.Sprintf("%s-%s", name, strings.ToLower(jobType))
	return kubectl.DeleteAppConfigMap(configMapName, namespace)
}
