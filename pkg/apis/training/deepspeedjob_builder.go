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

package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type DeepSpeedJobBuilder struct {
	args      *types.SubmitDeepSpeedJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewDeepSpeedJobBuilder() *DeepSpeedJobBuilder {
	args := &types.SubmitDeepSpeedJobArgs{
		CommonSubmitArgs:      DefaultCommonSubmitArgs,
		SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
	}
	return &DeepSpeedJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitDeepSpeedJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *DeepSpeedJobBuilder) Name(name string) *DeepSpeedJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Shell is used to set bash or sh
func (b *DeepSpeedJobBuilder) Shell(shell string) *DeepSpeedJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *DeepSpeedJobBuilder) Command(args []string) *DeepSpeedJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *DeepSpeedJobBuilder) WorkingDir(dir string) *DeepSpeedJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *DeepSpeedJobBuilder) Envs(envs map[string]string) *DeepSpeedJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *DeepSpeedJobBuilder) GPUCount(count int) *DeepSpeedJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *DeepSpeedJobBuilder) Image(image string) *DeepSpeedJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *DeepSpeedJobBuilder) Tolerations(tolerations []string) *DeepSpeedJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *DeepSpeedJobBuilder) ConfigFiles(files map[string]string) *DeepSpeedJobBuilder {
	if len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *DeepSpeedJobBuilder) NodeSelectors(selectors map[string]string) *DeepSpeedJobBuilder {
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
func (b *DeepSpeedJobBuilder) Annotations(annotations map[string]string) *DeepSpeedJobBuilder {
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
func (b *DeepSpeedJobBuilder) Labels(labels map[string]string) *DeepSpeedJobBuilder {
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
func (b *DeepSpeedJobBuilder) Datas(volumes map[string]string) *DeepSpeedJobBuilder {
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
func (b *DeepSpeedJobBuilder) DataDirs(volumes map[string]string) *DeepSpeedJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// LogDir is used to set log directory,match option --logdir
func (b *DeepSpeedJobBuilder) LogDir(dir string) *DeepSpeedJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *DeepSpeedJobBuilder) Priority(priority string) *DeepSpeedJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma,match option --rdma
func (b *DeepSpeedJobBuilder) EnableRDMA() *DeepSpeedJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image,match option --sync-image
func (b *DeepSpeedJobBuilder) SyncImage(image string) *DeepSpeedJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (b *DeepSpeedJobBuilder) SyncMode(mode string) *DeepSpeedJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source,match option --sync-source
func (b *DeepSpeedJobBuilder) SyncSource(source string) *DeepSpeedJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *DeepSpeedJobBuilder) EnableTensorboard() *DeepSpeedJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to enable tensorboard image
func (b *DeepSpeedJobBuilder) TensorboardImage(image string) *DeepSpeedJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (b *DeepSpeedJobBuilder) ImagePullSecrets(secrets []string) *DeepSpeedJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = &secrets
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *DeepSpeedJobBuilder) WorkerCount(count int) *DeepSpeedJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// CPU assign cpu limits,match option --cpu
func (b *DeepSpeedJobBuilder) CPU(cpu string) *DeepSpeedJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *DeepSpeedJobBuilder) Memory(memory string) *DeepSpeedJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Build is used to build the job
func (b *DeepSpeedJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.DeepSpeedTrainingJob, b.args), nil
}
