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

type DistributedServingJobBuilder struct {
	args      *types.DistributedServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewDistributedServingJobBuilder() *DistributedServingJobBuilder {
	args := &types.DistributedServingArgs{
		CustomServingArgs: types.CustomServingArgs{
			CommonServingArgs: types.CommonServingArgs{
				ImagePullPolicy: "IfNotPresent",
				Replicas:        1,
				Shell:           "sh",
				Namespace:       "default",
			},
		},
	}
	return &DistributedServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewDistributedServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *DistributedServingJobBuilder) Name(name string) *DistributedServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *DistributedServingJobBuilder) Namespace(namespace string) *DistributedServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *DistributedServingJobBuilder) Shell(shell string) *DistributedServingJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *DistributedServingJobBuilder) Command(args []string) *DistributedServingJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *DistributedServingJobBuilder) GPUCount(count int) *DistributedServingJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *DistributedServingJobBuilder) GPUMemory(memory int) *DistributedServingJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// GPUCore is used to set gpu core for the job, match the option --gpucore
func (b *DistributedServingJobBuilder) GPUCore(core int) *DistributedServingJobBuilder {
	if core > 0 {
		b.args.GPUCore = core
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *DistributedServingJobBuilder) Image(image string) *DistributedServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *DistributedServingJobBuilder) ImagePullPolicy(policy string) *DistributedServingJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *DistributedServingJobBuilder) CPU(cpu string) *DistributedServingJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *DistributedServingJobBuilder) Memory(memory string) *DistributedServingJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *DistributedServingJobBuilder) Envs(envs map[string]string) *DistributedServingJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// EnvsFromSecret is used to set env of job containers,match option --env-from-secret
func (b *DistributedServingJobBuilder) EnvsFromSecret(envs map[string]string) *DistributedServingJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env-from-secret"] = &envSlice
	}
	return b
}

