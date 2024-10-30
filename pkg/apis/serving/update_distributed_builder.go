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
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type UpdateDistributedServingJobBuilder struct {
	args      *types.UpdateDistributedServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewUpdateDistributedServingJobBuilder() *UpdateDistributedServingJobBuilder {
	args := &types.UpdateDistributedServingArgs{
		CommonUpdateServingArgs: types.CommonUpdateServingArgs{
			Replicas: 1,
		},
	}
	return &UpdateDistributedServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewUpdateDistributedServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *UpdateDistributedServingJobBuilder) Name(name string) *UpdateDistributedServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *UpdateDistributedServingJobBuilder) Namespace(namespace string) *UpdateDistributedServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Version is used to set serving job version, match the option --version
func (b *UpdateDistributedServingJobBuilder) Version(version string) *UpdateDistributedServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Command is used to set job command
func (b *UpdateDistributedServingJobBuilder) Command(args []string) *UpdateDistributedServingJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *UpdateDistributedServingJobBuilder) Image(image string) *UpdateDistributedServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *UpdateDistributedServingJobBuilder) Envs(envs map[string]string) *UpdateDistributedServingJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// Tolerations are used to set tolerations for tolerate nodes, match option --toleration
func (b *UpdateDistributedServingJobBuilder) Tolerations(tolerations []string) *UpdateDistributedServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job, match option --selector
func (b *UpdateDistributedServingJobBuilder) NodeSelectors(selectors map[string]string) *UpdateDistributedServingJobBuilder {
	if len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["selector"] = &selectorsSlice
	}
	return b
}

// Annotations is used to add annotations for job pods,match option --annotation
func (b *UpdateDistributedServingJobBuilder) Annotations(annotations map[string]string) *UpdateDistributedServingJobBuilder {
	if len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

// Labels is used to add labels for job
func (b *UpdateDistributedServingJobBuilder) Labels(labels map[string]string) *UpdateDistributedServingJobBuilder {
	if len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

// Replicas is used to set serving job replicas,match the option --replicas
func (b *UpdateDistributedServingJobBuilder) Replicas(count int) *UpdateDistributedServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// Workers is used to set worker pods number,match the option --workers
func (b *UpdateDistributedServingJobBuilder) Workers(workers int) *UpdateDistributedServingJobBuilder {
	if workers > 0 {
		b.args.Workers = workers
	}
	return b
}

// MasterCpu is used to set master pods cpu,match the option --master-cpu
func (b *UpdateDistributedServingJobBuilder) MasterCpu(cpu string) *UpdateDistributedServingJobBuilder {
	if cpu != "" {
		b.args.MasterCpu = cpu
	}
	return b
}

// WorkerCpu is used to set worker pods cpu,match the option --worker-cpu
func (b *UpdateDistributedServingJobBuilder) WorkerCpu(cpu string) *UpdateDistributedServingJobBuilder {
	if cpu != "" {
		b.args.WorkerCpu = cpu
	}
	return b
}

// MasterGpus is used to set master pods gpus,match the option --master-gpus
func (b *UpdateDistributedServingJobBuilder) MasterGpus(gpus int) *UpdateDistributedServingJobBuilder {
	if gpus > 0 {
		b.args.MasterGPUCount = gpus
	}
	return b
}

// WorkerGpus is used to set worker pods gpus,match the option --worker-gpus
func (b *UpdateDistributedServingJobBuilder) WorkerGpus(gpus int) *UpdateDistributedServingJobBuilder {
	if gpus > 0 {
		b.args.WorkerGPUCount = gpus
	}
	return b
}

// MasterGPUMemory is used to set master pods memory,match the option --master-gpumemory
func (b *UpdateDistributedServingJobBuilder) MasterGPUMemory(gpuMemory int) *UpdateDistributedServingJobBuilder {
	if gpuMemory > 0 {
		b.args.MasterGPUMemory = gpuMemory
	}
	return b
}

// WorkerGPUMemory is used to set worker pods memory,match the option --worker-gpumemory
func (b *UpdateDistributedServingJobBuilder) WorkerGPUMemory(gpuMemory int) *UpdateDistributedServingJobBuilder {
	if gpuMemory > 0 {
		b.args.WorkerGPUMemory = gpuMemory
	}
	return b
}

// MasterGPUCore is used to set master pods gpucore,match the option --master-gpucore
func (b *UpdateDistributedServingJobBuilder) MasterGPUCore(gpucore int) *UpdateDistributedServingJobBuilder {
	if gpucore > 0 {
		b.args.MasterGPUCore = gpucore
	}
	return b
}

// WorkerGPUCore is used to set worker pods gpucore,match the option --worker-gpucore
func (b *UpdateDistributedServingJobBuilder) WorkerGPUCore(gpucore int) *UpdateDistributedServingJobBuilder {
	if gpucore > 0 {
		b.args.WorkerGPUCore = gpucore
	}
	return b
}

// MasterMemory is used to set master pods memory,match the option --master-memory
func (b *UpdateDistributedServingJobBuilder) MasterMemory(memory string) *UpdateDistributedServingJobBuilder {
	if memory != "" {
		b.args.MasterMemory = memory
	}
	return b
}

// WorkerMemory is used to set worker pods memory,match the option --worker-memory
func (b *UpdateDistributedServingJobBuilder) WorkerMemory(memory string) *UpdateDistributedServingJobBuilder {
	if memory != "" {
		b.args.WorkerMemory = memory
	}
	return b
}

// MasterCommand is used to set master pods command,match the option --master-command
func (b *UpdateDistributedServingJobBuilder) MasterCommand(command string) *UpdateDistributedServingJobBuilder {
	if command != "" {
		b.args.MasterCommand = command
	}
	return b
}

// WorkerCommand is used to set worker pods command,match the option --worker-command
func (b *UpdateDistributedServingJobBuilder) WorkerCommand(command string) *UpdateDistributedServingJobBuilder {
	if command != "" {
		b.args.WorkerCommand = command
	}
	return b
}

// Build is used to build the job
func (b *UpdateDistributedServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.DistributedServingJob, b.args), nil
}
