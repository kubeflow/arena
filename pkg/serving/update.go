// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serving

import (
	"fmt"
	"strconv"
	"strings"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	"github.com/kserve/kserve/pkg/constants"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util/kubectl"
)

const (
	ResourceGPU       corev1.ResourceName = "nvidia.com/gpu"
	ResourceGPUMemory corev1.ResourceName = "aliyun.com/gpu-mem"
	ResourceGPUCore   corev1.ResourceName = "aliyun.com/gpu-core.percentage"
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

	if len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.Annotations[k] = v
		}
	}

	if len(args.Labels) > 0 {
		for k, v := range args.Labels {
			deploy.Labels[k] = v
			deploy.Spec.Template.Labels[k] = v
		}
	}

	if len(args.NodeSelectors) > 0 {
		if deploy.Spec.Template.Spec.NodeSelector == nil {
			deploy.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			deploy.Spec.Template.Spec.NodeSelector[k] = v
		}
	}

	if len(args.Tolerations) > 0 {
		if deploy.Spec.Template.Spec.Tolerations == nil {
			deploy.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
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
				deploy.Spec.Template.Spec.Tolerations = append(deploy.Spec.Template.Spec.Tolerations, corev1.Toleration{
					Key:      toleration.Key,
					Value:    toleration.Value,
					Effect:   corev1.TaintEffect(toleration.Effect),
					Operator: corev1.TolerationOperator(toleration.Operator),
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

	if args.Command == "" && args.ModelRepository != "" {
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

	if len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.Annotations[k] = v
		}
	}

	if len(args.Labels) > 0 {
		for k, v := range args.Labels {
			deploy.Labels[k] = v
			deploy.Spec.Template.Labels[k] = v
		}
	}

	if len(args.NodeSelectors) > 0 {
		if deploy.Spec.Template.Spec.NodeSelector == nil {
			deploy.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			deploy.Spec.Template.Spec.NodeSelector[k] = v
		}
	}

	if len(args.Tolerations) > 0 {
		if deploy.Spec.Template.Spec.Tolerations == nil {
			deploy.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
		}
		for _, toleration := range args.Tolerations {
			deploy.Spec.Template.Spec.Tolerations = append(deploy.Spec.Template.Spec.Tolerations, corev1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   corev1.TaintEffect(toleration.Effect),
				Operator: corev1.TolerationOperator(toleration.Operator),
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

	if len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.Annotations[k] = v
		}
	}

	if len(args.Labels) > 0 {
		for k, v := range args.Labels {
			deploy.Labels[k] = v
			deploy.Spec.Template.Labels[k] = v
		}
	}

	if len(args.NodeSelectors) > 0 {
		if deploy.Spec.Template.Spec.NodeSelector == nil {
			deploy.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			deploy.Spec.Template.Spec.NodeSelector[k] = v
		}
	}

	if len(args.Tolerations) > 0 {
		if deploy.Spec.Template.Spec.Tolerations == nil {
			deploy.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
		}
		exist := map[string]bool{}
		var tolerations []corev1.Toleration
		for _, toleration := range args.Tolerations {
			tolerations = append(tolerations, corev1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   corev1.TaintEffect(toleration.Effect),
				Operator: corev1.TolerationOperator(toleration.Operator),
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

	if len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			inferenceService.Annotations[k] = v
		}
	}

	if len(args.Labels) > 0 {
		for k, v := range args.Labels {
			inferenceService.Labels[k] = v
		}
	}

	if len(args.NodeSelectors) > 0 {
		if inferenceService.Spec.Predictor.NodeSelector == nil {
			inferenceService.Spec.Predictor.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			inferenceService.Spec.Predictor.NodeSelector[k] = v
		}
	}

	if len(args.Tolerations) > 0 {
		if inferenceService.Spec.Predictor.Tolerations == nil {
			inferenceService.Spec.Predictor.Tolerations = []corev1.Toleration{}
		}
		exist := map[string]bool{}
		var tolerations []corev1.Toleration
		for _, toleration := range args.Tolerations {
			tolerations = append(tolerations, corev1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   corev1.TaintEffect(toleration.Effect),
				Operator: corev1.TolerationOperator(toleration.Operator),
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

func UpdateDistributedServing(args *types.UpdateDistributedServingArgs) error {
	lwsJob, err := findAndBuildLWSJob(args)
	if err != nil {
		return nil
	}

	if len(args.Annotations) > 0 {
		for k, v := range args.Annotations {
			lwsJob.Annotations[k] = v
			lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Annotations[k] = v
			lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Annotations[k] = v
		}
	}

	if len(args.Labels) > 0 {
		for k, v := range args.Labels {
			lwsJob.Labels[k] = v
			lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Labels[k] = v
			lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[k] = v
		}
	}

	if len(args.NodeSelectors) > 0 {
		if lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.NodeSelector == nil {
			lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.NodeSelector = map[string]string{}
		}
		if lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.NodeSelector == nil {
			lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.NodeSelector = map[string]string{}
		}
		for k, v := range args.NodeSelectors {
			lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.NodeSelector[k] = v
			lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.NodeSelector[k] = v
		}
	}

	if len(args.Tolerations) > 0 {
		if lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Tolerations == nil {
			lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Tolerations = []corev1.Toleration{}
		}
		if lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Tolerations == nil {
			lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Tolerations = []corev1.Toleration{}
		}
		exist := map[string]bool{}
		var tolerations []corev1.Toleration
		for _, toleration := range args.Tolerations {
			tolerations = append(tolerations, corev1.Toleration{
				Key:      toleration.Key,
				Value:    toleration.Value,
				Effect:   corev1.TaintEffect(toleration.Effect),
				Operator: corev1.TolerationOperator(toleration.Operator),
			})
			exist[toleration.Key+toleration.Value] = true
		}

		for _, preToleration := range lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Tolerations {
			if !exist[preToleration.Key+preToleration.Value] {
				tolerations = append(tolerations, preToleration)
			}
		}
		for _, preToleration := range lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Tolerations {
			if !exist[preToleration.Key+preToleration.Value] {
				tolerations = append(tolerations, preToleration)
			}
		}
		lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Tolerations = tolerations
		lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Tolerations = tolerations
	}

	return updateLWSJob(args.Name, args.Version, lwsJob)
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
	case types.TritonServingJob:
		suffix = "tritoninferenceserver"
	case types.CustomServingJob:
		suffix = "custom-serving"
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
		resourceLimits = make(map[corev1.ResourceName]resource.Quantity)
	}

	if args.GPUCount > 0 {
		resourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.GPUCount))
		delete(resourceLimits, ResourceGPUMemory)
	}

	if args.GPUMemory > 0 {
		resourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.GPUMemory))
		delete(resourceLimits, ResourceGPU)
	}

	if args.GPUCore > 0 && args.GPUCore%5 == 0 {
		resourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.GPUCore))
		delete(resourceLimits, ResourceGPU)
	}

	if args.Cpu != "" {
		resourceLimits[corev1.ResourceCPU] = resource.MustParse(args.Cpu)
	}

	if args.Memory != "" {
		resourceLimits[corev1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	deploy.Spec.Template.Spec.Containers[0].Resources.Limits = resourceLimits

	var newEnvs []corev1.EnvVar
	exist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := corev1.EnvVar{
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
		MinReplicas := int32(args.MinReplicas)
		inferenceService.Spec.Predictor.MinReplicas = &MinReplicas
	}
	if args.MaxReplicas > 0 {
		inferenceService.Spec.Predictor.MaxReplicas = int32(args.MaxReplicas)
	}
	if args.ScaleTarget > 0 {
		ScaleTarget := int32(args.ScaleTarget)
		inferenceService.Spec.Predictor.ScaleTarget = &ScaleTarget
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

func findAndBuildLWSJob(args *types.UpdateDistributedServingArgs) (*lwsv1.LeaderWorkerSet, error) {
	job, err := SearchServingJob(args.Namespace, args.Name, args.Version, args.Type)
	if err != nil {
		return nil, err
	}
	if args.Version == "" {
		jobInfo := job.Convert2JobInfo()
		args.Version = jobInfo.Version
	}

	lwsName := fmt.Sprintf("%s-%s-%s", args.Name, args.Version, "distributed-serving")
	lwsJob, err := kubectl.GetLWSJob(lwsName, args.Namespace)
	if err != nil {
		return nil, err
	}

	if args.Image != "" {
		lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Image = args.Image
		lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image = args.Image
	}

	if args.Replicas > 0 {
		replicas := int32(args.Replicas)
		lwsJob.Spec.Replicas = &replicas
	}

	// update resource
	masterResourceLimits := lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Limits
	workerResourceLimits := lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits
	if masterResourceLimits == nil {
		masterResourceLimits = make(map[corev1.ResourceName]resource.Quantity)
	}
	if workerResourceLimits == nil {
		workerResourceLimits = make(map[corev1.ResourceName]resource.Quantity)
	}
	if args.MasterCpu != "" {
		masterResourceLimits[corev1.ResourceCPU] = resource.MustParse(args.MasterCpu)
	}
	if args.WorkerCpu != "" {
		workerResourceLimits[corev1.ResourceCPU] = resource.MustParse(args.WorkerCpu)
	}
	if args.MasterGPUCount > 0 {
		masterResourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.MasterGPUCount))
		delete(masterResourceLimits, ResourceGPUMemory)
	}
	if args.WorkerGPUCount > 0 {
		workerResourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.WorkerGPUCount))
		delete(workerResourceLimits, ResourceGPUMemory)
	}
	if args.MasterGPUMemory > 0 {
		masterResourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.MasterGPUMemory))
		delete(masterResourceLimits, ResourceGPU)
	}
	if args.WorkerGPUMemory > 0 {
		workerResourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.WorkerGPUMemory))
		delete(workerResourceLimits, ResourceGPU)
	}
	if args.MasterGPUCore > 0 {
		masterResourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.MasterGPUCore))
		delete(masterResourceLimits, ResourceGPU)
	}
	if args.WorkerGPUCore > 0 {
		workerResourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.WorkerGPUCore))
		delete(workerResourceLimits, ResourceGPU)
	}
	if args.MasterMemory != "" {
		masterResourceLimits[corev1.ResourceMemory] = resource.MustParse(args.MasterMemory)
	}
	if args.WorkerMemory != "" {
		workerResourceLimits[corev1.ResourceMemory] = resource.MustParse(args.WorkerMemory)
	}
	lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Limits = masterResourceLimits
	lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits = workerResourceLimits

	// update env
	var masterEnvs []corev1.EnvVar
	var workerEnvs []corev1.EnvVar
	masterExist := map[string]bool{}
	workerExist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := corev1.EnvVar{
				Name:  k,
				Value: v,
			}
			masterEnvs = append(masterEnvs, envVar)
			workerEnvs = append(workerEnvs, envVar)
			masterExist[k] = true
			workerExist[k] = true
		}
	}
	for _, env := range lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Env {
		if !masterExist[env.Name] {
			masterEnvs = append(masterEnvs, env)
		}
	}
	for _, env := range lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Env {
		if !workerExist[env.Name] {
			workerEnvs = append(workerEnvs, env)
		}
	}
	lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Env = masterEnvs
	lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Env = workerEnvs

	// update command
	if args.MasterCommand != "" {
		// commands: sh -c xxx
		masterCommand := lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Command
		newMasterCommand := make([]string, len(masterCommand))
		copy(newMasterCommand, masterCommand)
		newMasterCommand[len(newMasterCommand)-1] = args.MasterCommand
		lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Command = newMasterCommand
		lwsJob.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Args = []string{}
	}
	if args.WorkerCommand != "" {
		// commands: sh -c xxx
		workerCommand := lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Command
		newWorkerCommand := make([]string, len(workerCommand))
		copy(newWorkerCommand, workerCommand)
		newWorkerCommand[len(newWorkerCommand)-1] = args.WorkerCommand
		lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Command = newWorkerCommand
		lwsJob.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Args = []string{}
	}

	return lwsJob, nil
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

