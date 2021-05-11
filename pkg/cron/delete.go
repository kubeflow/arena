package cron

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	"strings"
)

func DeleteCron(name, namespace, jobType string) error {
	err := GetCronHandler().DeleteCron(namespace, name)
	if err != nil {
		return err
	}
	out := fmt.Sprintf("cron %s has deleted", name)
	fmt.Println(out)

	configMapName := fmt.Sprintf("%s-%s", name, strings.ToLower(jobType))
	return kubectl.DeleteAppConfigMap(configMapName, namespace)
}
