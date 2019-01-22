package workflow

import (
	"os"

	"github.com/kubeflow/arena/util/helm"
	"github.com/kubeflow/arena/util/kubectl"
	log "github.com/sirupsen/logrus"
)

/**
*	install application
**/

func SubmitJob(name string, trainingType string, namespace string, values interface{}, chart string) error {
	// 1. Generate value file
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return err
	}

	// 2. Keep value file in configmap
	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return err
	}

	err = kubectl.CreateConfigmap(name, trainingType, namespace, valueFileName, chartName, chartVersion)
	if err != nil {
		return err
	}

	// 3. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, chartName)
	if err != nil {
		return err
	}

	// 4. Create Application
	err = kubectl.InstallApps(name, template)
	if err != nil {
		return err
	}

	if log.GetLevel() != log.DebugLevel {
		err = os.Remove(valueFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", valueFileName, err)
		}

		err = os.Remove(template)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", template, err)
		}
	}

	return err
}
