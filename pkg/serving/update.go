package serving

import (
	"fmt"
	"strconv"
	"strings"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/kserve/kserve/pkg/constants"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util/kubectl"
)

const (
	ResourceGPU       v1.ResourceName = "nvidia.com/gpu"
	ResourceGPUMemory v1.ResourceName = "aliyun.com/gpu-mem"
	ResourceGPUCore   v1.ResourceName = "aliyun.com/gpu-core.percentage"
)

func UpdateTensorflowServing(args *types.UpdateTensorFlowServingArgs) error {
	deploy, err := findAndBuildDeployment(&args.CommonUpdateServingArgs)
	if err != nil {
		return err
	}

	if args.Command == "" {
		containerArgs := deploy.Spec.Template.Spec.Containers[0].Args
		if len(containerArgs) > 0 {
			servingArgs := containerArgs[0]

			if strings.HasSuffix(servingArgs, "\n") {
				servingArgs = strings.TrimSpace(servingArgs[:len(servingArgs)-1])
			}
			arr := strings.Split(servingArgs, "--")
			params := make(map[string]string)
			for index, argItem := range arr {
				if index == 0 {
					continue
				}
				pair := strings.Split(argItem, "=")
				if len(pair) <= 1 {
					continue
				}
				params[fmt.Sprintf("--%s", pair[0])] = argItem[len(pair[0])+1:]
			}
			if args.ModelName != "" {
				params["--model_name"] = args.ModelName
			}
			if args.ModelPath != "" {
				params["--model_base_path"] = args.ModelPath
			}
			if args.ModelConfigFile != "" {
				params["--model_config_file"] = args.ModelConfigFile
			}
			if args.MonitoringConfigFile != "" {
				params["--monitoring_config_file"] = args.MonitoringConfigFile
			}

			var newArgs []string
			newArgs = append(newArgs, "/usr/bin/tensorflow_model_server")
			for k, v := range params {
				newArgs = append(newArgs, fmt.Sprintf("%s=%s", k, v))
			}

			deploy.Spec.Template.Spec.Containers[0].Args = []string{strings.Join(newArgs, " ")}
		}
	}

	if args.Annotations != nil && len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.Annotations[k] = v
		}
	}

	if args.Labels != nil && len(args.Labels) > 0 {
		for k, v := range args.Labels {
			deploy.Labels[k] = v
			deploy.Spec.Template.Labels[k] = v
		}
	}

	if args.NodeSelectors != nil && len(args.NodeSelectors) > 0 {
		if deploy.Spec.Template.Spec.NodeSelector == nil {
			deploy.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			deploy.Spec.Template.Spec.NodeSelector[k] = v
		}
	}

	if args.Tolerations != nil && len(args.Tolerations) > 0 {
		if deploy.Spec.Template.Spec.Tolerations == nil {
			deploy.Spec.Template.Spec.Tolerations = []v1.Toleration{}
		}
		mapSet := make(map[string]interface{})
		for _, toleration := range deploy.Spec.Template.Spec.Tolerations {
			mapSet[fmt.Sprintf("%s=%s:%s,%s", toleration.Key,
				toleration.Value,
				toleration.Effect,
				toleration.Operator)] = nil
		}
		for _, toleration := range args.Tolerations {
			if _, ok := mapSet[fmt.Sprintf("%s=%s:%s,%s", toleration.Key,
				toleration.Value,
				toleration.Effect,
				toleration.Operator)]; !ok {
				deploy.Spec.Template.Spec.Tolerations = append(deploy.Spec.Template.Spec.Tolerations, v1.Toleration{
					Key:      toleration.Key,
					Value:    toleration.Value,
					Effect:   v1.TaintEffect(toleration.Effect),
					Operator: v1.TolerationOperator(toleration.Operator),
				})
			}

		}
	}

	return updateDeployment(args.Name, args.Version, deploy)
}

