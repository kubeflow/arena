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
		configName := fmt.Sprintf("%s-%s", name, trainingType)
		found := kubectl.CheckAppConfigMap(configName, namespace)
		if found {
			cms = append(cms, trainingType)
		}
	}

	return cms
}
