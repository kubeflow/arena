// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License
package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubeflow/arena/pkg/apis/types"
)

type SubmitRayJobArgsBuilder struct {
	args        *types.SubmitRayJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitRayJobArgsBuilder(args *types.SubmitRayJobArgs) ArgsBuilder {
	args.TrainingType = types.RayJob
	s := &SubmitRayJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewSubmitArgsBuilder(&s.args.CommonSubmitArgs),
		NewSubmitSyncCodeArgsBuilder(&s.args.SubmitSyncCodeArgs),
		NewSubmitTensorboardArgsBuilder(&s.args.SubmitTensorboardArgs),
	)
	return s
}

func (s *SubmitRayJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitRayJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitRayJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitRayJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}

	var (
		activeDeadline   time.Duration
		ttlAfterFinished = time.Second * 10
		preStopCmd       string
	)

	command.Flags().BoolVar(&s.args.ShutdownAfterJobFinishes, "shutdown-after-finished", true, "determine whether to delete the ray cluster once rayJob succeed or failed.")
	command.Flags().DurationVar(&ttlAfterFinished, "ttl-after-finished", ttlAfterFinished, "The TTL to clean up RayCluster. It's only working when ShutdownAfterJobFinishes set to true, like 10s.")
	command.Flags().DurationVar(&activeDeadline, "active-deadline-seconds", activeDeadline, "The duration in seconds that the RayJob may be active before KubeRay actively tries to terminate the RayJob, like 2m.")
	command.Flags().BoolVar(&s.args.Suspend, "suspend", false, "suspend specifies whether the RayJob controller should create a RayCluster instance.")
	command.Flags().StringVar(&s.args.ShareMemory, "share-memory", "", "the shared memory of each replica to run the job.")
	command.Flags().StringVar(&s.args.RayVersion, "ray-version", "", "the version of Ray you are using. Make sure all Ray containers are running this version of Ray.")
	command.Flags().StringVar(&preStopCmd, "pre-stop-command", "ray stop", "the command that needs to be executed before stopping.")
	command.Flags().BoolVar(&s.args.EnableInTreeAutoscaling, "enable-autoscaling", false, "enable-autoscaling indicates whether operator should create in tree autoscaling configs.")
	command.Flags().StringVar(&s.args.AutoscalerOptions.Cpu, "autoscaler-cpu", "500m", "autoscaler-cpu specifies optional resource request and limit overrides for the autoscaler container.")
	command.Flags().StringVar(&s.args.AutoscalerOptions.Memory, "autoscaler-memory", "512Mi", "autoscaler-memory specifies optional resource request and limit overrides for the autoscaler container.")
	command.Flags().StringVar(&s.args.AutoscalerOptions.Image, "autoscaler-image", "", "autoscaler-image optionally overrides the autoscaler's container image. This override is for provided for autoscaler testing and development.")
	command.Flags().StringVar(&s.args.AutoscalerOptions.ImagePullPolicy, "autoscaler-image-pull-policy", "IfNotPresent", "optionally overrides the autoscaler container's image pull policy.")
	command.Flags().Int32Var(&s.args.AutoscalerOptions.IdleTimeoutSeconds, "autoscaler-idle-timeout-seconds", 60, "is the number of seconds to wait before scaling down a worker pod which is not using Ray resources.")
	command.Flags().StringVar(&s.args.AutoscalerOptions.UpscalingMode, "autoscaler-upscaling-mode", "Default", "autoscaler-upscaling-mode is Conservative, Default, or Aggressive.")
	command.Flags().StringVar(&s.args.HeadGroupSpec.Image, "head-image", "", "the image for head pod.")
	command.Flags().StringVar(&s.args.HeadGroupSpec.Cpu, "head-cpu", "", "the cpu resource to HeadPod for the training, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.HeadGroupSpec.Memory, "head-memory", "", "the memory resource to HeadPod for the training, like 2Gi.")
	command.Flags().IntVar(&s.args.HeadGroupSpec.Gpu, "head-gpu", 0, "the GPU count of HeadPod to run the training.")
	command.Flags().StringVar(&s.args.HeadGroupSpec.ServiceType, "head-service-type", "ClusterIP", "is Kubernetes service type of the head service. it will be used by the workers to connect to the head pod.")
	command.Flags().StringVar(&s.args.WorkerGroupSpec.Image, "worker-image", "", "the image for worker pod.")
	command.Flags().StringVar(&s.args.WorkerGroupSpec.Cpu, "worker-cpu", "", "the cpu resource to WorkerPod for the training, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.WorkerGroupSpec.Memory, "worker-memory", "", "the memory resource to WorkerPod for the training, like 1Gi.")
	command.Flags().IntVar(&s.args.WorkerGroupSpec.Gpu, "worker-gpu", 0, "the GPU count of each worker to run the training.")
	command.Flags().Int32Var(&s.args.WorkerGroupSpec.Replicas, "worker-replicas", 0, "worker-replicas denotes the number of desired Pods for this worker group.")
	_ = command.MarkFlagRequired("worker-replicas")
	command.Flags().Int32Var(&s.args.WorkerGroupSpec.MinReplicas, "worker-min-replicas", 0, "worker-min-replicas denotes the minimum number of desired Pods for this worker group, default --worke-replicas.")
	command.Flags().Int32Var(&s.args.WorkerGroupSpec.MaxReplicas, "worker-max-replicas", 0, "worker-max-replicas denotes the maximum number of desired Pods for this worker group, default --worke-replicas.")
	command.Flags().Int32Var(&s.args.WorkerGroupSpec.NumOfHosts, "worker-num-of-hosts", 1, "worker-num-of-hosts denotes the number of hosts to create per replica.")

	s.AddArgValue("active-deadline-seconds", &activeDeadline).
		AddArgValue("ttl-after-finished", &ttlAfterFinished).
		AddArgValue("pre-stop-command", &preStopCmd)
}

func (s *SubmitRayJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	s.AddArgValue(ShareDataPrefix+"dataset", s.args.DataSet)
	return nil
}

func (s *SubmitRayJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.setRunPolicy(); err != nil {
		return err
	}
	if err := s.setPreStopCmd(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.addHeadAddr(); err != nil {
		return err
	}
	if err := s.addRequestGPUsToAnnotation(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitRayJobArgsBuilder) setRunPolicy() error {
	// Get active deadline
	if ad, ok := s.argValues["active-deadline-seconds"]; ok {
		activeDeadline := ad.(*time.Duration)
		s.args.ActiveDeadlineSeconds = int32(activeDeadline.Seconds())
	}

	// Get ttlSecondsAfterFinished
	if ft, ok := s.argValues["ttl-after-finished"]; ok {
		ttlAfterFinished := ft.(*time.Duration)
		s.args.TTLSecondsAfterFinished = int32(ttlAfterFinished.Seconds())
	}
	return nil
}

func (s *SubmitRayJobArgsBuilder) setPreStopCmd() error {
	if psc, ok := s.argValues["pre-stop-command"]; ok {
		preStopCmd := psc.(*string)
		s.args.PreStopCmd = []string{"/bin/sh", "-c", *preStopCmd}
	}
	return nil
}

func (s *SubmitRayJobArgsBuilder) check() error {
	if (s.args.HeadGroupSpec.Image == "" || s.args.WorkerGroupSpec.Image == "") && s.args.Image == "" {
		return fmt.Errorf("--image must be set when neither --head-image nor --worker-image is provided")
	}
	if s.args.ActiveDeadlineSeconds < 0 {
		return fmt.Errorf("--active-deadline-seconds is invalid")
	}
	if s.args.TTLSecondsAfterFinished < 0 {
		return fmt.Errorf("--ttl-after-finished is invalid")
	}
	if s.args.ShareMemory != "" {
		_, err := resource.ParseQuantity(s.args.ShareMemory)
		if err != nil {
			return fmt.Errorf("--share-memory is invalid")
		}
	}
	if s.args.AutoscalerOptions.Cpu != "" {
		_, err := resource.ParseQuantity(s.args.AutoscalerOptions.Cpu)
		if err != nil {
			return fmt.Errorf("--autoscaler-cpu is invalid")
		}
	}
	if s.args.AutoscalerOptions.Memory != "" {
		_, err := resource.ParseQuantity(s.args.AutoscalerOptions.Memory)
		if err != nil {
			return fmt.Errorf("--autoscaler-memory is invalid")
		}
	}
	// check autoscaler-image-pull-policy
	switch s.args.AutoscalerOptions.ImagePullPolicy {
	case "Always", "IfNotPresent", "Never":
		log.Debugf("Supported imagePullPolicy: %s", s.args.AutoscalerOptions.ImagePullPolicy)
	default:
		return fmt.Errorf("unsupported imagePullPolicy: %s", s.args.AutoscalerOptions.ImagePullPolicy)
	}
	// check autoscaler-upscaling-mode
	switch s.args.AutoscalerOptions.UpscalingMode {
	case "Conservative", "Default", "Aggressive":
		log.Debugf("Supported autoscalerUpscalingMode: %s", s.args.AutoscalerOptions.UpscalingMode)
	default:
		return fmt.Errorf("unsupported autoscalerUpscalingMode: %s", s.args.AutoscalerOptions.UpscalingMode)
	}
	if s.args.HeadGroupSpec.Gpu < 0 {
		return fmt.Errorf("--head-gpu is invalid")
	}
	if s.args.HeadGroupSpec.Cpu != "" {
		_, err := resource.ParseQuantity(s.args.HeadGroupSpec.Cpu)
		if err != nil {
			return fmt.Errorf("--head-cpu is invalid")
		}
	}
	if s.args.HeadGroupSpec.Memory != "" {
		_, err := resource.ParseQuantity(s.args.HeadGroupSpec.Memory)
		if err != nil {
			return fmt.Errorf("--head-memory is invalid")
		}
	}
	// check head-service-type
	switch s.args.HeadGroupSpec.ServiceType {
	case "ClusterIP", "NodePort", "LoadBalancer", "ExternalName":
		log.Debugf("Supported headServiceType: %s", s.args.HeadGroupSpec.ServiceType)
	default:
		return fmt.Errorf("unsupported headServiceType: %s", s.args.HeadGroupSpec.ServiceType)
	}
	if s.args.WorkerGroupSpec.Gpu < 0 {
		return fmt.Errorf("--worker-gpu is invalid")
	}
	if s.args.WorkerGroupSpec.Cpu != "" {
		_, err := resource.ParseQuantity(s.args.WorkerGroupSpec.Cpu)
		if err != nil {
			return fmt.Errorf("--worker-cpu is invalid")
		}
	}
	if s.args.WorkerGroupSpec.Memory != "" {
		_, err := resource.ParseQuantity(s.args.WorkerGroupSpec.Memory)
		if err != nil {
			return fmt.Errorf("--worker-memory is invalid")
		}
	}
	if s.args.WorkerGroupSpec.Replicas < 0 {
		return fmt.Errorf("--worker-replicas is invalid")
	}
	if s.args.WorkerGroupSpec.MinReplicas < 0 {
		return fmt.Errorf("--worker-min-replicas is invalid")
	}
	if s.args.WorkerGroupSpec.MaxReplicas < 0 {
		return fmt.Errorf("--worker-max-replicas is invalid")
	}
	if s.args.WorkerGroupSpec.NumOfHosts < 0 {
		return fmt.Errorf("--worker-num-of-hosts is invalid")
	}
	return nil
}

func (s *SubmitRayJobArgsBuilder) addRequestGPUsToAnnotation() error {
	headGpu, workerGpu := s.args.HeadGroupSpec.Gpu, s.args.WorkerGroupSpec.Gpu
	gpus := headGpu + workerGpu*int(s.args.WorkerGroupSpec.Replicas)
	s.args.Annotations[types.RequestGPUsOfJobAnnoKey] = fmt.Sprintf("%v", gpus)
	return nil
}

func (s *SubmitRayJobArgsBuilder) addHeadAddr() error {
	if s.args.EnableRDMA {
		if s.args.Envs == nil {
			s.args.Envs = map[string]string{}
		}
		s.args.Envs["HEAD_ADDR"] = fmt.Sprintf("%v-head-0", s.args.Name)
	}
	return nil
}
