// Copyright 2018 The Kubeflow Authors
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
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder/runtime"
	"github.com/kubeflow/arena/pkg/util"
)

const (
	disableTFConfigAnnotation = "arena.kubeflow.org/disable-tf-config"

	TFJobSuccessPolicyDefault     = ""
	TFJobSuccessPolicyChiefWorker = "ChiefWorker"
	TFJobSuccessPolicyAllWorkers  = "AllWorkers"
)

type SubmitTFJobArgsBuilder struct {
	args        *types.SubmitTFJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitTFJobArgsBuilder(args *types.SubmitTFJobArgs) ArgsBuilder {
	args.TrainingType = types.TFTrainingJob
	s := &SubmitTFJobArgsBuilder{
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

func (s *SubmitTFJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitTFJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitTFJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitTFJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		workerSelectors    []string
		chiefSelectors     []string
		psSelectors        []string
		evaluatorSelectors []string
		roleSequence       string
		runningTimeout     time.Duration
		startingTimeout    time.Duration
		ttlAfterFinished   time.Duration
	)
	command.Flags().StringVar(&s.args.WorkerImage, "workerImage", "", "the docker image for tensorflow workers")
	_ = command.Flags().MarkDeprecated("workerImage", "please use --worker-image instead")
	command.Flags().StringVar(&s.args.WorkerImage, "worker-image", "", "the docker image for tensorflow workers")

	command.Flags().StringVar(&s.args.PSImage, "psImage", "", "the docker image for tensorflow workers")
	_ = command.Flags().MarkDeprecated("psImage", "please use --ps-image instead")
	command.Flags().StringVar(&s.args.PSImage, "ps-image", "", "the docker image for tensorflow workers")

	command.Flags().IntVar(&s.args.PSCount, "ps", 0, "the number of the parameter servers.")

	command.Flags().IntVar(&s.args.PSPort, "psPort", 0, "the port of the parameter server.")
	_ = command.Flags().MarkDeprecated("psPort", "please use --ps-port instead")
	command.Flags().IntVar(&s.args.PSPort, "ps-port", 0, "the port of the parameter server.")

	command.Flags().IntVar(&s.args.WorkerPort, "workerPort", 0, "the port of the worker.")
	_ = command.Flags().MarkDeprecated("workerPort", "please use --worker-port instead")
	command.Flags().IntVar(&s.args.WorkerPort, "worker-port", 0, "the port of the worker.")

	command.Flags().StringVar(&s.args.WorkerCpu, "workerCpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")
	_ = command.Flags().MarkDeprecated("workerCpu", "please use --worker-cpu instead")
	command.Flags().StringVar(&s.args.WorkerCpu, "worker-cpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.WorkerCpuLimit, "worker-cpu-limit", "", "the cpu resource limit to use for the worker, like 1 for 1 core.")

	command.Flags().StringVar(&s.args.WorkerMemory, "workerMemory", "", "the memory resource to use for the worker, like 1Gi.")
	_ = command.Flags().MarkDeprecated("workerMemory", "please use --worker-memory instead")
	command.Flags().StringVar(&s.args.WorkerMemory, "worker-memory", "", "the memory resource to use for the worker, like 1Gi.")
	command.Flags().StringVar(&s.args.WorkerMemoryLimit, "worker-memory-limit", "", "the memory resource limit to use for the worker, like 1Gi.")

	command.Flags().StringVar(&s.args.PSCpu, "psCpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")
	_ = command.Flags().MarkDeprecated("psCpu", "please use --ps-cpu instead")
	command.Flags().StringVar(&s.args.PSCpu, "ps-cpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.PSCpuLimit, "ps-cpu-limit", "", "the cpu resource limit to use for the parameter servers, like 1 for 1 core.")

	command.Flags().IntVar(&s.args.PSGpu, "ps-gpus", 0, "the gpu resource to use for the parameter servers, like 1 for 1 gpu.")

	command.Flags().StringVar(&s.args.PSMemory, "psMemory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	_ = command.Flags().MarkDeprecated("psMemory", "please use --ps-memory instead")
	command.Flags().StringVar(&s.args.PSMemory, "ps-memory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	command.Flags().StringVar(&s.args.PSMemoryLimit, "ps-memory-limit", "", "the memory resource limit to use for the parameter servers, like 1Gi.")

	command.Flags().StringVar(&s.args.SuccessPolicy, "success-policy", TFJobSuccessPolicyChiefWorker, "Specifies the policy to mark the TFJob as succeeded. Available options are ChiefWorker and AllWorkers. Default to ChiefWorker.")

	// How to clean up Task
	command.Flags().StringVar(&s.args.CleanPodPolicy, "cleanTaskPolicy", "Running", "How to clean tasks after Training is done, support Running, None and All.")
	_ = command.Flags().MarkDeprecated("cleanTaskPolicy", "please use --clean-task-policy instead")
	command.Flags().StringVar(&s.args.CleanPodPolicy, "clean-task-policy", "Running", "How to clean tasks after Training is done, support Running, None and All.")

	command.Flags().DurationVar(&runningTimeout, "running-timeout", runningTimeout, "Specifies the duration since startTime during which the job can remain active before it is terminated(e.g. '5s', '1m', '2h22m').")
	command.Flags().DurationVar(&startingTimeout, "starting-timeout", startingTimeout, "Specifies the duration since createTime during which the job can remain pending before it is terminated(e.g. '5s', '1m', '2h22m').")
	command.Flags().DurationVar(&ttlAfterFinished, "ttl-after-finished", ttlAfterFinished, "Defines the TTL for cleaning up finished TFJobs(e.g. '5s', '1m', '2h22m'). Defaults to infinite.")

	// Estimator
	command.Flags().BoolVar(&s.args.UseChief, "chief", false, "enable chief, which is required for estimator.")
	command.Flags().BoolVar(&s.args.UseEvaluator, "evaluator", false, "enable evaluator, which is optional for estimator.")
	command.Flags().StringVar(&s.args.ChiefCpu, "ChiefCpu", "", "the cpu resource to use for the Chief, like 1 for 1 core.")
	_ = command.Flags().MarkDeprecated("ChiefCpu", "please use --chief-cpu instead")
	command.Flags().StringVar(&s.args.ChiefCpu, "chief-cpu", "", "the cpu resource to use for the Chief, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.ChiefCpuLimit, "chief-cpu-limit", "", "the cpu resource limit to use for the Chief, like 1 for 1 core.")

	command.Flags().StringVar(&s.args.ChiefMemory, "ChiefMemory", "", "the memory resource to use for the Chief, like 1Gi.")
	_ = command.Flags().MarkDeprecated("ChiefMemory", "please use --chief-memory instead")
	command.Flags().StringVar(&s.args.ChiefMemory, "chief-memory", "", "the memory resource to use for the Chief, like 1Gi.")
	command.Flags().StringVar(&s.args.ChiefMemoryLimit, "chief-memory-limit", "", "the memory liit resource to use for the Chief, like 1Gi.")

	command.Flags().StringVar(&s.args.EvaluatorCpu, "evaluatorCpu", "", "the cpu resource to use for the evaluator, like 1 for 1 core.")
	_ = command.Flags().MarkDeprecated("evaluatorCpu", "please use --evaluator-cpu instead")
	command.Flags().StringVar(&s.args.EvaluatorCpu, "evaluator-cpu", "", "the cpu resource to use for the evaluator, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.EvaluatorCpuLimit, "evaluator-cpu-limit", "", "the cpu resource limit to use for the evaluator, like 1 for 1 core.")

	command.Flags().StringVar(&s.args.EvaluatorMemory, "evaluatorMemory", "", "the memory resource to use for the evaluator, like 1Gi.")
	_ = command.Flags().MarkDeprecated("evaluatorMemory", "please use --evaluator-memory instead")
	command.Flags().StringVar(&s.args.EvaluatorMemory, "evaluator-memory", "", "the memory resource to use for the evaluator, like 1Gi.")
	command.Flags().StringVar(&s.args.EvaluatorMemoryLimit, "evaluator-memory-limit", "", "the memory resource limit to use for the evaluator, like 1Gi.")

	command.Flags().IntVar(&s.args.ChiefPort, "chiefPort", 0, "the port of the chief.")
	_ = command.Flags().MarkDeprecated("chiefPort", "please use --chief-port instead")
	command.Flags().IntVar(&s.args.ChiefPort, "chief-port", 0, "the port of the chief.")
	command.Flags().StringArrayVar(&workerSelectors, "worker-selector", []string{}, `assigning jobs with "Worker" role to some k8s particular nodes(this option would cover --selector), usage: "--worker-selector=key=value"`)
	command.Flags().StringArrayVar(&chiefSelectors, "chief-selector", []string{}, `assigning jobs with "Chief" role to some k8s particular nodes(this option would cover --selector), usage: "--chief-selector=key=value"`)
	command.Flags().StringArrayVar(&evaluatorSelectors, "evaluator-selector", []string{}, `assigning jobs with "Evaluator" role to some k8s particular nodes(this option would cover --selector), usage: "--evaluator-selector=key=value"`)
	command.Flags().StringArrayVar(&psSelectors, "ps-selector", []string{}, `assigning jobs with "PS" role to some k8s particular nodes(this option would cover --selector), usage: "--ps-selector=key=value"`)
	command.Flags().StringVar(&roleSequence, "role-sequence", "", `specify the tfjob role sequence,like: "Worker,PS,Chief,Evaluator" or "w,p,c,e"`)
	command.Flags().StringVar(&s.args.ShareMemory, "share-memory", "2Gi", "the shared memory of each replica to run the job, default 2Gi.")

	s.AddArgValue("worker-selector", &workerSelectors).
		AddArgValue("chief-selector", &chiefSelectors).
		AddArgValue("evaluator-selector", &evaluatorSelectors).
		AddArgValue("ps-selector", &psSelectors).
		AddArgValue("role-sequence", &roleSequence).
		AddArgValue("running-timeout", &runningTimeout).
		AddArgValue("starting-timeout", &startingTimeout).
		AddArgValue("ttl-after-finished", &ttlAfterFinished)
}

func (s *SubmitTFJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	s.AddArgValue(ShareDataPrefix+"dataset", s.args.DataSet)
	return nil
}

func (s *SubmitTFJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.setStandaloneMode(); err != nil {
		return err
	}
	if err := s.transform(); err != nil {
		return err
	}
	if err := s.setTFNodeSelectors(); err != nil {
		return err
	}
	if err := s.checkGangCapablitiesInCluster(); err != nil {
		return err
	}
	if err := s.setRuntime(); err != nil {
		return err
	}
	if err := s.setRunPolicy(); err != nil {
		return err
	}
	if err := s.checkRoleSequence(); err != nil {
		return err
	}
	if err := s.addRequestGPUsToAnnotation(); err != nil {
		return err
	}
	if err := s.addPodGroupLabel(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitTFJobArgsBuilder) setRuntime() error {
	// Get the runtime name
	annotations := s.args.CommonSubmitArgs.Annotations
	name := annotations["runtime"]
	s.args.TFRuntime = runtime.GetTFRuntime(name)
	return s.args.TFRuntime.Check(s.args)
}

func (s *SubmitTFJobArgsBuilder) setRunPolicy() error {
	// Get active deadline
	if rt, ok := s.argValues["running-timeout"]; ok {
		runningTimeout := rt.(*time.Duration)
		s.args.ActiveDeadlineSeconds = int64(runningTimeout.Seconds())
	}

	// Get starting deadline
	if sd, ok := s.argValues["starting-timeout"]; ok {
		startingTimeout := sd.(*time.Duration)
		s.args.StartingDeadlineSeconds = int64(startingTimeout.Seconds())
	}

	// Get ttlSecondsAfterFinished
	if ft, ok := s.argValues["ttl-after-finished"]; ok {
		ttlAfterFinished := ft.(*time.Duration)
		s.args.TTLSecondsAfterFinished = int32(ttlAfterFinished.Seconds())
	}

	return nil
}

func (s *SubmitTFJobArgsBuilder) check() error {
	switch s.args.SuccessPolicy {
	case TFJobSuccessPolicyDefault, TFJobSuccessPolicyAllWorkers:
		log.Debugf("Supported successPolicy: %s", s.args.SuccessPolicy)
	default:
		return fmt.Errorf("unsupported successPolicy %s", s.args.SuccessPolicy)
	}

	switch s.args.CleanPodPolicy {
	case "None", "Running", "All":
		log.Debugf("Supported cleanTaskPolicy: %s", s.args.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", s.args.CleanPodPolicy)
	}

	if s.args.WorkerCount == 0 && !s.args.UseChief {
		return fmt.Errorf("--workers must be greater than 0 in distributed training")
	}
	if s.args.WorkerImage == "" {
		return fmt.Errorf("--image or --workerImage must be set")
	}
	if s.args.PSCount > 0 {
		if s.args.PSImage == "" {
			return fmt.Errorf("--image or --psImage must be set")
		}
	}

	if s.args.GPUCount < 0 {
		return fmt.Errorf("--gpus is invalid")
	}
	if s.args.ChiefCpu != "" {
		_, err := resource.ParseQuantity(s.args.ChiefCpu)
		if err != nil {
			return fmt.Errorf("--chief-cpu is invalid")
		}
	}
	if s.args.ChiefMemory != "" {
		_, err := resource.ParseQuantity(s.args.ChiefMemory)
		if err != nil {
			return fmt.Errorf("--chief-memory is invalid")
		}
	}
	if s.args.PSCpu != "" {
		_, err := resource.ParseQuantity(s.args.PSCpu)
		if err != nil {
			return fmt.Errorf("--ps-cpu is invalid")
		}
	}
	if s.args.PSMemory != "" {
		_, err := resource.ParseQuantity(s.args.PSMemory)
		if err != nil {
			return fmt.Errorf("--ps-memory is invalid")
		}
	}
	if s.args.PSGpu < 0 {
		return fmt.Errorf("--ps-gpus is invalid")
	}
	if s.args.EvaluatorCpu != "" {
		_, err := resource.ParseQuantity(s.args.EvaluatorCpu)
		if err != nil {
			return fmt.Errorf("--evaluator-cpu is invalid")
		}
	}
	if s.args.EvaluatorMemory != "" {
		_, err := resource.ParseQuantity(s.args.EvaluatorMemory)
		if err != nil {
			return fmt.Errorf("--evaluator-memory is invalid")
		}
	}
	if s.args.WorkerCpu != "" {
		_, err := resource.ParseQuantity(s.args.WorkerCpu)
		if err != nil {
			return fmt.Errorf("--worker-cpu is invalid")
		}
	}
	if s.args.WorkerMemory != "" {
		_, err := resource.ParseQuantity(s.args.WorkerMemory)
		if err != nil {
			return fmt.Errorf("--worker-memory is invalid")
		}
	}
	if s.args.ActiveDeadlineSeconds < 0 {
		return fmt.Errorf("--running-timeout is invalid")
	}
	if s.args.StartingDeadlineSeconds < 0 {
		return fmt.Errorf("--starting-timeout is invalid")
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
	return nil
}

func (s *SubmitTFJobArgsBuilder) setStandaloneMode() error {
	if s.args.PSCount < 1 && s.args.WorkerCount == 1 {
		if s.args.Annotations == nil {
			s.args.Annotations = map[string]string{}
		}
		s.args.Annotations[disableTFConfigAnnotation] = "true"
		s.args.UseChief = true
		s.args.WorkerCount = 0
	}
	return nil
}

func (s *SubmitTFJobArgsBuilder) transform() error {
	arenaConfiger := config.GetArenaConfiger()
	if s.args.WorkerImage == "" {
		s.args.WorkerImage = s.args.Image
	}

	if s.args.WorkerCount > 0 {
		autoSelectWorkerPort, err := util.SelectAvailablePortWithDefault(arenaConfiger.GetClientSet(), s.args.WorkerPort)
		if err != nil {
			return fmt.Errorf("failed to select worker port: %++v", err)
		}
		s.args.WorkerPort = autoSelectWorkerPort
	}

	if s.args.UseChief {
		autoSelectChiefPort, err := util.SelectAvailablePortWithDefault(arenaConfiger.GetClientSet(), s.args.ChiefPort)
		if err != nil {
			return fmt.Errorf("failed to select chief port: %++v", err)
		}
		s.args.ChiefPort = autoSelectChiefPort
		s.args.ChiefCount = 1
	}

	if s.args.PSCount > 0 {
		autoSelectPsPort, err := util.SelectAvailablePortWithDefault(arenaConfiger.GetClientSet(), s.args.PSPort)
		if err != nil {
			return fmt.Errorf("failed to select ps port: %++v", err)
		}
		s.args.PSPort = autoSelectPsPort
		if s.args.PSImage == "" {
			s.args.PSImage = s.args.Image
		}
	}

	if s.args.UseEvaluator {
		s.args.EvaluatorCount = 1
	}

	if s.args.SuccessPolicy == TFJobSuccessPolicyChiefWorker {
		// The value of chief worker policy actually is empty string in training-operator.
		s.args.SuccessPolicy = TFJobSuccessPolicyDefault
	}

	return nil
}

func (s *SubmitTFJobArgsBuilder) checkGangCapablitiesInCluster() error {
	s.args.HasGangScheduler = false
	arenaConfiger := config.GetArenaConfiger()
	podList, err := arenaConfiger.GetClientSet().CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%v", gangSchdName),
	})
	if err != nil {
		log.Debugf("Failed to find %s due to %v", gangSchdName, err)
		return nil
	}
	if len(podList.Items) == 0 {
		log.Debugf("not found %v scheduler,it represents that you don't deploy it", gangSchdName)
		return nil
	}
	log.Debugf("Found %s successfully, the gang scheduler is enabled in the cluster.", gangSchdName)
	s.args.HasGangScheduler = true
	return nil
}

// add node selectors
func (s *SubmitTFJobArgsBuilder) setTFNodeSelectors() error {
	s.args.TFNodeSelectors = map[string]map[string]string{}
	var (
		psSelectors        *[]string
		workerSelectors    *[]string
		chiefSelectors     *[]string
		evaluatorSelectors *[]string
	)
	item1, ok := s.argValues["ps-selector"]
	if ok {
		psSelectors = item1.(*[]string)
	}
	item2, ok := s.argValues["worker-selector"]
	if ok {
		workerSelectors = item2.(*[]string)
	}
	item3, ok := s.argValues["chief-selector"]
	if ok {
		chiefSelectors = item3.(*[]string)
	}
	item4, ok := s.argValues["evaluator-selector"]
	if ok {
		evaluatorSelectors = item4.(*[]string)
	}
	for _, role := range []string{"PS", "Worker", "Evaluator", "Chief"} {
		switch {
		case role == "PS":
			s.transformSelectorArrayToMap(psSelectors, "PS")
		case role == "Worker":
			s.transformSelectorArrayToMap(workerSelectors, "Worker")
		case role == "Chief":
			s.transformSelectorArrayToMap(chiefSelectors, "Chief")
		case role == "Evaluator":
			s.transformSelectorArrayToMap(evaluatorSelectors, "Evaluator")
		}
	}
	return nil
}

func (s *SubmitTFJobArgsBuilder) transformSelectorArrayToMap(selectorArray *[]string, role string) {
	s.args.TFNodeSelectors[role] = map[string]string{}
	if selectorArray != nil && len(*selectorArray) != 0 {
		log.Debugf("%v Selectors: %v", role, selectorArray)
		s.args.TFNodeSelectors[role] = transformSliceToMap(*selectorArray, "=")
		return
	}
	// set the default node selectors to tf role node selectors
	log.Debugf("use to Node Selectors %v to %v Selector", s.args.NodeSelectors, role)
	s.args.TFNodeSelectors[role] = s.args.NodeSelectors

}

func (s *SubmitTFJobArgsBuilder) addRequestGPUsToAnnotation() error {
	gpus := 0
	gpus += s.args.ChiefCount * s.args.GPUCount
	gpus += s.args.EvaluatorCount * s.args.GPUCount
	gpus += s.args.WorkerCount * s.args.GPUCount
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	s.args.Annotations[types.RequestGPUsOfJobAnnoKey] = fmt.Sprintf("%v", gpus)
	return nil
}

func (s *SubmitTFJobArgsBuilder) checkRoleSequence() error {
	roleSequence := ""
	var getRoleSeqFromConfig = func() {
		configs := config.GetArenaConfiger().GetConfigsFromConfigFile()
		value, ok := configs["tfjob_role_sequence"]
		if !ok {
			return
		}
		log.Debugf("read tfjob role sequence from config file")
		roleSequence = strings.Trim(value, " ")
	}
	item, ok := s.argValues["role-sequence"]
	if !ok {
		getRoleSeqFromConfig()
	} else {
		v := item.(*string)
		roleSequence = *v
		if roleSequence == "" {
			getRoleSeqFromConfig()
		} else {
			log.Debugf("read tfjob role sequence from command option")
		}
	}
	if roleSequence == "" {
		return nil
	}
	roles := []string{}
	for _, r := range strings.Split(roleSequence, ",") {
		switch strings.ToLower(strings.Trim(r, " ")) {
		case "worker", "w":
			roles = append(roles, "Worker")
		case "ps", "p":
			roles = append(roles, "PS")
		case "evaluator", "e":
			roles = append(roles, "Evaluator")
		case "chief", "c":
			roles = append(roles, "Chief")
		default:
			return fmt.Errorf("Unknown role: %v, the tfjob only supports:[Worker(w)|PS(p)|Evaluator(e)|Chief(c)]", r)
		}
	}
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	s.args.Annotations["job-role-sequence"] = strings.Join(roles, ",")
	return nil
}

func (s *SubmitTFJobArgsBuilder) addPodGroupLabel() error {
	if s.args.Coscheduling {
		s.args.PodGroupName = fmt.Sprintf("%v-%v", s.args.TrainingType, s.args.Name)
		s.args.PodGroupMinAvailable = fmt.Sprintf("%v", s.args.WorkerCount+s.args.PSCount+s.args.ChiefCount+s.args.EvaluatorCount)
	}
	return nil
}
