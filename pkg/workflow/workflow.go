package workflow

import (
	"fmt"
	"os"

	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubeclient"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kubeflow/arena/pkg/apis/types"
)

/**
*	delete training job with the job name
**/

func DeleteJob(name, namespace, trainingType string) error {
	jobName := fmt.Sprintf("%s-%s", name, trainingType)

	appInfoFileName, err := kubectl.SaveAppConfigMapToFile(jobName, "app", namespace)
	if err != nil {
		log.Debugf("Failed to SaveAppConfigMapToFile due to %v", err)
		return err
	}

	err = kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	if err != nil {
		log.Warnf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
		log.Warnln("manually delete the following resource:")
	}

	err = kubectl.DeleteAppConfigMap(jobName, namespace)
	if err != nil {
		log.Warningf("Delete configmap %s failed, please clean it manually due to %v.", jobName, err)
		log.Warningf("Please run `kubectl delete -n %s cm %s`", namespace, jobName)
	}

	return nil
}

/**
*	Submit operation, scaleIn or scaleOut
**/

func SubmitOps(name string, trainingType string, namespace string, values interface{}, chart string, options ...string) error {
	found := kubectl.CheckAppConfigMap(fmt.Sprintf("%s-%s", name, trainingType), namespace)
	if found {
		return fmt.Errorf("the job configmap %v-%v is already exist, please delete it first.", name, trainingType)
	}

	// 1. Generate value file
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return err
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, chart, options...)
	if err != nil {
		return err
	}

	// 3. Generate AppInfo file
	appInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
		return err
	}

	// 4. Create Application
	err = kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
	}

	result, err := kubectl.InstallApps(template, namespace)
	fmt.Printf("%s", result)
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

		err = os.Remove(appInfoFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", appInfoFileName, err)
		}
	}

	return nil
}

/**
*	Submit training job
**/

func SubmitJob(name string, trainingType string, namespace string, values interface{}, chart string, options ...string) error {
	_, err := kubeclient.GetConfigMap(namespace, fmt.Sprintf("%v-%v", name, trainingType))
	if err == nil {
		return fmt.Errorf("the job configmap %v-%v is already exist, please delete it first.", name, trainingType)
	}
	if !k8serrors.IsNotFound(err) {
		return err
	}
	// 1. Generate value file
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return err
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, chart, options...)
	if err != nil {
		return err
	}

	// 3. Generate AppInfo file
	appInfoFileName, err := kubectl.SaveAppInfo(template, namespace)
	if err != nil {
		return err
	}

	// 4. Keep value file in configmap
	chartName := helm.GetChartName(chart)
	chartVersion, err := helm.GetChartVersion(chart)
	if err != nil {
		return err
	}
	err = kubeclient.CreateAppConfigmap(name,
		trainingType,
		namespace,
		valueFileName,
		appInfoFileName,
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

	result, err := kubectl.InstallApps(template, namespace)
	fmt.Printf("%s", result)
	if err != nil {
		// clean configmap
		log.Infof("clean up the config map %s because creating application failed.", name)
		log.Warnf("Please clean up the training job by using `arena delete %s --type %s`", name, trainingType)
		return err
	}

	// 6. Patch OwnerReference for tfjob / pytorchjob
	if trainingType == string(types.TFTrainingJob) ||
		trainingType == string(types.PytorchTrainingJob) {
		err := kubectl.PatchOwnerReferenceWithAppInfoFile(name, trainingType, appInfoFileName, namespace)
		if err != nil {
			log.Warnf("Failed to patch ownerReference %s due to %v`", name, err)
			return err
		}
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

		err = os.Remove(appInfoFileName)
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", appInfoFileName, err)
		}
	}

	return nil
}
