package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type CustomServingJobBuilder struct {
	args      *types.CustomServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewCustomServingJobBuilder() *CustomServingJobBuilder {
	args := &types.CustomServingArgs{
		CommonServingArgs: types.CommonServingArgs{
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Shell:           "sh",
		},
	}
	return &CustomServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewCustomServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *CustomServingJobBuilder) Name(name string) *CustomServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *CustomServingJobBuilder) Namespace(namespace string) *CustomServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *CustomServingJobBuilder) Shell(shell string) *CustomServingJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *CustomServingJobBuilder) Command(args []string) *CustomServingJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *CustomServingJobBuilder) GPUCount(count int) *CustomServingJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *CustomServingJobBuilder) GPUMemory(memory int) *CustomServingJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *CustomServingJobBuilder) Image(image string) *CustomServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *CustomServingJobBuilder) ImagePullPolicy(policy string) *CustomServingJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *CustomServingJobBuilder) CPU(cpu string) *CustomServingJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *CustomServingJobBuilder) Memory(memory string) *CustomServingJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *CustomServingJobBuilder) Envs(envs map[string]string) *CustomServingJobBuilder {
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
func (b *CustomServingJobBuilder) Replicas(count int) *CustomServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *CustomServingJobBuilder) EnableIstio() *CustomServingJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *CustomServingJobBuilder) ExposeService() *CustomServingJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *CustomServingJobBuilder) Version(version string) *CustomServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *CustomServingJobBuilder) Tolerations(tolerations []string) *CustomServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *CustomServingJobBuilder) NodeSelectors(selectors map[string]string) *CustomServingJobBuilder {
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
func (b *CustomServingJobBuilder) Annotations(annotations map[string]string) *CustomServingJobBuilder {
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
func (b *CustomServingJobBuilder) Labels(labels map[string]string) *CustomServingJobBuilder {
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
func (b *CustomServingJobBuilder) Datas(volumes map[string]string) *CustomServingJobBuilder {
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
func (b *CustomServingJobBuilder) DataDirs(volumes map[string]string) *CustomServingJobBuilder {
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
func (b *CustomServingJobBuilder) Port(port int) *CustomServingJobBuilder {
	if port > 0 {
		b.args.Port = port
	}
	return b
}

// RestfulPort is used to set restful port,match the option --restful-port
func (b *CustomServingJobBuilder) RestfulPort(port int) *CustomServingJobBuilder {
	if port > 0 {
		b.args.RestfulPort = port
	}
	return b
}

// MetricsPort is used to set metrics port,match the option --metrics-port
func (b *CustomServingJobBuilder) MetricsPort(port int) *CustomServingJobBuilder {
	if port > 0 {
		b.args.MetricsPort = port
	}
	return b
}

// Build is used to build the job
func (b *CustomServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.CustomServingJob, b.args), nil
}
