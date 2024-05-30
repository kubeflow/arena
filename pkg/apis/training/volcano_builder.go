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

package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type VolcanoJobBuilder struct {
	args      *types.SubmitVolcanoJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewVolcanoJobBuilder() *VolcanoJobBuilder {
	args := &types.SubmitVolcanoJobArgs{
		MinAvailable:  1,
		Queue:         "default",
		SchedulerName: "volcano",
		TaskName:      "task",
		TaskImages:    []string{"ubuntu", "nginx", "busybox"},
		TaskReplicas:  1,
		TaskCPU:       "250m",
		TaskMemory:    "128Mi",
		TaskPort:      2222,
	}
	return &VolcanoJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitVolcanoJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *VolcanoJobBuilder) Name(name string) *VolcanoJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Command is used to set job command
func (b *VolcanoJobBuilder) Command(args []string) *VolcanoJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

func (b *VolcanoJobBuilder) MinAvailable(minAvailable int) *VolcanoJobBuilder {
	if minAvailable > 0 {
		b.args.MinAvailable = minAvailable
	}
	return b
}

func (b *VolcanoJobBuilder) Queue(queue string) *VolcanoJobBuilder {
	if queue != "" {
		b.args.Queue = queue
	}
	return b
}

func (b *VolcanoJobBuilder) SchedulerName(name string) *VolcanoJobBuilder {
	if name != "" {
		b.args.SchedulerName = name
	}
	return b
}

func (b *VolcanoJobBuilder) TaskImages(images []string) *VolcanoJobBuilder {
	if len(images) != 0 {
		b.args.TaskImages = images
	}
	return b
}

func (b *VolcanoJobBuilder) TaskName(name string) *VolcanoJobBuilder {
	if name != "" {
		b.args.TaskName = name
	}
	return b
}

func (b *VolcanoJobBuilder) TaskReplicas(replicas int) *VolcanoJobBuilder {
	if replicas > 0 {
		b.args.TaskReplicas = replicas
	}
	return b
}

func (b *VolcanoJobBuilder) TaskCPU(cpu string) *VolcanoJobBuilder {
	if cpu != "" {
		b.args.TaskCPU = cpu
	}
	return b
}

func (b *VolcanoJobBuilder) TaskMemory(mem string) *VolcanoJobBuilder {
	if mem != "" {
		b.args.TaskMemory = mem
	}
	return b
}

func (b *VolcanoJobBuilder) TaskPort(port int) *VolcanoJobBuilder {
	if port > 0 {
		b.args.TaskPort = port
	}
	return b
}

func (b *VolcanoJobBuilder) Labels(labels map[string]string) *VolcanoJobBuilder {
	if len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

func (b *VolcanoJobBuilder) Annotations(annotations map[string]string) *VolcanoJobBuilder {
	if len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

// Build is used to build the job
func (b *VolcanoJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.VolcanoTrainingJob, b.args), nil
}