func updateLWSJob(name, version string, lwsJob *lwsv1.LeaderWorkerSet) error {
	err := kubectl.UpdateLWSJob(lwsJob)
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
		resourceLimits = make(map[corev1.ResourceName]resource.Quantity)
	}
	if args.GPUCount > 0 {
		resourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.GPUCount))
		delete(resourceLimits, ResourceGPUMemory)
	}
	if args.GPUMemory > 0 {
		resourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.GPUMemory))
		delete(resourceLimits, ResourceGPU)
	}
	if args.GPUCore > 0 && args.GPUCore%5 == 0 {
		resourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.GPUCore))
		delete(resourceLimits, ResourceGPU)
	}
	if args.Cpu != "" {
		resourceLimits[corev1.ResourceCPU] = resource.MustParse(args.Cpu)
	}
	if args.Memory != "" {
		resourceLimits[corev1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	inferenceService.Spec.Predictor.Model.Resources.Limits = resourceLimits

	// set env
	var newEnvs []corev1.EnvVar
	exist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := corev1.EnvVar{
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

	//set volume
	if len(args.ModelDirs) != 0 {
		log.Debugf("update modelDirs: [%+v]", args.ModelDirs)
		var volumes []corev1.Volume
		var volumeMounts []corev1.VolumeMount

		for pvName, mountPath := range args.ModelDirs {
			volumes = append(volumes, corev1.Volume{
				Name: pvName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvName,
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      pvName,
				MountPath: mountPath,
			})
		}
		inferenceService.Spec.Predictor.Containers[0].VolumeMounts = volumeMounts
		inferenceService.Spec.Predictor.Volumes = volumes
	}

	// set resources requests
	resourceRequests := inferenceService.Spec.Predictor.Containers[0].Resources.Requests
	if resourceRequests == nil {
		resourceRequests = make(map[corev1.ResourceName]resource.Quantity)
	}
	if args.Cpu != "" {
		resourceRequests[corev1.ResourceCPU] = resource.MustParse(args.Cpu)
	}
	if args.Memory != "" {
		resourceRequests[corev1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	inferenceService.Spec.Predictor.Containers[0].Resources.Requests = resourceRequests

	// set resources limits
	resourceLimits := inferenceService.Spec.Predictor.Containers[0].Resources.Limits
	if resourceLimits == nil {
		resourceLimits = make(map[corev1.ResourceName]resource.Quantity)
	}
	if args.GPUCount > 0 {
		resourceLimits[ResourceGPU] = resource.MustParse(strconv.Itoa(args.GPUCount))
		delete(resourceLimits, ResourceGPUMemory)
	}
	if args.GPUMemory > 0 {
		resourceLimits[ResourceGPUMemory] = resource.MustParse(strconv.Itoa(args.GPUMemory))
		delete(resourceLimits, ResourceGPU)
	}
	if args.GPUCore > 0 && args.GPUCore%5 == 0 {
		resourceLimits[ResourceGPUCore] = resource.MustParse(strconv.Itoa(args.GPUCore))
		delete(resourceLimits, ResourceGPU)
	}
	if args.Cpu != "" {
		resourceLimits[corev1.ResourceCPU] = resource.MustParse(args.Cpu)
	}
	if args.Memory != "" {
		resourceLimits[corev1.ResourceMemory] = resource.MustParse(args.Memory)
	}
	inferenceService.Spec.Predictor.Containers[0].Resources.Limits = resourceLimits

	// set env
	var newEnvs []corev1.EnvVar
	exist := map[string]bool{}
	if args.Envs != nil {
		for k, v := range args.Envs {
			envVar := corev1.EnvVar{
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
