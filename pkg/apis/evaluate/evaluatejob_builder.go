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

package evaluate

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type EvaluateJobBuilder struct {
	args      *types.EvaluateJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewEvaluateJobBuilder() *EvaluateJobBuilder {
	args := &types.EvaluateJobArgs{}
	return &EvaluateJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewEvaluateJobArgsBuilder(args),
	}
}

// Name is used to set job name, match option --name
func (e *EvaluateJobBuilder) Name(name string) *EvaluateJobBuilder {
	if name != "" {
		e.args.Name = name
	}
	return e
}

// Namespace is used to set job namespace, match option --namespace
func (e *EvaluateJobBuilder) Namespace(namespace string) *EvaluateJobBuilder {
	if namespace != "" {
		e.args.Namespace = namespace
	}
	return e
}

// Command is used to set job command
func (e *EvaluateJobBuilder) Command(args []string) *EvaluateJobBuilder {
	if e.args.Command == "" {
		e.args.Command = strings.Join(args, " ")
	}
	return e
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (e *EvaluateJobBuilder) WorkingDir(dir string) *EvaluateJobBuilder {
	if dir != "" {
		e.args.WorkingDir = dir
	}
	return e
}

// Envs is used to set env of job containers,match option --env
func (e *EvaluateJobBuilder) Envs(envs map[string]string) *EvaluateJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		e.argValues["env"] = &envSlice
	}
	return e
}

// Image is used to set job image,match the option --image
func (e *EvaluateJobBuilder) Image(image string) *EvaluateJobBuilder {
	if image != "" {
		e.args.Image = image
	}
	return e
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (e *EvaluateJobBuilder) Tolerations(tolerations []string) *EvaluateJobBuilder {
	e.argValues["toleration"] = &tolerations
	return e
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (e *EvaluateJobBuilder) NodeSelectors(selectors map[string]string) *EvaluateJobBuilder {
	if len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		e.argValues["selector"] = &selectorsSlice
	}
	return e
}

// Annotations is used to add annotations for job pods,match option --annotation
func (e *EvaluateJobBuilder) Annotations(annotations map[string]string) *EvaluateJobBuilder {
	if len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		e.argValues["annotation"] = &s
	}
	return e
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (e *EvaluateJobBuilder) DataDirs(volumes map[string]string) *EvaluateJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		e.argValues["data-dir"] = &s
	}
	return e
}

// Datas is used to mount host files to job containers,match option --data
func (e *EvaluateJobBuilder) Datas(volumes map[string]string) *EvaluateJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		e.argValues["data"] = &s
	}
	return e
}

// SyncImage is used to set syncing image,match option --sync-image
func (e *EvaluateJobBuilder) SyncImage(image string) *EvaluateJobBuilder {
	if image != "" {
		e.args.SyncImage = image
	}
	return e
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (e *EvaluateJobBuilder) SyncMode(mode string) *EvaluateJobBuilder {
	if mode != "" {
		e.args.SyncMode = mode
	}
	return e
}

// SyncSource is used to set syncing source,match option --sync-source
func (e *EvaluateJobBuilder) SyncSource(source string) *EvaluateJobBuilder {
	if source != "" {
		e.args.SyncSource = source
	}
	return e
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (e *EvaluateJobBuilder) ImagePullSecrets(secrets []string) *EvaluateJobBuilder {
	if secrets != nil {
		e.argValues["image-pull-secret"] = &secrets
	}
	return e
}

// ModelName is used to set job model name, match option --model-name
func (e *EvaluateJobBuilder) ModelName(modelName string) *EvaluateJobBuilder {
	if modelName != "" {
		e.args.ModelName = modelName
	}
	return e
}

// ModelPath is used to set job model path, match option --model-path
func (e *EvaluateJobBuilder) ModelPath(modelPath string) *EvaluateJobBuilder {
	if modelPath != "" {
		e.args.ModelPath = modelPath
	}
	return e
}

// ModelVersion is used to set job model version, match option --model-version
func (e *EvaluateJobBuilder) ModelVersion(modelVersion string) *EvaluateJobBuilder {
	if modelVersion != "" {
		e.args.ModelVersion = modelVersion
	}
	return e
}

// DatasetPath is used to set job dataset path, match option --dataset-path
func (e *EvaluateJobBuilder) DatasetPath(datasetPath string) *EvaluateJobBuilder {
	if datasetPath != "" {
		e.args.DatasetPath = datasetPath
	}
	return e
}

// MetricsPath is used to set job metrics path, match option --metrics-path
func (e *EvaluateJobBuilder) MetricsPath(metricsPath string) *EvaluateJobBuilder {
	if metricsPath != "" {
		e.args.MetricsPath = metricsPath
	}
	return e
}

// Cpu is used to set job cpu, match option --cpu
func (e *EvaluateJobBuilder) Cpu(cpu string) *EvaluateJobBuilder {
	if cpu != "" {
		e.args.Cpu = cpu
	}
	return e
}

// Memory is used to set job memory, match option --memory
func (e *EvaluateJobBuilder) Memory(memory string) *EvaluateJobBuilder {
	if memory != "" {
		e.args.Memory = memory
	}
	return e
}

// Gpu is used to set job gpu, match option --gpus
func (e *EvaluateJobBuilder) Gpu(gpu int) *EvaluateJobBuilder {
	if gpu > 0 {
		e.args.GPUCount = gpu
	}
	return e
}

// Build is used to build the job
func (e *EvaluateJobBuilder) Build() (*EvaluateJob, error) {
	for key, value := range e.argValues {
		e.AddArgValue(key, value)
	}
	if err := e.PreBuild(); err != nil {
		return nil, err
	}
	if err := e.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewEvaluateJob(e.args.Name, e.args), nil
}
