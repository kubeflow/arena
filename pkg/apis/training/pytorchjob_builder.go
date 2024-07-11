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

type PytorchJobBuilder struct {
	args      *types.SubmitPyTorchJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewPytorchJobBuilder() *PytorchJobBuilder {
	args := &types.SubmitPyTorchJobArgs{
		CleanPodPolicy:        "Running",
		CommonSubmitArgs:      DefaultCommonSubmitArgs,
		SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
	}
	return &PytorchJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitPytorchJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *PytorchJobBuilder) Name(name string) *PytorchJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Shell is used to set bash or sh
func (b *PytorchJobBuilder) Shell(shell string) *PytorchJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *PytorchJobBuilder) Command(args []string) *PytorchJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *PytorchJobBuilder) WorkingDir(dir string) *PytorchJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *PytorchJobBuilder) Envs(envs map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) GPUCount(count int) *PytorchJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *PytorchJobBuilder) Image(image string) *PytorchJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *PytorchJobBuilder) Tolerations(tolerations []string) *PytorchJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *PytorchJobBuilder) ConfigFiles(files map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) NodeSelectors(selectors map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) Annotations(annotations map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) Labels(labels map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) Datas(volumes map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) DataDirs(volumes map[string]string) *PytorchJobBuilder {
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
func (b *PytorchJobBuilder) LogDir(dir string) *PytorchJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *PytorchJobBuilder) Priority(priority string) *PytorchJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma,match option --rdma
func (b *PytorchJobBuilder) EnableRDMA() *PytorchJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image,match option --sync-image
func (b *PytorchJobBuilder) SyncImage(image string) *PytorchJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (b *PytorchJobBuilder) SyncMode(mode string) *PytorchJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source,match option --sync-source
func (b *PytorchJobBuilder) SyncSource(source string) *PytorchJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *PytorchJobBuilder) EnableTensorboard() *PytorchJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to enable tensorboard image
func (b *PytorchJobBuilder) TensorboardImage(image string) *PytorchJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (b *PytorchJobBuilder) ImagePullSecrets(secrets []string) *PytorchJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = &secrets
	}
	return b
}

// CleanPodPolicy is used to set cleaning pod policy,match option --clean-task-policy
func (b *PytorchJobBuilder) CleanPodPolicy(policy string) *PytorchJobBuilder {
	if policy != "" {
		b.args.CleanPodPolicy = policy
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *PytorchJobBuilder) WorkerCount(count int) *PytorchJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// CPU assign cpu limts,match option --cpu
func (b *PytorchJobBuilder) CPU(cpu string) *PytorchJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *PytorchJobBuilder) Memory(memory string) *PytorchJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// ActiveDeadlineSeconds match option --running-timeout
func (b *PytorchJobBuilder) ActiveDeadlineSeconds(act int64) *PytorchJobBuilder {
	if act > 0 {
		b.args.ActiveDeadlineSeconds = act
	}
	return b
}

// TTLSecondsAfterFinished match option --ttl-after-finished
func (b *PytorchJobBuilder) TTLSecondsAfterFinished(ttl int32) *PytorchJobBuilder {
	if ttl > 0 {
		b.args.TTLSecondsAfterFinished = ttl
	}
	return b
}

func (b *PytorchJobBuilder) ShareMemory(shm string) *PytorchJobBuilder {
	if shm != "" {
		b.args.ShareMemory = shm
	}
	return b
}

// Build is used to build the job
func (b *PytorchJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.PytorchTrainingJob, b.args), nil
}
