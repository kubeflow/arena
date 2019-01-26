package workflow

import (
	"fmt"
	"os"

	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
)

/**
*	delete training job with the job name
**/

func DeleteJob(name, namespace, trainingType string) error {
	jobName := fmt.Sprintf("%s-%s", name, trainingType)

	appInfoFilename, err := kubectl.SaveAppConfigMapToFile(jobName, "app", namespace)
	if err != nil {
		log.Debugf("Failed to SaveAppConfigMapToFile due to %v", err)
		return err
	}

	err = kubectl.UninstallAppsWithAppInfoFile(appInfoFilename, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
		return err
	}

	err = kubectl.DeleteAppConfigMap(jobName, namespace)
	if err != nil {
		log.Warningf("Delete configmap %s failed, please clean it manually due to %v.", jobName, err)
		log.Warningf("Please run `kubectl delete -n %s cm %s`", namespace, jobName)
	}

	return nil
}

/**
*	Submit training job
**/

func SubmitJob(name string, trainingType string, namespace string, values interface{}, chart string) error {
	found := kubectl.CheckAppConfigMap(fmt.Sprintf("%s-%s", name, trainingType), namespace)
	if found {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", name, name)
	}

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

	// 5. Create Application
	err = kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
	}

	err = kubectl.InstallApps(template, namespace)
	if err != nil {
		// clean configmap
		log.Infof("clean up the config map %s because creating application failed.", name)
		log.Warnf("Please clean up the training job by using `arena delete %s --type %s`", name, trainingType)
		return err
	}

	// 6. Clean up the template file
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
