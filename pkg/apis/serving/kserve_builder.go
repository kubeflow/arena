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

type KServeJobBuilder struct {
	args      *types.KServeArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewKServeJobBuilder() *KServeJobBuilder {
	args := &types.KServeArgs{
		CommonServingArgs: types.CommonServingArgs{
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Namespace:       "default",
			Shell:           "sh",
		},
	}
	return &KServeJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewKServeArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *KServeJobBuilder) Name(name string) *KServeJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *KServeJobBuilder) Namespace(namespace string) *KServeJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *KServeJobBuilder) Shell(shell string) *KServeJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *KServeJobBuilder) Command(args []string) *KServeJobBuilder {
	// If the user does not specifies `--command`, args are used as container commands.
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *KServeJobBuilder) GPUCount(count int) *KServeJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *KServeJobBuilder) GPUMemory(memory int) *KServeJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// GPUCore is used to set gpu core for the job, match the option --gpucore
func (b *KServeJobBuilder) GPUCore(core int) *KServeJobBuilder {
	if core > 0 {
		b.args.GPUCore = core
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *KServeJobBuilder) Image(image string) *KServeJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *KServeJobBuilder) ImagePullPolicy(policy string) *KServeJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *KServeJobBuilder) CPU(cpu string) *KServeJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *KServeJobBuilder) Memory(memory string) *KServeJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *KServeJobBuilder) Envs(envs map[string]string) *KServeJobBuilder {
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
func (b *KServeJobBuilder) Replicas(count int) *KServeJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *KServeJobBuilder) EnableIstio() *KServeJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *KServeJobBuilder) ExposeService() *KServeJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *KServeJobBuilder) Version(version string) *KServeJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *KServeJobBuilder) Tolerations(tolerations []string) *KServeJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *KServeJobBuilder) NodeSelectors(selectors map[string]string) *KServeJobBuilder {
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
func (b *KServeJobBuilder) Annotations(annotations map[string]string) *KServeJobBuilder {
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
func (b *KServeJobBuilder) Labels(labels map[string]string) *KServeJobBuilder {
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
func (b *KServeJobBuilder) Datas(volumes map[string]string) *KServeJobBuilder {
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
func (b *KServeJobBuilder) DataDirs(volumes map[string]string) *KServeJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *KServeJobBuilder) ConfigFiles(files map[string]string) *KServeJobBuilder {
	if len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// ModelFormat the ModelFormat being served.
func (b *KServeJobBuilder) ModelFormat(modelFormat *types.ModelFormat) *KServeJobBuilder {
	if modelFormat != nil {
		b.args.ModelFormat = modelFormat
	}
	return b
}

// Runtime specific ClusterServingRuntime/ServingRuntime name to use for deployment.
func (b *KServeJobBuilder) Runtime(runtime string) *KServeJobBuilder {
	if runtime != "" {
		b.args.Runtime = runtime
	}
	return b
}

// StorageUri is used to set storage uri,match the option --storage-uri
func (b *KServeJobBuilder) StorageUri(uri string) *KServeJobBuilder {
	if uri != "" {
		b.args.StorageUri = uri
	}
	return b
}

// RuntimeVersion of the predictor docker image
func (b *KServeJobBuilder) RuntimeVersion(runtimeVersion string) *KServeJobBuilder {
	if runtimeVersion != "" {
		b.args.RuntimeVersion = runtimeVersion
	}
	return b
}

// ProtocolVersion use by the predictor (i.e. v1 or v2 or grpc-v1 or grpc-v2)
func (b *KServeJobBuilder) ProtocolVersion(protocolVersion string) *KServeJobBuilder {
	if protocolVersion != "" {
		b.args.ProtocolVersion = protocolVersion
	}
	return b
}

// MinReplicas number of replicas, defaults to 1 but can be set to 0 to enable scale-to-zero.
func (b *KServeJobBuilder) MinReplicas(minReplicas int) *KServeJobBuilder {
	if minReplicas >= 0 {
		b.args.MinReplicas = minReplicas
	}
	return b
}

// MaxReplicas number of replicas for autoscaling.
func (b *KServeJobBuilder) MaxReplicas(maxReplicas int) *KServeJobBuilder {
	if maxReplicas > 0 {
		b.args.MaxReplicas = maxReplicas
	}
	return b
}

// ScaleTarget number of replicas for autoscaling.
func (b *KServeJobBuilder) ScaleTarget(scaleTarget int) *KServeJobBuilder {
	if scaleTarget > 0 {
		b.args.ScaleTarget = scaleTarget
	}
	return b
}

// ScaleMetric watched by autoscaler. possible values are concurrency, rps, cpu, memory. concurrency, rps are supported via KPA
func (b *KServeJobBuilder) ScaleMetric(scaleMetric string) *KServeJobBuilder {
	if scaleMetric != "" {
		b.args.ScaleMetric = scaleMetric
	}
	return b
}

// ContainerConcurrency specifies how many requests can be processed concurrently
func (b *KServeJobBuilder) ContainerConcurrency(containerConcurrency int64) *KServeJobBuilder {
	if containerConcurrency > 0 {
		b.args.ContainerConcurrency = containerConcurrency
	}
	return b
}

// TimeoutSeconds specifies the number of seconds to wait before timing out a request to the component.
func (b *KServeJobBuilder) TimeoutSeconds(timeoutSeconds int64) *KServeJobBuilder {
	if timeoutSeconds > 0 {
		b.args.TimeoutSeconds = timeoutSeconds
	}
	return b
}

// CanaryTrafficPercent defines the traffic split percentage between the candidate revision and the last ready revision
func (b *KServeJobBuilder) CanaryTrafficPercent(canaryTrafficPercent int64) *KServeJobBuilder {
	if canaryTrafficPercent > 0 {
		b.args.CanaryTrafficPercent = canaryTrafficPercent
	}
	return b
}

// Port the port of tcp listening port, default is 8080 in kserve
func (b *KServeJobBuilder) Port(port int) *KServeJobBuilder {
	if port > 0 {
		b.args.Port = port
	}
	return b
}

// Build is used to build the job
func (b *KServeJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.KServeJob, b.args), nil
}