func UpdateTritonServing(args *types.UpdateTritonServingArgs) error {
	deploy, err := findAndBuildDeployment(&args.CommonUpdateServingArgs)
	if err != nil {
		return err
	}

	if args.Command == "" {
		containerArgs := deploy.Spec.Template.Spec.Containers[0].Args

		servingArgs := containerArgs[0]
		if strings.HasSuffix(servingArgs, "\n") {
			servingArgs = strings.TrimSpace(servingArgs[:len(servingArgs)-1])
		}
		arr := strings.Split(servingArgs, "--")

		params := make(map[string]string)
		for index, argItem := range arr {
			if index == 0 {
				continue
			}
			pair := strings.Split(argItem, "=")
			if len(pair) <= 1 {
				continue
			}
			params[fmt.Sprintf("--%s", pair[0])] = argItem[len(pair[0])+1:]
		}

		if args.ModelRepository != "" {
			params["--model-repository"] = args.ModelRepository
		}

		var newArgs []string
		newArgs = append(newArgs, "tritonserver")
		for k, v := range params {
			newArgs = append(newArgs, fmt.Sprintf("%s=%s", k, v))
		}

		deploy.Spec.Template.Spec.Containers[0].Args = []string{strings.Join(newArgs, " ")}
	}

	if args.Annotations != nil && len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.Annotations[k] = v
		}
	}

	if args.Labels != nil && len(args.Labels) > 0 {
		for k, v := range args.Labels {
			deploy.Labels[k] = v
			deploy.Spec.Template.Labels[k] = v
		}
	}

	if args.NodeSelectors != nil && len(args.NodeSelectors) > 0 {
		if deploy.Spec.Template.Spec.NodeSelector == nil {
			deploy.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			deploy.Spec.Template.Spec.NodeSelector[k] = v
		}
	}

	if args.Tolerations != nil && len(args.Tolerations) > 0 {
		if deploy.Spec.Template.Spec.Tolerations == nil {
			deploy.Spec.Template.Spec.Tolerations = []v1.Toleration{}
		}
		for _, toleration := range args.Tolerations {
			deploy.Spec.Template.Spec.Tolerations = append(deploy.Spec.Template.Spec.Tolerations, v1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   v1.TaintEffect(toleration.Effect),
				Operator: v1.TolerationOperator(toleration.Operator),
			})
		}
	}

	return updateDeployment(args.Name, args.Version, deploy)
}

func UpdateCustomServing(args *types.UpdateCustomServingArgs) error {
	deploy, err := findAndBuildDeployment(&args.CommonUpdateServingArgs)
	if err != nil {
		return err
	}

	if args.Annotations != nil && len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.Annotations[k] = v
		}
	}

	if args.Labels != nil && len(args.Labels) > 0 {
		for k, v := range args.Labels {
			deploy.Labels[k] = v
			deploy.Spec.Template.Labels[k] = v
		}
	}

	if args.NodeSelectors != nil && len(args.NodeSelectors) > 0 {
		if deploy.Spec.Template.Spec.NodeSelector == nil {
			deploy.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			deploy.Spec.Template.Spec.NodeSelector[k] = v
		}
	}

	if args.Tolerations != nil && len(args.Tolerations) > 0 {
		if deploy.Spec.Template.Spec.Tolerations == nil {
			deploy.Spec.Template.Spec.Tolerations = []v1.Toleration{}
		}
		exist := map[string]bool{}
		var tolerations []v1.Toleration
		for _, toleration := range args.Tolerations {
			tolerations = append(tolerations, v1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   v1.TaintEffect(toleration.Effect),
				Operator: v1.TolerationOperator(toleration.Operator),
			})
			exist[toleration.Key+toleration.Value] = true
		}

		for _, preToleration := range deploy.Spec.Template.Spec.Tolerations {
			if !exist[preToleration.Key+preToleration.Value] {
				tolerations = append(tolerations, preToleration)
			}
		}
		deploy.Spec.Template.Spec.Tolerations = tolerations
	}

	return updateDeployment(args.Name, args.Version, deploy)
}

