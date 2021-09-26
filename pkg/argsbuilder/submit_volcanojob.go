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
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type SubmitVolcanoJobArgsBuilder struct {
	args        *types.SubmitVolcanoJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitVolcanoJobArgsBuilder(args *types.SubmitVolcanoJobArgs) ArgsBuilder {
	args.TrainingType = types.VolcanoTrainingJob
	s := &SubmitVolcanoJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	return s
}

func (s *SubmitVolcanoJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitVolcanoJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitVolcanoJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitVolcanoJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		annotations []string
		labels      []string
	)
	command.Flags().StringVar(&s.args.Name, "name", "", "assign the job name")
	command.MarkFlagRequired("name")
	command.Flags().IntVar(&(s.args.MinAvailable), "minAvailable", 1, "The minimal available pods to run for this Job. default value is 1")
	command.Flags().MarkDeprecated("minAvailable", "please use --min-available instead")
	command.Flags().IntVar(&(s.args.MinAvailable), "min-available", 1, "The minimal available pods to run for this Job. default value is 1")
	command.Flags().StringVar(&s.args.Queue, "queue", "default", "Specifies the queue that will be used in the scheduler, default queue is used this leaves empty")
	command.Flags().StringVar(&s.args.SchedulerName, "schedulerName", "volcano", "Specifies the scheduler Name, default is volcano when not specified")
	command.Flags().MarkDeprecated("schedulerName", "please use --scheduler-name instead")
	command.Flags().StringVar(&s.args.SchedulerName, "scheduler-name", "volcano", "Specifies the scheduler Name, default is volcano when not specified")
	// each task related information name,image,replica number
	command.Flags().StringVar(&s.args.TaskName, "taskName", "task", "the task name of volcano job, default value is task")
	command.Flags().MarkDeprecated("taskName", "please use --task-name instead")
	command.Flags().StringVar(&s.args.TaskName, "task-name", "task", "the task name of volcano job, default value is task")
	command.Flags().StringArrayVar(&s.args.TaskImages, "taskImages", []string{"ubuntu", "nginx", "busybox"}, "the docker images of different tasks of volcano job. default used 3 tasks with ubuntu,nginx and busybox images")
	command.Flags().MarkDeprecated("taskImages", "please use --task-images instead")
	command.Flags().StringArrayVar(&s.args.TaskImages, "task-images", []string{"ubuntu", "nginx", "busybox"}, "the docker images of different tasks of volcano job. default used 3 tasks with ubuntu,nginx and busybox images")
	command.Flags().IntVar(&s.args.TaskReplicas, "taskReplicas", 1, "the task replica's number to run the distributed tasks. default value is 1")
	command.Flags().MarkDeprecated("taskReplicas", "please use --task-replicas instead")
	command.Flags().IntVar(&s.args.TaskReplicas, "task-replicas", 1, "the task replica's number to run the distributed tasks. default value is 1")
	// cpu and memory request
	command.Flags().StringVar(&s.args.TaskCPU, "taskCPU", "250m", "cpu request for each task replica / pod. default value is 250m")
	command.Flags().MarkDeprecated("taskCPU", "please use --task-cpu instead")
	command.Flags().StringVar(&s.args.TaskCPU, "task-cpu", "250m", "cpu request for each task replica / pod. default value is 250m")
	command.Flags().StringVar(&s.args.TaskMemory, "taskMemory", "128Mi", "memory request for each task replica/pod.default value is 128Mi)")
	command.Flags().MarkDeprecated("taskMemory", "please use --task-memory instead")
	command.Flags().StringVar(&s.args.TaskMemory, "task-memory", "128Mi", "memory request for each task replica/pod.default value is 128Mi)")
	command.Flags().IntVar(&s.args.TaskPort, "taskPort", 2222, "the task port number. default value is 2222")
	command.Flags().MarkDeprecated("taskPort", "please use --task-port instead")
	command.Flags().IntVar(&s.args.TaskPort, "task-port", 2222, "the task port number. default value is 2222")
	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, "the annotations")
	command.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "specify the label")
	s.AddArgValue("annotation", &annotations).AddArgValue("label", &labels)
}

func (s *SubmitVolcanoJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SubmitVolcanoJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.setAnnotations(); err != nil {
		return err
	}
	if err := s.setLabels(); err != nil {
		return err
	}
	if err := s.setUserNameAndUserId(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitVolcanoJobArgsBuilder) check() error {
	if len(s.args.TaskImages) == 0 {
		return fmt.Errorf("TaskImages should be set")
	}
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *SubmitVolcanoJobArgsBuilder) setAnnotations() error {
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	argKey := "annotation"
	var annotations *[]string
	item, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	annotations = item.(*[]string)
	if len(*annotations) <= 0 {
		return nil
	}
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	for key, val := range transformSliceToMap(*annotations, "=") {
		s.args.Annotations[key] = val
	}
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *SubmitVolcanoJobArgsBuilder) setLabels() error {
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	argKey := "label"
	var labels *[]string
	item, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	labels = item.(*[]string)
	if len(*labels) <= 0 {
		return nil
	}
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	for key, val := range transformSliceToMap(*labels, "=") {
		s.args.Labels[key] = val
	}
	return nil
}

func (s *SubmitVolcanoJobArgsBuilder) setUserNameAndUserId() error {
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	arenaConfiger := config.GetArenaConfiger()
	user := arenaConfiger.GetUser()
	s.args.Labels[types.UserNameIdLabel] = user.GetId()
	s.args.Annotations[types.UserNameNameLabel] = user.GetName()
	return nil
}
