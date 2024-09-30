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

type RayJobBuilder struct {
	args      *types.SubmitRayJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewRayJobBuilder() *RayJobBuilder {
	args := &types.SubmitRayJobArgs{
		CommonSubmitArgs: types.CommonSubmitArgs{
			Namespace:  "default",
			Shell:      "sh",
			WorkingDir: "/root",
		},
		SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
	}
	return &RayJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitRayJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *RayJobBuilder) Name(name string) *RayJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Shell is used to set bash or sh
func (b *RayJobBuilder) Shell(shell string) *RayJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *RayJobBuilder) Command(args []string) *RayJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *RayJobBuilder) WorkingDir(dir string) *RayJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *RayJobBuilder) Envs(envs map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) GPUCount(count int) *RayJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *RayJobBuilder) Image(image string) *RayJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *RayJobBuilder) Tolerations(tolerations []string) *RayJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *RayJobBuilder) ConfigFiles(files map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) NodeSelectors(selectors map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) Annotations(annotations map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) Labels(labels map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) Datas(volumes map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) DataDirs(volumes map[string]string) *RayJobBuilder {
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
func (b *RayJobBuilder) LogDir(dir string) *RayJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *RayJobBuilder) Priority(priority string) *RayJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma,match option --rdma
func (b *RayJobBuilder) EnableRDMA() *RayJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image,match option --sync-image
func (b *RayJobBuilder) SyncImage(image string) *RayJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (b *RayJobBuilder) SyncMode(mode string) *RayJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source,match option --sync-source
func (b *RayJobBuilder) SyncSource(source string) *RayJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *RayJobBuilder) EnableTensorboard() *RayJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to enable tensorboard image
func (b *RayJobBuilder) TensorboardImage(image string) *RayJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (b *RayJobBuilder) ImagePullSecrets(secrets []string) *RayJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = &secrets
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *RayJobBuilder) WorkerCount(count int) *RayJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// ActiveDeadlineSeconds match option --running-timeout
func (b *RayJobBuilder) ActiveDeadlineSeconds(act int32) *RayJobBuilder {
	if act > 0 {
		b.args.ActiveDeadlineSeconds = act
	}
	return b
}

// TTLSecondsAfterFinished match option --ttl-after-finished
func (b *RayJobBuilder) TTLSecondsAfterFinished(ttl int32) *RayJobBuilder {
	if ttl > 0 {
		b.args.TTLSecondsAfterFinished = ttl
	}
	return b
}

func (b *RayJobBuilder) ShareMemory(shm string) *RayJobBuilder {
	if shm != "" {
		b.args.ShareMemory = shm
	}
	return b
}

// Build is used to build the job
func (b *RayJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.RayJob, b.args), nil
}