func UpdateKServe(args *types.UpdateKServeArgs) error {
	inferenceService, err := findAndBuildInferenceService(args)
	if err != nil {
		return err
	}

	if args.Annotations != nil && len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			inferenceService.Annotations[k] = v
		}
	}

	if args.Labels != nil && len(args.Labels) > 0 {
		for k, v := range args.Labels {
			inferenceService.Labels[k] = v
		}
	}

	if args.NodeSelectors != nil && len(args.NodeSelectors) > 0 {
		if inferenceService.Spec.Predictor.NodeSelector == nil {
			inferenceService.Spec.Predictor.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			inferenceService.Spec.Predictor.NodeSelector[k] = v
		}
	}

	if args.Tolerations != nil && len(args.Tolerations) > 0 {
		if inferenceService.Spec.Predictor.Tolerations == nil {
			inferenceService.Spec.Predictor.Tolerations = []v1.Toleration{}
		}
		exist := map[string]bool{}
		var tolerations []v1.Toleration
		for _, toleration := range args.Tolerations {
			tolerations = append(tolerations, v1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   v1.TaintEffect(toleration.Effect),
				Operator: v1.TolerationOperator(toleration.Operator),
			})
			exist[toleration.Key+toleration.Value] = true
		}

		for _, preToleration := range inferenceService.Spec.Predictor.Tolerations {
			if !exist[preToleration.Key+preToleration.Value] {
				tolerations = append(tolerations, preToleration)
			}
		}
		inferenceService.Spec.Predictor.Tolerations = tolerations
	}

	return updateInferenceService(args.Name, args.Version, inferenceService)
}

func findAndBuildDeployment(args *types.CommonUpdateServingArgs) (*appsv1.Deployment, error) {
	job, err := SearchServingJob(args.Namespace, args.Name, args.Version, args.Type)
	if err != nil {
		return nil, err
	}
	if args.Version == "" {
		jobInfo := job.Convert2JobInfo()
		args.Version = jobInfo.Version
	}

	var suffix string
	switch args.Type {
	case types.TFServingJob:
		suffix = "tensorflow-serving"
		break
	case types.TritonServingJob:
		suffix = "tritoninferenceserver"
		break
	case types.CustomServingJob:
		suffix = "custom-serving"
		break
	default:
		return nil, fmt.Errorf("invalid serving job type [%s]", args.Type)
	}

	deployName := fmt.Sprintf("%s-%s-%s", args.Name, args.Version, suffix)
	deploy, err := kubectl.GetDeployment(deployName, args.Namespace)
	if err != nil {
		return nil, err
	}

	if args.Image != "" {
		deploy.Spec.Template.Spec.Containers[0].Image = args.Image
	}

	if args.Replicas >= 0 {
		replicas := int32(args.Replicas)
		deploy.Spec.Replicas = &replicas
	}

	resourceLimits := deploy.Spec.Template.Spec.Containers[0].Resources.Limits
	if resourceLimits == nil {
		resourceLimits = make(map[v1.ResourceName]resource.Quantity)
	}

	if args.GPUCount > 0 {
		resourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.GPUCount))
		if _, ok := resourceLimits[ResourceGPUMemory]; ok {
			delete(resourceLimits, ResourceGPUMemory)
		}
	}

	if args.GPUMemory > 0 {
		resourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.GPUMemory))
		if _, ok := resourceLimits[ResourceGPU]; ok {
			delete(resourceLimits, ResourceGPU)
		}
	}

	if args.GPUCore > 0 && args.GPUCore%5 == 0 {
		resourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.GPUCore))
		if _, ok := resourceLimits[ResourceGPU]; ok {
			delete(resourceLimits, ResourceGPU)
		}
	}

	if args.Cpu != "" {
		resourceLimits[v1.ResourceCPU] = resource.MustParse(args.Cpu)
	}

	if args.Memory != "" {
		resourceLimits[v1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	deploy.Spec.Template.Spec.Containers[0].Resources.Limits = resourceLimits

	var newEnvs []v1.EnvVar
	exist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := v1.EnvVar{
				Name:  k,
				Value: v,
			}
			newEnvs = append(newEnvs, envVar)
			exist[k] = true
		}
	}
	for _, env := range deploy.Spec.Template.Spec.Containers[0].Env {
		if !exist[env.Name] {
			newEnvs = append(newEnvs, env)
		}
	}
	deploy.Spec.Template.Spec.Containers[0].Env = newEnvs

	if args.Command != "" {
		// commands: sh -c xxx
		commands := deploy.Spec.Template.Spec.Containers[0].Command
		shell := commands[0]
		newCommands := []string{shell, "-c", args.Command}
		deploy.Spec.Template.Spec.Containers[0].Command = newCommands
		deploy.Spec.Template.Spec.Containers[0].Args = []string{}
	}

	return deploy, nil
}

