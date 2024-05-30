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

type UpdateTritonServingJobBuilder struct {
	args      *types.UpdateTritonServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewUpdateTritonServingJobBuilder() *UpdateTritonServingJobBuilder {
	args := &types.UpdateTritonServingArgs{
		CommonUpdateServingArgs: types.CommonUpdateServingArgs{
			Image:     argsbuilder.DefaultTfServingImage,
			Replicas:  1,
			Namespace: "default",
		},
	}
	return &UpdateTritonServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewUpdateTritonServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *UpdateTritonServingJobBuilder) Name(name string) *UpdateTritonServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *UpdateTritonServingJobBuilder) Namespace(namespace string) *UpdateTritonServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *UpdateTritonServingJobBuilder) Shell(shell string) *UpdateTritonServingJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *UpdateTritonServingJobBuilder) Command(args []string) *UpdateTritonServingJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *UpdateTritonServingJobBuilder) Image(image string) *UpdateTritonServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *UpdateTritonServingJobBuilder) Envs(envs map[string]string) *UpdateTritonServingJobBuilder {
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
func (b *UpdateTritonServingJobBuilder) Annotations(annotations map[string]string) *UpdateTritonServingJobBuilder {
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
func (b *UpdateTritonServingJobBuilder) Labels(labels map[string]string) *UpdateTritonServingJobBuilder {
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
func (b *UpdateTritonServingJobBuilder) Replicas(count int) *UpdateTritonServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// Version is used to set serving job version,match the option --version
func (b *UpdateTritonServingJobBuilder) Version(version string) *UpdateTritonServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// ModelRepository is used to set model store,match the option --model-repository
func (b *UpdateTritonServingJobBuilder) ModelRepository(modelRepository string) *UpdateTritonServingJobBuilder {
	if modelRepository != "" {
		b.args.ModelRepository = modelRepository
	}
	return b
}

// AllowMetrics is enable metric,match the option --allow-metrics
func (b *UpdateTritonServingJobBuilder) AllowMetrics() *UpdateTritonServingJobBuilder {
	b.args.AllowMetrics = true
	return b
}

// Build is used to build the job
func (b *UpdateTritonServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.TritonServingJob, b.args), nil
}