// Replicas is used to set serving job replicas,match the option --replicas
func (b *DistributedServingJobBuilder) Replicas(count int) *DistributedServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *DistributedServingJobBuilder) EnableIstio() *DistributedServingJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *DistributedServingJobBuilder) ExposeService() *DistributedServingJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *DistributedServingJobBuilder) Version(version string) *DistributedServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *DistributedServingJobBuilder) Tolerations(tolerations []string) *DistributedServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *DistributedServingJobBuilder) NodeSelectors(selectors map[string]string) *DistributedServingJobBuilder {
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
func (b *DistributedServingJobBuilder) Annotations(annotations map[string]string) *DistributedServingJobBuilder {
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
func (b *DistributedServingJobBuilder) Labels(labels map[string]string) *DistributedServingJobBuilder {
	if len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

// Datas is used to mount k8s pvc to job pods,match option --data
func (b *DistributedServingJobBuilder) Datas(volumes map[string]string) *DistributedServingJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

// DataSubPathExprs is used to mount k8s pvc subpath to job pods,match option data-subpath-expr
func (b *DistributedServingJobBuilder) DataSubPathExprs(exprs map[string]string) *DistributedServingJobBuilder {
	if len(exprs) != 0 {
		s := []string{}
		for key, value := range exprs {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-subpath-expr"] = &s
	}
	return b
}

func (b *DistributedServingJobBuilder) TempDirs(volumes map[string]string) *DistributedServingJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["temp-dir"] = &s
	}
	return b
}

func (b *DistributedServingJobBuilder) EmptyDirSubPathExprs(exprs map[string]string) *DistributedServingJobBuilder {
	if len(exprs) != 0 {
		s := []string{}
		for key, value := range exprs {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["temp-dir-subpath-expr"] = &s
	}
	return b
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (b *DistributedServingJobBuilder) DataDirs(volumes map[string]string) *DistributedServingJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// Port is used to set port,match the option --port
func (b *DistributedServingJobBuilder) Port(port int) *DistributedServingJobBuilder {
	if port > 0 {
		b.args.Port = port
	}
	return b
}

// RestfulPort is used to set restful port,match the option --restful-port
func (b *DistributedServingJobBuilder) RestfulPort(port int) *DistributedServingJobBuilder {
	if port > 0 {
		b.args.RestfulPort = port
	}
	return b
}

// MetricsPort is used to set metrics port,match the option --metrics-port
func (b *DistributedServingJobBuilder) MetricsPort(port int) *DistributedServingJobBuilder {
	if port > 0 {
		b.args.MetricsPort = port
	}
	return b
}

// Masters is used to set master pods number,match the option --masters
func (b *DistributedServingJobBuilder) Masters(masters int) *DistributedServingJobBuilder {
	if masters > 0 {
		b.args.Masters = masters
	}
	return b
}

// Workers is used to set worker pods number,match the option --workers
func (b *DistributedServingJobBuilder) Workers(workers int) *DistributedServingJobBuilder {
	if workers > 0 {
		b.args.Workers = workers
	}
	return b
}

// MasterCpu is used to set master pods cpu,match the option --master-cpu
func (b *DistributedServingJobBuilder) MasterCpu(cpu string) *DistributedServingJobBuilder {
	if cpu != "" {
		b.args.MasterCpu = cpu
	}
	return b
}

// WorkerCpu is used to set worker pods cpu,match the option --worker-cpu
func (b *DistributedServingJobBuilder) WorkerCpu(cpu string) *DistributedServingJobBuilder {
	if cpu != "" {
		b.args.WorkerCpu = cpu
	}
	return b
}

// MasterGpus is used to set master pods gpus,match the option --master-gpus
func (b *DistributedServingJobBuilder) MasterGpus(gpus int) *DistributedServingJobBuilder {
	if gpus > 0 {
		b.args.MasterGPUCount = gpus
	}
	return b
}

// WorkerGpus is used to set worker pods gpus,match the option --worker-gpus
func (b *DistributedServingJobBuilder) WorkerGpus(gpus int) *DistributedServingJobBuilder {
	if gpus > 0 {
		b.args.WorkerGPUCount = gpus
	}
	return b
}

// MasterMemory is used to set master pods memory,match the option --master-memory
func (b *DistributedServingJobBuilder) MasterMemory(memory string) *DistributedServingJobBuilder {
	if memory != "" {
		b.args.MasterMemory = memory
	}
	return b
}

// WorkerMemory is used to set worker pods memory,match the option --worker-memory
func (b *DistributedServingJobBuilder) WorkerMemory(memory string) *DistributedServingJobBuilder {
	if memory != "" {
		b.args.WorkerMemory = memory
	}
	return b
}

// MasterGPUMemory is used to set master pods memory,match the option --master-gpumemory
func (b *DistributedServingJobBuilder) MasterGPUMemory(gpuMemory int) *DistributedServingJobBuilder {
	if gpuMemory > 0 {
		b.args.MasterGPUMemory = gpuMemory
	}
	return b
}

// WorkerGPUMemory is used to set worker pods memory,match the option --worker-gpumemory
func (b *DistributedServingJobBuilder) WorkerGPUMemory(gpuMemory int) *DistributedServingJobBuilder {
	if gpuMemory > 0 {
		b.args.WorkerGPUMemory = gpuMemory
	}
	return b
}

// MasterGPUCore is used to set master pods gpucore,match the option --master-gpucore
func (b *DistributedServingJobBuilder) MasterGPUCore(gpucore int) *DistributedServingJobBuilder {
	if gpucore > 0 {
		b.args.MasterGPUCore = gpucore
	}
	return b
}

// WorkerGPUCore is used to set worker pods gpucore,match the option --worker-gpucore
func (b *DistributedServingJobBuilder) WorkerGPUCore(gpucore int) *DistributedServingJobBuilder {
	if gpucore > 0 {
		b.args.WorkerGPUCore = gpucore
	}
	return b
}

// MasterCommand is used to set master pods command,match the option --master-command
func (b *DistributedServingJobBuilder) MasterCommand(command string) *DistributedServingJobBuilder {
	if command != "" {
		b.args.MasterCommand = command
	}
	return b
}

// WorkerCommand is used to set worker pods command,match the option --worker-command
func (b *DistributedServingJobBuilder) WorkerCommand(command string) *DistributedServingJobBuilder {
	if command != "" {
		b.args.WorkerCommand = command
	}
	return b
}

// InitBackend is used to set init backend,match the option --init-backend
func (b *DistributedServingJobBuilder) InitBackend(backend string) *DistributedServingJobBuilder {
	if backend != "" {
		b.args.InitBackend = backend
	}
	return b
}

// Build is used to build the job
func (b *DistributedServingJobBuilder) Build() (*Job, error) {
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
