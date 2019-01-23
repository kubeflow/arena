package workflow

import (
	"os"

	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
)

/**
*	delete training job
**/

func DeleteJob(name, namespace string) error {
	appInfoFilename, err := kubectl.SaveAppConfigMapToFile(jobName, "app", namespace)
	if err != nil {
		return err
	}

	err = kubectl.UninstallAppsWithAppInfoFile(appInfoFilename, namespace)
	if err != nil {
		return err
	}
}

/**
*	Submit training job
**/

func SubmitJob(name string, trainingType string, namespace string, values interface{}, chart string) error {
	// 1. Generate value file
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return err
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, chart)
	if err != nil {
		return err
	}

	// 3. Generate AppInfo file
	AppInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
		return err
	}

	// 4. Keep value file in configmap
	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return err
	}

	err = kubectl.CreateAppConfigmap(name,
		trainingType,
		namespace,
		valueFileName,
		AppInfoFileName,
		chartName,
		chartVersion)

	if err != nil {
		return err
	}

	// 5. Delete Application
	err = kubectl.DeleteAppConfigMap(name, namespace)
	if err != nil {
		log.Debugf("Ignore delete configmap %s's error %v", name, err)
	}

	// 6. Create Application
	err = kubectl.InstallApps(template, namespace)
	if err != nil {
		// clean configmap
		log.Infof("clean up the config map %s because creating application failed.", name)
		kubectl.DeleteAppConfigMap(name, namespace)
		return err
	}

	// 7. Clean up the template file
	if log.GetLevel() != log.DebugLevel {
		err = os.Remove(valueFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", valueFileName, err)
		}

		err = os.Remove(template)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", template, err)
		}

		err = os.Remove(AppInfoFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", AppInfoFileName, err)
		}
	}

	return nil
}
