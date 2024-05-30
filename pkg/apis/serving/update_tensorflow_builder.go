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

type UpdateTFServingJobBuilder struct {
	args      *types.UpdateTensorFlowServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewUpdateTFServingJobBuilder() *UpdateTFServingJobBuilder {
	args := &types.UpdateTensorFlowServingArgs{
		CommonUpdateServingArgs: types.CommonUpdateServingArgs{
			Image:     argsbuilder.DefaultTfServingImage,
			Replicas:  1,
			Namespace: "default",
		},
	}
	return &UpdateTFServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewUpdateTensorflowServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *UpdateTFServingJobBuilder) Name(name string) *UpdateTFServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *UpdateTFServingJobBuilder) Namespace(namespace string) *UpdateTFServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *UpdateTFServingJobBuilder) Shell(shell string) *UpdateTFServingJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *UpdateTFServingJobBuilder) Command(args []string) *UpdateTFServingJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *UpdateTFServingJobBuilder) Image(image string) *UpdateTFServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *UpdateTFServingJobBuilder) Envs(envs map[string]string) *UpdateTFServingJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// Annotations is used to add annotations for job pods,match option --annotation
func (b *UpdateTFServingJobBuilder) Annotations(annotations map[string]string) *UpdateTFServingJobBuilder {
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
func (b *UpdateTFServingJobBuilder) Labels(labels map[string]string) *UpdateTFServingJobBuilder {
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
func (b *UpdateTFServingJobBuilder) Replicas(count int) *UpdateTFServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// Version is used to set serving job version,match the option --version
func (b *UpdateTFServingJobBuilder) Version(version string) *UpdateTFServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// ModelConfigFile is used to set model config file,match the option --model-config-file
func (b *UpdateTFServingJobBuilder) ModelConfigFile(filePath string) *UpdateTFServingJobBuilder {
	if filePath != "" {
		b.args.ModelConfigFile = filePath
	}
	return b
}

// MonitoringConfigFile is used to set monitoring config file,match the option --monitoring-config-file
func (b *UpdateTFServingJobBuilder) MonitoringConfigFile(filePath string) *UpdateTFServingJobBuilder {
	if filePath != "" {
		b.args.MonitoringConfigFile = filePath
	}
	return b
}

// ModelName is used to set model name,match the option --model-name
func (b *UpdateTFServingJobBuilder) ModelName(name string) *UpdateTFServingJobBuilder {
	if name != "" {
		b.args.ModelName = name
	}
	return b
}

// ModelPath is used to set model path,match the option --model-path
func (b *UpdateTFServingJobBuilder) ModelPath(path string) *UpdateTFServingJobBuilder {
	if path != "" {
		b.args.ModelPath = path
	}
	return b
}

// Build is used to build the job
func (b *UpdateTFServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.TFServingJob, b.args), nil
}