func findAndBuildInferenceService(args *types.UpdateKServeArgs) (*kservev1beta1.InferenceService, error) {
	_, err := SearchServingJob(args.Namespace, args.Name, args.Version, args.Type)
	if err != nil {
		return nil, err
	}

	inferenceName := args.Name
	inferenceService, err := kubectl.GetInferenceService(inferenceName, args.Namespace)
	if err != nil {
		return nil, err
	}

	// check inference model type: modelFormat or custom
	if inferenceService.Spec.Predictor.Model != nil {
		setInferenceServiceForFrameworkModel(args, inferenceService)
	} else {
		setInferenceServiceForCustomModel(args, inferenceService)
	}

	if args.MinReplicas >= 0 {
		inferenceService.Spec.Predictor.MinReplicas = &args.MinReplicas
	}
	if args.MaxReplicas > 0 {
		inferenceService.Spec.Predictor.MaxReplicas = args.MaxReplicas
	}
	if args.ScaleTarget > 0 {
		inferenceService.Spec.Predictor.ScaleTarget = &args.ScaleTarget
	}
	if args.ScaleMetric != "" {
		scaleMetric := kservev1beta1.ScaleMetric(args.ScaleMetric)
		inferenceService.Spec.Predictor.ScaleMetric = &scaleMetric
	}
	if args.ContainerConcurrency > 0 {
		inferenceService.Spec.Predictor.ContainerConcurrency = &args.ContainerConcurrency
	}
	if args.TimeoutSeconds > 0 {
		inferenceService.Spec.Predictor.TimeoutSeconds = &args.TimeoutSeconds
	}
	if args.CanaryTrafficPercent >= 0 {
		inferenceService.Spec.Predictor.CanaryTrafficPercent = &args.CanaryTrafficPercent
	}

	return inferenceService, nil
}

func updateDeployment(name, version string, deploy *appsv1.Deployment) error {
	err := kubectl.UpdateDeployment(deploy)
	if err == nil {
		log.Infof("The serving job %s with version %s has been updated successfully", name, version)
	} else {
		log.Errorf("The serving job %s with version %s update failed", name, version)
	}
	return err
}

func updateInferenceService(name, version string, inferenceService *kservev1beta1.InferenceService) error {
	err := kubectl.UpdateInferenceService(inferenceService)
	if err != nil {
		log.Errorf("The serving job %s with version %s update failed", name, version)
		return err
	}

	log.Infof("The serving job %s with version %s has been updated successfully", name, version)
	return nil
}

