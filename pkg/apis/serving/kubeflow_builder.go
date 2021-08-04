package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type KFServingJobBuilder struct {
	args      *types.KFServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewKFServingJobBuilder() *KFServingJobBuilder {
	args := &types.KFServingArgs{
		ModelType: "custom",
		CommonServingArgs: types.CommonServingArgs{
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Namespace:       "default",
		},
	}
	return &KFServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewKFServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *KFServingJobBuilder) Name(name string) *KFServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *KFServingJobBuilder) Namespace(namespace string) *KFServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Command is used to set job command
func (b *KFServingJobBuilder) Command(args []string) *KFServingJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *KFServingJobBuilder) GPUCount(count int) *KFServingJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *KFServingJobBuilder) GPUMemory(memory int) *KFServingJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *KFServingJobBuilder) Image(image string) *KFServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *KFServingJobBuilder) ImagePullPolicy(policy string) *KFServingJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *KFServingJobBuilder) CPU(cpu string) *KFServingJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *KFServingJobBuilder) Memory(memory string) *KFServingJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *KFServingJobBuilder) Envs(envs map[string]string) *KFServingJobBuilder {
	if envs != nil && len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// Replicas is used to set serving job replicas,match the option --replicas
func (b *KFServingJobBuilder) Replicas(count int) *KFServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *KFServingJobBuilder) EnableIstio() *KFServingJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *KFServingJobBuilder) ExposeService() *KFServingJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *KFServingJobBuilder) Version(version string) *KFServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *KFServingJobBuilder) Tolerations(tolerations []string) *KFServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *KFServingJobBuilder) NodeSelectors(selectors map[string]string) *KFServingJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["selector"] = &selectorsSlice
	}
	return b
}

// Annotations is used to add annotations for job pods,match option --annotation
func (b *KFServingJobBuilder) Annotations(annotations map[string]string) *KFServingJobBuilder {
	if annotations != nil && len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

// Labels is used to add labels for job
func (b *KFServingJobBuilder) Labels(labels map[string]string) *KFServingJobBuilder {
	if labels != nil && len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

// Datas is used to mount k8s pvc to job pods,match option --data
func (b *KFServingJobBuilder) Datas(volumes map[string]string) *KFServingJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (b *KFServingJobBuilder) DataDirs(volumes map[string]string) *KFServingJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// Port is used to set port,match the option --port
func (b *KFServingJobBuilder) Port(port int) *KFServingJobBuilder {
	if port > 0 {
		b.args.Port = port
	}
	return b
}

// ModeType is used to set mode type,match the option --mode-type
func (b *KFServingJobBuilder) ModelType(modeType string) *KFServingJobBuilder {
	if modeType != "" {
		b.args.ModelType = modeType
	}
	return b
}

// CanaryPercent is used to set Canary percent,match the option --canary-percent
func (b *KFServingJobBuilder) CanaryPercent(percent int) *KFServingJobBuilder {
	if percent > 0 {
		b.args.CanaryPercent = percent
	}
	return b
}

// StorageUri is used to set storage uri,match the option --storage-uri
func (b *KFServingJobBuilder) StorageUri(uri string) *KFServingJobBuilder {
	if uri != "" {
		b.args.StorageUri = uri
	}
	return b
}

// Build is used to build the job
func (b *KFServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.KFServingJob, b.args), nil
}
