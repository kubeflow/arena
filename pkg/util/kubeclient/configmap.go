package kubeclient

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var configmapTemplate = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: arena-configmap
  namespace: default
  labels:
    createdBy: arena
data: {}
`
var ErrConfigMapNotFound = errors.New("configmap of job is not found")

func UpdateConfigMapLabelsAndAnnotations(namespace string, name string, labels map[string]string, annotations map[string]string) error {
	arenaConfiger := config.GetArenaConfiger()
	client := arenaConfiger.GetClientSet()
	oldConfigMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	newConfigMap := oldConfigMap.DeepCopy()
	if len(newConfigMap.ObjectMeta.Annotations) == 0 {
		newConfigMap.ObjectMeta.Annotations = map[string]string{}
	}
	if len(newConfigMap.ObjectMeta.Labels) == 0 {
		newConfigMap.ObjectMeta.Labels = map[string]string{}
	}
	for k, v := range labels {
		newConfigMap.ObjectMeta.Labels[k] = v
	}
	for k, v := range annotations {
		newConfigMap.ObjectMeta.Annotations[k] = v
	}
	_, err = client.CoreV1().ConfigMaps(newConfigMap.ObjectMeta.Namespace).Update(context.TODO(), newConfigMap, metav1.UpdateOptions{})
	return err
}

func DeleteConfigMap(namespace string, name string) error {
	arenaConfiger := config.GetArenaConfiger()
	client := arenaConfiger.GetClientSet()
	return client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	arenaConfiger := config.GetArenaConfiger()
	client := arenaConfiger.GetClientSet()
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func CheckJobIsOwnedByUser(namespace, jobName string, jobType types.TrainingJobType) (bool, error) {
	configmap, err := GetConfigMap(namespace, fmt.Sprintf("%v-%v", jobName, jobType))
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, ErrConfigMapNotFound
		}
		return false, err
	}
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() && configmap.Labels[types.UserNameIdLabel] != arenaConfiger.GetUser().GetId() {
		return false, nil
	}
	return true, nil
}

func CreateAppConfigmap(name, trainingType, namespace, configFileName, appInfoFileName, chartName, chartVersion string) (err error) {
	data := map[string]string{
		chartName: chartVersion,
	}
	content, err := os.ReadFile(configFileName)
	if err != nil {
		return err
	}
	data["values"] = string(content)
	content, err = os.ReadFile(appInfoFileName)
	if err != nil {
		return err
	}
	data["app"] = string(content)

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(configmapTemplate), nil, nil)
	if err != nil {
		return err
	}
	configmap := obj.(*corev1.ConfigMap)
	configmap.Name = fmt.Sprintf("%v-%v", name, trainingType)
	configmap.Namespace = namespace
	configmap.Data = data
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() {
		configmap.Labels[types.UserNameIdLabel] = arenaConfiger.GetUser().GetId()
	}
	client := arenaConfiger.GetClientSet()
	_, err = client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configmap, metav1.CreateOptions{})
	return err
}
