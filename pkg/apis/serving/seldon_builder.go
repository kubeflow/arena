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

type SeldonJobBuilder struct {
	args      *types.SeldonServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewSeldonServingJobBuilder() *SeldonJobBuilder {
	args := &types.SeldonServingArgs{
		Implementation: "TENSORFLOW_SERVER",
		CommonServingArgs: types.CommonServingArgs{
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Namespace:       "default",
			Shell:           "sh",
		},
	}
	return &SeldonJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSeldonServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *SeldonJobBuilder) Name(name string) *SeldonJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *SeldonJobBuilder) Namespace(namespace string) *SeldonJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *SeldonJobBuilder) Shell(shell string) *SeldonJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *SeldonJobBuilder) Command(args []string) *SeldonJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *SeldonJobBuilder) GPUCount(count int) *SeldonJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *SeldonJobBuilder) GPUMemory(memory int) *SeldonJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// GPUCore is used to set gpu core for the job, match the option --gpucore
func (b *SeldonJobBuilder) GPUCore(core int) *SeldonJobBuilder {
	if core > 0 {
		b.args.GPUCore = core
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *SeldonJobBuilder) Image(image string) *SeldonJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *SeldonJobBuilder) ImagePullPolicy(policy string) *SeldonJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *SeldonJobBuilder) CPU(cpu string) *SeldonJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *SeldonJobBuilder) Memory(memory string) *SeldonJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *SeldonJobBuilder) Envs(envs map[string]string) *SeldonJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// Replicas is used to set serving job replicas,match the option --replicas
func (b *SeldonJobBuilder) Replicas(count int) *SeldonJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *SeldonJobBuilder) EnableIstio() *SeldonJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *SeldonJobBuilder) ExposeService() *SeldonJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *SeldonJobBuilder) Version(version string) *SeldonJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *SeldonJobBuilder) Tolerations(tolerations []string) *SeldonJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *SeldonJobBuilder) NodeSelectors(selectors map[string]string) *SeldonJobBuilder {
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
func (b *SeldonJobBuilder) Annotations(annotations map[string]string) *SeldonJobBuilder {
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
func (b *SeldonJobBuilder) Labels(labels map[string]string) *SeldonJobBuilder {
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
func (b *SeldonJobBuilder) Datas(volumes map[string]string) *SeldonJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (b *SeldonJobBuilder) DataDirs(volumes map[string]string) *SeldonJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// Implementation defines the serving model framework --implementation
func (b *SeldonJobBuilder) Implementation(implementation string) *SeldonJobBuilder {
	if implementation != "" {
		b.args.Implementation = implementation
	}
	return b
}

// ModelUri defines the model uri --mode-uri
func (b *SeldonJobBuilder) ModelUri(modelUri string) *SeldonJobBuilder {
	if modelUri != "" {
		b.args.ModelUri = modelUri
	}
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *SeldonJobBuilder) ConfigFiles(files map[string]string) *SeldonJobBuilder {
	if len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// Build is used to build the job
func (b *SeldonJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.SeldonServingJob, b.args), nil
}
