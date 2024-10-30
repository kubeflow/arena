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
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type UpdateDistributedServingArgsBuilder struct {
	args        *types.UpdateDistributedServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewUpdateDistributedServingArgsBuilder(args *types.UpdateDistributedServingArgs) ArgsBuilder {
	args.Type = types.DistributedServingJob
	s := &UpdateDistributedServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewUpdateServingArgsBuilder(&s.args.CommonUpdateServingArgs),
	)
	return s
}

func (s *UpdateDistributedServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *UpdateDistributedServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *UpdateDistributedServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *UpdateDistributedServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}

	command.Flags().IntVar(&s.args.Workers, "workers", 0, "the number of the worker pods")
	command.Flags().StringVar(&s.args.MasterCpu, "master-cpu", "", "the cpu resource to use for the master pods, like 1 for 1 core")
	command.Flags().StringVar(&s.args.WorkerCpu, "worker-cpu", "", "the cpu resource to use for the worker pods, like 1 for 1 core")
	command.Flags().IntVar(&s.args.MasterGPUCount, "master-gpus", 0, "the gpu resource to use for the master pods, like 1 for 1 gpu")
	command.Flags().IntVar(&s.args.WorkerGPUCount, "worker-gpus", 0, "the gpu resource to use for the master pods, like 1 for 1 gpu")
	command.Flags().IntVar(&s.args.MasterGPUMemory, "master-gpumemory", 0, "the limit GPU memory of master pod to run the serve.")
	command.Flags().IntVar(&s.args.WorkerGPUMemory, "worker-gpumemory", 0, "the limit GPU memory of each worker pods to run the serve.")
	command.Flags().IntVar(&s.args.MasterGPUCore, "master-gpucore", 0, "the limit GPU core of master pod to run the serve.")
	command.Flags().IntVar(&s.args.WorkerGPUCore, "worker-gpucore", 0, "the limit GPU core of each worker pods to run the serve.")
	command.Flags().StringVar(&s.args.MasterMemory, "master-memory", "", "the memory resource to use for the master pods, like 1Gi")
	command.Flags().StringVar(&s.args.WorkerMemory, "worker-memory", "", "the memory resource to use for the worker pods, like 1Gi")
	command.Flags().StringVar(&s.args.MasterCommand, "master-command", "", "the command to run for the master pod")
	command.Flags().StringVar(&s.args.WorkerCommand, "worker-command", "", "the command to run of each worker pods")

	_ = command.Flags().MarkHidden("cpu")
	_ = command.Flags().MarkHidden("memory")
	_ = command.Flags().MarkHidden("gpus")
	_ = command.Flags().MarkHidden("gpumemory")
	_ = command.Flags().MarkHidden("gpucore")
}

func (s *UpdateDistributedServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}

	return nil
}

func (s *UpdateDistributedServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.setCommand(); err != nil {
		return err
	}

	return nil
}

func (s *UpdateDistributedServingArgsBuilder) setCommand() error {
	if s.args.Command != "" {
		s.args.MasterCommand = s.args.Command
		s.args.WorkerCommand = s.args.Command
	}
	return nil
}

func (s *UpdateDistributedServingArgsBuilder) check() error {
	if s.args.Workers < 0 {
		return fmt.Errorf("--workers can not be negative")
	}
	if s.args.Command != "" {
		if s.args.MasterCommand != "" || s.args.WorkerCommand != "" {
			return fmt.Errorf("--command and --master-command/--worker-command can not be set at the same time")
		}
	}
	return nil
}
