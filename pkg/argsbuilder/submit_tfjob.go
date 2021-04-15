// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder/runtime"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	)
	command.Flags().StringVar(&s.args.WorkerImage, "workerImage", "", "the docker image for tensorflow workers")
	command.Flags().MarkDeprecated("workerImage", "please use --worker-image instead")
	command.Flags().StringVar(&s.args.WorkerImage, "worker-image", "", "the docker image for tensorflow workers")

	command.Flags().StringVar(&s.args.PSImage, "psImage", "", "the docker image for tensorflow workers")
	command.Flags().MarkDeprecated("psImage", "please use --ps-image instead")
	command.Flags().StringVar(&s.args.PSImage, "ps-image", "", "the docker image for tensorflow workers")

	command.Flags().IntVar(&s.args.PSCount, "ps", 0, "the number of the parameter servers.")

	command.Flags().IntVar(&s.args.PSPort, "psPort", 0, "the port of the parameter server.")
	command.Flags().MarkDeprecated("psPort", "please use --ps-port instead")
	command.Flags().IntVar(&s.args.PSPort, "ps-port", 0, "the port of the parameter server.")

	command.Flags().IntVar(&s.args.WorkerPort, "workerPort", 0, "the port of the worker.")
	command.Flags().MarkDeprecated("workerPort", "please use --worker-port instead")
	command.Flags().IntVar(&s.args.WorkerPort, "worker-port", 0, "the port of the worker.")

	command.Flags().StringVar(&s.args.WorkerCpu, "workerCpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")
	command.Flags().MarkDeprecated("workerCpu", "please use --worker-cpu instead")
	command.Flags().StringVar(&s.args.WorkerCpu, "worker-cpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")

	command.Flags().StringVar(&s.args.WorkerMemory, "workerMemory", "", "the memory resource to use for the worker, like 1Gi.")
	command.Flags().MarkDeprecated("workerMemory", "please use --worker-memory instead")
	command.Flags().StringVar(&s.args.WorkerMemory, "worker-memory", "", "the memory resource to use for the worker, like 1Gi.")

	command.Flags().StringVar(&s.args.PSCpu, "psCpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")
	command.Flags().MarkDeprecated("psCpu", "please use --ps-cpu instead")
	command.Flags().StringVar(&s.args.PSCpu, "ps-cpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")

	command.Flags().IntVar(&s.args.PSGpu, "ps-gpus", 0, "the gpu resource to use for the parameter servers, like 1 for 1 gpu.")

	command.Flags().StringVar(&s.args.PSMemory, "psMemory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	command.Flags().MarkDeprecated("psMemory", "please use --ps-memory instead")
	command.Flags().StringVar(&s.args.PSMemory, "ps-memory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	// How to clean up Task
	command.Flags().StringVar(&s.args.CleanPodPolicy, "cleanTaskPolicy", "Running", "How to clean tasks after Training is done, only support Running, None.")
	command.Flags().MarkDeprecated("cleanTaskPolicy", "please use --clean-task-policy instead")
	command.Flags().StringVar(&s.args.CleanPodPolicy, "clean-task-policy", "Running", "How to clean tasks after Training is done, only support Running, None.")

	// Estimator
	command.Flags().BoolVar(&s.args.UseChief, "chief", false, "enable chief, which is required for estimator.")
	command.Flags().BoolVar(&s.args.UseEvaluator, "evaluator", false, "enable evaluator, which is optional for estimator.")
	command.Flags().StringVar(&s.args.ChiefCpu, "ChiefCpu", "", "the cpu resource to use for the Chief, like 1 for 1 core.")
	command.Flags().MarkDeprecated("ChiefCpu", "please use --chief-cpu instead")
	command.Flags().StringVar(&s.args.ChiefCpu, "chief-cpu", "", "the cpu resource to use for the Chief, like 1 for 1 core.")

	command.Flags().StringVar(&s.args.ChiefMemory, "ChiefMemory", "", "the memory resource to use for the Chief, like 1Gi.")
	command.Flags().MarkDeprecated("ChiefMemory", "please use --chief-memory instead")
	command.Flags().StringVar(&s.args.ChiefMemory, "chief-memory", "", "the memory resource to use for the Chief, like 1Gi.")

	command.Flags().StringVar(&s.args.EvaluatorCpu, "evaluatorCpu", "", "the cpu resource to use for the evaluator, like 1 for 1 core.")
	command.Flags().MarkDeprecated("evaluatorCpu", "please use --evaluator-cpu instead")
	command.Flags().StringVar(&s.args.EvaluatorCpu, "evaluator-cpu", "", "the cpu resource to use for the evaluator, like 1 for 1 core.")

	command.Flags().StringVar(&s.args.EvaluatorMemory, "evaluatorMemory", "", "the memory resource to use for the evaluator, like 1Gi.")
	command.Flags().MarkDeprecated("evaluatorMemory", "please use --evaluator-memory instead")
	command.Flags().StringVar(&s.args.EvaluatorMemory, "evaluator-memory", "", "the memory resource to use for the evaluator, like 1Gi.")

	command.Flags().IntVar(&s.args.ChiefPort, "chiefPort", 0, "the port of the chief.")
	command.Flags().MarkDeprecated("chiefPort", "please use --chief-port instead")
	command.Flags().IntVar(&s.args.ChiefPort, "chief-port", 0, "the port of the chief.")
	command.Flags().StringSliceVar(&workerSelectors, "worker-selector", []string{}, `assigning jobs with "Worker" role to some k8s particular nodes(this option would cover --selector), usage: "--worker-selector=key=value"`)
	command.Flags().StringSliceVar(&chiefSelectors, "chief-selector", []string{}, `assigning jobs with "Chief" role to some k8s particular nodes(this option would cover --selector), usage: "--chief-selector=key=value"`)
	command.Flags().StringSliceVar(&evaluatorSelectors, "evaluator-selector", []string{}, `assigning jobs with "Evaluator" role to some k8s particular nodes(this option would cover --selector), usage: "--evaluator-selector=key=value"`)
	command.Flags().StringSliceVar(&psSelectors, "ps-selector", []string{}, `assigning jobs with "PS" role to some k8s particular nodes(this option would cover --selector), usage: "--ps-selector=key=value"`)
	command.Flags().StringVar(&roleSequence, "role-sequence", "", `specify the tfjob role sequence,like: "Worker,PS,Chief,Evaluator" or "w,p,c,e"`)

	s.AddArgValue("worker-selector", &workerSelectors).
		AddArgValue("chief-selector", &chiefSelectors).
		AddArgValue("evaluator-selector", &evaluatorSelectors).
		AddArgValue("ps-selector", &psSelectors).
		AddArgValue("role-sequence", &roleSequence)
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
	if err := s.checkRoleSequence(); err != nil {
		return err
	}
	if err := s.addRequestGPUsToAnnotation(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitTFJobArgsBuilder) setCommand(args []string) error {
	s.args.CommonSubmitArgs.Command = strings.Join(args, " ")
	return nil
}

func (s *SubmitTFJobArgsBuilder) setRuntime() error {
	// Get the runtime name
	annotations := s.args.CommonSubmitArgs.Annotations
	name := annotations["runtime"]
	s.args.TFRuntime = runtime.GetTFRuntime(name)
	return s.args.TFRuntime.Check(s.args)
}

func (s *SubmitTFJobArgsBuilder) check() error {
	switch s.args.CleanPodPolicy {
	case "None", "Running":
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
	return nil
}

func (s *SubmitTFJobArgsBuilder) setStandaloneMode() error {
	if s.args.PSCount < 1 && s.args.WorkerCount == 1 {
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
	item4, ok := s.argValues["chief-selector"]
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
	gpus += s.args.ChiefCount
	gpus += s.args.EvaluatorCount
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
