// Copyright 2023 The Kubeflow Authors
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

package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubeflow/arena/pkg/apis/types"
)

type SubmitDeepSpeedJobArgsBuilder struct {
	args        *types.SubmitDeepSpeedJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitDeepSpeedJobArgsBuilder(args *types.SubmitDeepSpeedJobArgs) ArgsBuilder {
	args.TrainingType = types.DeepSpeedTrainingJob
	s := &SubmitDeepSpeedJobArgsBuilder{
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

func (s *SubmitDeepSpeedJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitDeepSpeedJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitDeepSpeedJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitDeepSpeedJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		launcherSelectors   []string
		launcherAnnotations []string
		workerAnnotations   []string
	)

	command.Flags().StringVar(&s.args.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")
	command.Flags().StringArrayVarP(&launcherSelectors, "launcher-selector", "", []string{}, `assigning launcher pod to some k8s particular nodes, usage: "--launcher-selector=key=value" or "--launcher-selector key=value" `)
	command.Flags().StringVar(&s.args.JobRestartPolicy, "job-restart-policy", "", "deepspeed job restart policy, support: Never and OnFailure. default Never.")
	command.Flags().IntVar(&s.args.JobBackoffLimit, "job-backoff-limit", 6, "the max restart count of deepspeed job, default is six")
	command.Flags().StringVar(&s.args.SSHSecret, "ssh-secret", "", "Use an existing secret name for job ssh key.")
	command.Flags().StringArrayVar(&launcherAnnotations, "launcher-annotation", []string{}, `the launcher annotations, usage: "--launcher-annotation=key=value" or "--launcher-annotation key=value"`)
	command.Flags().StringArrayVar(&workerAnnotations, "worker-annotation", []string{}, `the worker annotations, usage: "--worker-annotation=key=value" or "--worker-annotation key=value"`)

	s.argValues["launcher-selector"] = &launcherSelectors
	s.argValues["launcher-annotation"] = &launcherAnnotations
	s.argValues["worker-annotation"] = &workerAnnotations
}

func (s *SubmitDeepSpeedJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	s.AddArgValue(ShareDataPrefix+"dataset", s.args.DataSet)
	return nil
}

func (s *SubmitDeepSpeedJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.setEnv(); err != nil {
		return nil
	}
	if err := s.setAnnotations(); err != nil {
		return err
	}
	if err := s.setLauncherAnnotations(); err != nil {
		return nil
	}
	if err := s.setWorkerAnnotations(); err != nil {
		return nil
	}
	if err := s.setLauncherSelectors(); err != nil {
		return nil
	}
	return nil
}

func (s *SubmitDeepSpeedJobArgsBuilder) check() error {
	if s.args.Image == "" {
		return fmt.Errorf("--image must be set ")
	}
	if s.args.GPUCount < 0 {
		return fmt.Errorf("--gpus is invalid")
	}
	if s.args.Cpu != "" {
		_, err := resource.ParseQuantity(s.args.Cpu)
		if err != nil {
			return fmt.Errorf("--cpu is invalid")
		}
	}
	if s.args.Memory != "" {
		_, err := resource.ParseQuantity(s.args.Memory)
		if err != nil {
			return fmt.Errorf("--memory is invalid")
		}
	}
	return nil
}

func (s *SubmitDeepSpeedJobArgsBuilder) setEnv() error {
	// avoid deepspeed job handing
	if _, ok := s.args.Envs[NCCLAsyncErrorHanding]; !ok {
		s.args.Envs[NCCLAsyncErrorHanding] = "1"
	}
	return nil
}

func (s *SubmitDeepSpeedJobArgsBuilder) setLauncherSelectors() error {
	log.Debug("begin setLauncherSelector")
	s.args.LauncherSelectors = map[string]string{}
	argKey := "launcher-selector"
	var LauncherSelectors *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		log.Warnf("Fail to get key: %s", argKey)
		return nil
	}
	LauncherSelectors = value.(*[]string)
	s.args.LauncherSelectors = transformSliceToMap(*LauncherSelectors, "=")
	log.Debugf("success to transform launcher selector: %v", s.args.LauncherSelectors)
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *SubmitDeepSpeedJobArgsBuilder) setAnnotations() error {
	if s.args.SSHSecret != "" {
		s.args.Annotations[types.SSHSecretName] = s.args.SSHSecret
	}
	return nil
}

// setLauncherAnnotations is used to handle option --launcher-annotation
func (s *SubmitDeepSpeedJobArgsBuilder) setLauncherAnnotations() error {
	if s.args.LauncherAnnotations == nil {
		s.args.LauncherAnnotations = map[string]string{}
	}
	item, ok := s.argValues["launcher-annotation"]
	if !ok {
		return nil
	}
	var annotations *[]string
	annotations = item.(*[]string)
	if len(*annotations) == 0 {
		return nil
	}
	for key, val := range transformSliceToMap(*annotations, "=") {
		s.args.LauncherAnnotations[key] = val
	}
	return nil
}

// setLauncherAnnotations is used to handle option --launcher-annotation
func (s *SubmitDeepSpeedJobArgsBuilder) setWorkerAnnotations() error {
	if s.args.WorkerAnnotations == nil {
		s.args.WorkerAnnotations = map[string]string{}
	}
	item, ok := s.argValues["worker-annotation"]
	if !ok {
		return nil
	}
	var annotations *[]string
	annotations = item.(*[]string)
	if len(*annotations) == 0 {
		return nil
	}
	for key, val := range transformSliceToMap(*annotations, "=") {
		s.args.WorkerAnnotations[key] = val
	}
	return nil
}
