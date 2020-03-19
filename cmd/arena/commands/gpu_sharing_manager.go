package commands

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	runaiFractionGPUSuffix = "runai-fraction-gpu"
	runaiGPUFraction       = "gpu-fraction"
	runaiGPUIndex          = "runai-gpu"
	runaiVisibleDevices    = "RUNAI-VISIBLE-DEVICES"
	factorForGPUFraction   = 0.7
)

func handleSharedGPUsIfNeeded(clientSet kubernetes.Interface, name string, submitArgs *submitRunaiJobArgs) error {
	if submitArgs.GPU == nil {
		return nil
	}

	if float64(int(*submitArgs.GPU)) == *submitArgs.GPU {
		gpu := int(*submitArgs.GPU)
		submitArgs.GPUInt = &gpu

		return nil
	}

	err := validateFractionalGPUTask(submitArgs)
	if err != nil {
		return err
	}

	submitArgs.GPUFraction = fmt.Sprintf("%v", *submitArgs.GPU)
	submitArgs.GPUFractionFixed = fmt.Sprintf("%v", (*submitArgs.GPU)*factorForGPUFraction)

	return setConfigMapForFractionGPUJobs(clientSet, name)
}

func validateFractionalGPUTask(submitArgs *submitRunaiJobArgs) error {
	if submitArgs.Interactive == nil || *submitArgs.Interactive == false {
		return fmt.Errorf("Jobs that require a fractional number of GPUs must be interactive. Run the job with flag '--interactive'")
	}

	if submitArgs.Elastic != nil && *submitArgs.Elastic == true {
		return fmt.Errorf("Jobs that require a fractional number of GPUs can't be elastic jobs. Run the job without flag '--elastic'")
	}

	if *submitArgs.GPU > 1 {
		return fmt.Errorf("Jobs that require a fractional number of GPUs must require less than 1 GPU")
	}

	return nil
}

func setConfigMapForFractionGPUJobs(clientSet kubernetes.Interface, jobName string) error {
	configMapName := fmt.Sprintf("%v-%v", jobName, runaiFractionGPUSuffix)
	configMap, err := clientSet.CoreV1().ConfigMaps(defaultNamespace).Get(configMapName, metav1.GetOptions{})

	// Map already exists
	if err == nil {
		configMap.Data[runaiVisibleDevices] = ""
		_, err = clientSet.CoreV1().ConfigMaps(defaultNamespace).Update(configMap)
		return err
	}

	data := make(map[string]string)
	data[runaiVisibleDevices] = ""
	configMap = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: data,
	}

	_, err = clientSet.CoreV1().ConfigMaps(defaultNamespace).Create(configMap)
	return err
}
