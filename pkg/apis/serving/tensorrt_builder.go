package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type TRTServingJobBuilder struct {
	args      *types.TensorRTServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewTRTServingJobBuilder() *TRTServingJobBuilder {
	args := &types.TensorRTServingArgs{
		HttpPort:    8000,
		GrpcPort:    8001,
		MetricsPort: 8002,
		CommonServingArgs: types.CommonServingArgs{
			Image:           argsbuilder.DefaultTRTServingImage,
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Namespace:       "default",
		},
	}
	return &TRTServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewTensorRTServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *TRTServingJobBuilder) Name(name string) *TRTServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *TRTServingJobBuilder) Namespace(namespace string) *TRTServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Command is used to set job command
func (b *TRTServingJobBuilder) Command(args []string) *TRTServingJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *TRTServingJobBuilder) GPUCount(count int) *TRTServingJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *TRTServingJobBuilder) GPUMemory(memory int) *TRTServingJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *TRTServingJobBuilder) Image(image string) *TRTServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *TRTServingJobBuilder) ImagePullPolicy(policy string) *TRTServingJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *TRTServingJobBuilder) CPU(cpu string) *TRTServingJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *TRTServingJobBuilder) Memory(memory string) *TRTServingJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *TRTServingJobBuilder) Envs(envs map[string]string) *TRTServingJobBuilder {
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
func (b *TRTServingJobBuilder) Replicas(count int) *TRTServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *TRTServingJobBuilder) EnableIstio() *TRTServingJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *TRTServingJobBuilder) ExposeService() *TRTServingJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *TRTServingJobBuilder) Version(version string) *TRTServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *TRTServingJobBuilder) Tolerations(tolerations []string) *TRTServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *TRTServingJobBuilder) NodeSelectors(selectors map[string]string) *TRTServingJobBuilder {
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
func (b *TRTServingJobBuilder) Annotations(annotations map[string]string) *TRTServingJobBuilder {
	if annotations != nil && len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

// Datas is used to mount k8s pvc to job pods,match option --data
func (b *TRTServingJobBuilder) Datas(volumes map[string]string) *TRTServingJobBuilder {
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
func (b *TRTServingJobBuilder) DataDirs(volumes map[string]string) *TRTServingJobBuilder {
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
func (b *TRTServingJobBuilder) HttpPort(port int) *TRTServingJobBuilder {
	if port > 0 {
		b.args.HttpPort = port
	}
	return b
}

// RestfulPort is used to set restful port,match the option --restful-port
func (b *TRTServingJobBuilder) GrpcPort(port int) *TRTServingJobBuilder {
	if port > 0 {
		b.args.GrpcPort = port
	}
	return b
}

// MetricsPort is used to set metrics port,match the option --metric-port
func (b *TRTServingJobBuilder) MetricsPort(port int) *TRTServingJobBuilder {
	if port > 0 {
		b.args.MetricsPort = port
	}
	return b
}

// ModelStore is used to set model store,match the option --model-store
func (b *TRTServingJobBuilder) ModelStore(store string) *TRTServingJobBuilder {
	if store != "" {
		b.args.ModelStore = store
	}
	return b
}

// AllowMetrics is enable metric,match the option --allow-meetrics
func (b *TRTServingJobBuilder) AllowMetrics() *TRTServingJobBuilder {
	b.args.AllowMetrics = true
	return b
}

// Build is used to build the job
func (b *TRTServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.TRTServingJob, b.args), nil
}
