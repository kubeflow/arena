package workflow

import (
	"fmt"
	"os"

	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

/**
*	delete training job with the job name
**/

func DeleteJob(name, namespace, trainingType string, clientset *kubernetes.Clientset) error {
	jobName := fmt.Sprintf("%s-%s", name, trainingType)

	appInfoFileName, err := kubectl.SaveAppConfigMapToFile(jobName, "app", namespace)
	if err != nil {
		log.Debugf("Failed to SaveAppConfigMapToFile due to %v", err)
		return err
	}

	result, err := kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	if err != nil {
		log.Warnf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
		log.Warnln("manually delete the following resource:")
	}
	fmt.Printf("%s", result)

	_, err = clientset.CoreV1().ConfigMaps(namespace).Get(jobName, metav1.GetOptions{})

	if err != nil {
		log.Debugf("Skip deletion of ConfigMap %s, because the ConfigMap does not exists.", jobName)
		return nil
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

func GetDefaultValuesConfigmap(clientset *kubernetes.Clientset) (string, error) {
	configMap, err := clientset.CoreV1().ConfigMaps("runai").Get("cli-defaults", metav1.GetOptions{})
	if err != nil {
		return "", nil
	}

	values := configMap.Data["values"]
	valueFile, err := ioutil.TempFile(os.TempDir(), "values")
	if err != nil {
		return "", err
	}

	_, err = valueFile.WriteString(values)

	if err != nil {
		return "", err
	}

	log.Debugf("Wrote default cluster values file to path %s", valueFile.Name())

	return valueFile.Name(), nil
}

func UpdateConfigmapWithOwnerReference(name string, trainingType string, namespace string, clientset *kubernetes.Clientset) error {
	job, err := clientset.BatchV1().Jobs(namespace).Get(name, metav1.GetOptions{})

	// If job not found than do nothing
	if err != nil {
		return nil
	}

	log.Debugf("Found main job %s to use as owner reference", name)

	jobUID := job.UID

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(fmt.Sprintf("%s-%s", name, trainingType), metav1.GetOptions{})

	if err != nil {
		return err
	}

	log.Debugf("Found config map %s for updating", configMap.Name)

	configMap.OwnerReferences = append(configMap.OwnerReferences, metav1.OwnerReference{
		APIVersion: "batch/v1",
		Name:       name,
		UID:        jobUID,
		Kind:       "Job",
	})

	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(configMap)

	if err != nil {
		return err
	}

	log.Debugf("Updated owner reference of ConfigMap %s to Job %s", configMap.Name, name)

	return nil
}

func SubmitJob(name string, trainingType string, namespace string, values interface{}, chart string, clientset *kubernetes.Clientset) error {
	found := kubectl.CheckAppConfigMap(fmt.Sprintf("%s-%s", name, trainingType), namespace)
	if found {
		return fmt.Errorf("the job %s is already exist, please delete it first. use '%s delete %s'", name, config.CLIName, name)
	}

	// 1. Generate value file
	valueFileName, err := helm.GenerateValueFile(values)
	if err != nil {
		return err
	}

	defaultValuesFile, err := GetDefaultValuesConfigmap(clientset)

	if err != nil {
		log.Debugln(err)
		return fmt.Errorf("Error getting default values file of cluster")
	}

	// 2. Generate Template file
	template, err := helm.GenerateHelmTemplate(name, namespace, valueFileName, defaultValuesFile, chart)
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

	err = kubectl.CreateAppConfigmap(name,
		trainingType,
		namespace,
		valueFileName,
		defaultValuesFile,
		appInfoFileName,
		chartName,
		chartVersion)
	if err != nil {
		return err
	}
	err = kubectl.LabelAppConfigmap(name, trainingType, namespace, kubectl.JOB_CONFIG_LABEL)
	if err != nil {
		return err
	}

	// 5. Create Application
	_, err = kubectl.UninstallAppsWithAppInfoFile(appInfoFileName, namespace)
	if err != nil {
		log.Debugf("Failed to UninstallAppsWithAppInfoFile due to %v", err)
	}

	result, err := kubectl.InstallApps(template, namespace)
	fmt.Printf("%s", result)
	if err != nil {
		// clean configmap
		log.Infof("clean up the config map %s because creating application failed.", name)
		log.Warnf("Please clean up the training job by using `%s delete %s`", config.CLIName, name)
		return err
	}

	err = UpdateConfigmapWithOwnerReference(name, trainingType, namespace, clientset)
	if err != nil {
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
