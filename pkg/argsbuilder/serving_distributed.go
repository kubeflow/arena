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

package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/common"
)

type DistributedServingArgsBuilder struct {
	args        *types.DistributedServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewDistributedServingArgsBuilder(args *types.DistributedServingArgs) ArgsBuilder {
	args.Type = types.DistributedServingJob
	s := &DistributedServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewCustomServingArgsBuilder(&s.args.CustomServingArgs),
	)
	return s
}

func (s *DistributedServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *DistributedServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *DistributedServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *DistributedServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().IntVar(&s.args.Masters, "leader-num", 1, "the number of the leader pods (p.s. only support 1 leader currently)")
	command.Flags().IntVar(&s.args.Workers, "worker-num", 0, "the number of the worker pods")
	command.Flags().StringVar(&s.args.MasterCpu, "leader-cpu", "", "the cpu resource to use for the leader pod, like 1 for 1 core")
	command.Flags().StringVar(&s.args.WorkerCpu, "worker-cpu", "", "the cpu resource to use for each worker pods, like 1 for 1 core")
	command.Flags().IntVar(&s.args.MasterGPUCount, "leader-gpus", 0, "the gpu resource to use for the leader pod, like 1 for 1 gpu")
	command.Flags().IntVar(&s.args.WorkerGPUCount, "worker-gpus", 0, "the gpu resource to use for each worker pods, like 1 for 1 gpu")
	command.Flags().StringVar(&s.args.MasterMemory, "leader-memory", "", "the memory resource to use for the leader pod, like 1Gi")
	command.Flags().StringVar(&s.args.WorkerMemory, "worker-memory", "", "the memory resource to use for the worker pods, like 1Gi")
	command.Flags().IntVar(&s.args.MasterGPUMemory, "leader-gpumemory", 0, "the limit GPU memory of leader pod to run the serve")
	command.Flags().IntVar(&s.args.WorkerGPUMemory, "worker-gpumemory", 0, "the limit GPU memory of each worker pods to run the serve")
	command.Flags().IntVar(&s.args.MasterGPUCore, "leader-gpucore", 0, "the limit GPU core of leader pod to run the serve")
	command.Flags().IntVar(&s.args.WorkerGPUCore, "worker-gpucore", 0, "the limit GPU core of each worker pods to run the serve")
	command.Flags().StringVar(&s.args.MasterCommand, "leader-command", "", "the command to run for the leader pod")
	command.Flags().StringVar(&s.args.WorkerCommand, "worker-command", "", "the command to run of each worker pods")
	command.Flags().StringVar(&s.args.InitBackend, "init-backend", "", "specify the init backend for distributed serving job. Currently only support ray. support: ray")

	_ = command.Flags().MarkHidden("cpu")
	_ = command.Flags().MarkHidden("memory")
	_ = command.Flags().MarkHidden("gpus")
	_ = command.Flags().MarkHidden("gpumemory")
	_ = command.Flags().MarkHidden("gpucore")
}

func (s *DistributedServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *DistributedServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.setType(); err != nil {
		return err
	}
	if err := s.setNvidiaENV(); err != nil {
		return err
	}
	if err := s.setCommand(); err != nil {
		return err
	}
	return nil
}

func (s *DistributedServingArgsBuilder) setType() error {
	s.args.Type = types.DistributedServingJob
	return nil
}

func (s *DistributedServingArgsBuilder) setCommand() error {
	if s.args.Command != "" {
		s.args.MasterCommand = s.args.Command
		s.args.WorkerCommand = s.args.Command
	}
	return nil
}

func (s *DistributedServingArgsBuilder) setNvidiaENV() error {
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}
	// Since master and worker share the same envs, but they may have
	// different gpu resource, we delete the NVIDIA_VISIBLE_DEVICES env
	// and set it in helm chart manually
	delete(s.args.Envs, common.ENV_NVIDIA_VISIBLE_DEVICES)
	return nil
}

func (s *DistributedServingArgsBuilder) check() error {
	if s.args.Masters != 1 {
		return fmt.Errorf("can not change leader number, only support 1 leader currently")
	}
	if s.args.Command != "" {
		if s.args.MasterCommand != "" || s.args.WorkerCommand != "" {
			return fmt.Errorf("--command and --leader-command/--worker-command can not be set at the same time")
		}
	} else {
		if s.args.MasterCommand == "" || s.args.WorkerCommand == "" {
			return fmt.Errorf("--command or --leader-command/--worker-command must be set")
		}
	}
	if s.args.MasterGPUCount < 0 || s.args.WorkerGPUCount < 0 {
		return fmt.Errorf("--leader-gpus/--worker-gpus is invalid")
	}
	if s.args.MasterGPUMemory < 0 || s.args.WorkerGPUMemory < 0 {
		return fmt.Errorf("--leader-gpumemory/--worker-gpumemory is invalid")
	}
	if s.args.MasterGPUCore < 0 || s.args.WorkerGPUCore < 0 {
		return fmt.Errorf("--leader-gpucore/--worker-gpucore is invalid")
	}
	if s.args.InitBackend != "" {
		if s.args.InitBackend != "ray" {
			return fmt.Errorf("invalid init-backend value: %s, support: ray. ", s.args.InitBackend)
		}
	}
	return nil
}