func setInferenceServiceForFrameworkModel(args *types.UpdateKServeArgs, inferenceService *kservev1beta1.InferenceService) {
	if args.ModelFormat != nil {
		inferenceService.Spec.Predictor.Model.ModelFormat.Name = args.ModelFormat.Name
		inferenceService.Spec.Predictor.Model.ModelFormat.Version = args.ModelFormat.Version
	}
	if args.Runtime != "" {
		inferenceService.Spec.Predictor.Model.Runtime = &args.Runtime
	}
	if args.StorageUri != "" {
		inferenceService.Spec.Predictor.Model.StorageURI = &args.StorageUri
	}
	if args.RuntimeVersion != "" {
		inferenceService.Spec.Predictor.Model.RuntimeVersion = &args.RuntimeVersion
	}
	if args.ProtocolVersion != "" {
		protocolVersion := constants.InferenceServiceProtocol(args.ProtocolVersion)
		inferenceService.Spec.Predictor.Model.ProtocolVersion = &protocolVersion
	}
	if args.Image != "" {
		inferenceService.Spec.Predictor.Model.Image = args.Image
	}

	// set resources limits
	resourceLimits := inferenceService.Spec.Predictor.Model.Resources.Limits
	if resourceLimits == nil {
		resourceLimits = make(map[v1.ResourceName]resource.Quantity)
	}
	if args.GPUCount > 0 {
		resourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.GPUCount))
		if _, ok := resourceLimits[ResourceGPUMemory]; ok {
			delete(resourceLimits, ResourceGPUMemory)
		}
	}
	if args.GPUMemory > 0 {
		resourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.GPUMemory))
		if _, ok := resourceLimits[ResourceGPU]; ok {
			delete(resourceLimits, ResourceGPU)
		}
	}
	if args.GPUCore > 0 && args.GPUCore%5 == 0 {
		resourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.GPUCore))
		if _, ok := resourceLimits[ResourceGPU]; ok {
			delete(resourceLimits, ResourceGPU)
		}
	}
	if args.Cpu != "" {
		resourceLimits[v1.ResourceCPU] = resource.MustParse(args.Cpu)
	}
	if args.Memory != "" {
		resourceLimits[v1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	inferenceService.Spec.Predictor.Model.Resources.Limits = resourceLimits

	// set env
	var newEnvs []v1.EnvVar
	exist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := v1.EnvVar{
				Name:  k,
				Value: v,
			}
			newEnvs = append(newEnvs, envVar)
			exist[k] = true
		}
	}
	for _, env := range inferenceService.Spec.Predictor.Model.Env {
		if !exist[env.Name] {
			newEnvs = append(newEnvs, env)
		}
	}
	inferenceService.Spec.Predictor.Model.Env = newEnvs

	// set command
	if args.Command != "" {
		// commands: sh -c xxx
		commands := inferenceService.Spec.Predictor.Model.Command
		shell := "sh"
		if len(commands) > 0 {
			shell = commands[0]
		}
		newCommands := []string{shell, "-c", args.Command}
		inferenceService.Spec.Predictor.Model.Command = newCommands
		inferenceService.Spec.Predictor.Model.Args = []string{}
	}
}

func setInferenceServiceForCustomModel(args *types.UpdateKServeArgs, inferenceService *kservev1beta1.InferenceService) {
	if args.Image != "" {
		inferenceService.Spec.Predictor.Containers[0].Image = args.Image
	}

	// set resources limits
	resourceLimits := inferenceService.Spec.Predictor.Containers[0].Resources.Limits
	if resourceLimits == nil {
		resourceLimits = make(map[v1.ResourceName]resource.Quantity)
	}
	if args.GPUCount > 0 {
		resourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.GPUCount))
		if _, ok := resourceLimits[ResourceGPUMemory]; ok {
			delete(resourceLimits, ResourceGPUMemory)
		}
	}
	if args.GPUMemory > 0 {
		resourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.GPUMemory))
		if _, ok := resourceLimits[ResourceGPU]; ok {
			delete(resourceLimits, ResourceGPU)
		}
	}
	if args.GPUCore > 0 && args.GPUCore%5 == 0 {
		resourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.GPUCore))
		if _, ok := resourceLimits[ResourceGPU]; ok {
			delete(resourceLimits, ResourceGPU)
		}
	}
	if args.Cpu != "" {
		resourceLimits[v1.ResourceCPU] = resource.MustParse(args.Cpu)
	}
	if args.Memory != "" {
		resourceLimits[v1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	inferenceService.Spec.Predictor.Containers[0].Resources.Limits = resourceLimits

	// set env
	var newEnvs []v1.EnvVar
	exist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := v1.EnvVar{
				Name:  k,
				Value: v,
			}
			newEnvs = append(newEnvs, envVar)
			exist[k] = true
		}
	}
	for _, env := range inferenceService.Spec.Predictor.Containers[0].Env {
		if !exist[env.Name] {
			newEnvs = append(newEnvs, env)
		}
	}
	inferenceService.Spec.Predictor.Containers[0].Env = newEnvs

	// set command
	if args.Command != "" {
		// commands: sh -c xxx
		commands := inferenceService.Spec.Predictor.Containers[0].Command
		shell := "sh"
		if len(commands) > 0 {
			shell = commands[0]
		}
		newCommands := []string{shell, "-c", args.Command}
		inferenceService.Spec.Predictor.Containers[0].Command = newCommands
		inferenceService.Spec.Predictor.Containers[0].Args = []string{}
	}
}
