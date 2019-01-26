package commands

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/util/kubectl"
)

/*
* get App Configs by name, which is created by arena
 */
func getTrainingTypes(name, namespace string) (cms []string) {
	cms = []string{}
	for _, trainingType := range knownTrainingTypes {
		found := isTrainingConfigExist(name, trainingType, namespace)
		if found {
			cms = append(cms, trainingType)
		}
	}

	return cms
}

/**
*  check if the training config exist
 */
func isTrainingConfigExist(name, trainingType, namespace string) bool {
	configName := fmt.Sprintf("%s-%s", name, trainingType)
	return kubectl.CheckAppConfigMap(configName, namespace)
}
