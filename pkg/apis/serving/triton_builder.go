package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type TritonServingJobBuilder struct {
	args      *types.TritonServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewTritonServingJobBuilder() *TritonServingJobBuilder {
	args := &types.TritonServingArgs{
		HttpPort:    8000,
		GrpcPort:    8001,
		MetricsPort: 8002,
		CommonServingArgs: types.CommonServingArgs{
			Image:           argsbuilder.DefaultTritonServingImage,
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Namespace:       "default",
			Shell:           "sh",
		},
	}
	return &TritonServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewTritonServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *TritonServingJobBuilder) Name(name string) *TritonServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *TritonServingJobBuilder) Namespace(namespace string) *TritonServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *TritonServingJobBuilder) Shell(shell string) *TritonServingJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *TritonServingJobBuilder) Command(args []string) *TritonServingJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *TritonServingJobBuilder) GPUCount(count int) *TritonServingJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *TritonServingJobBuilder) GPUMemory(memory int) *TritonServingJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// GPUCore is used to set gpu core for the job,match the option --gpucore
func (b *TritonServingJobBuilder) GPUCore(core int) *TritonServingJobBuilder {
	if core > 0 {
		b.args.GPUCore = core
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *TritonServingJobBuilder) Image(image string) *TritonServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *TritonServingJobBuilder) ImagePullPolicy(policy string) *TritonServingJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *TritonServingJobBuilder) CPU(cpu string) *TritonServingJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *TritonServingJobBuilder) Memory(memory string) *TritonServingJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *TritonServingJobBuilder) Envs(envs map[string]string) *TritonServingJobBuilder {
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
func (b *TritonServingJobBuilder) Replicas(count int) *TritonServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *TritonServingJobBuilder) EnableIstio() *TritonServingJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *TritonServingJobBuilder) ExposeService() *TritonServingJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *TritonServingJobBuilder) Version(version string) *TritonServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *TritonServingJobBuilder) Tolerations(tolerations []string) *TritonServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *TritonServingJobBuilder) NodeSelectors(selectors map[string]string) *TritonServingJobBuilder {
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
func (b *TritonServingJobBuilder) Annotations(annotations map[string]string) *TritonServingJobBuilder {
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
func (b *TritonServingJobBuilder) Labels(labels map[string]string) *TritonServingJobBuilder {
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
func (b *TritonServingJobBuilder) Datas(volumes map[string]string) *TritonServingJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

// DataSubPathExprs is used to mount k8s pvc subpath to job pods,match option data-subpath-expr
func (b *TritonServingJobBuilder) DataSubPathExprs(exprs map[string]string) *TritonServingJobBuilder {
	if exprs != nil && len(exprs) != 0 {
		s := []string{}
		for key, value := range exprs {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-subpath-expr"] = &s
	}
	return b
}

// TempDirs specify the deployment empty dir
func (b *TritonServingJobBuilder) TempDirs(volumes map[string]string) *TritonServingJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["temp-dir"] = &s
	}
	return b
}

// EmptyDirSubPathExprs specify the datasource subpath to mount to the pod by expression
func (b *TritonServingJobBuilder) EmptyDirSubPathExprs(exprs map[string]string) *TritonServingJobBuilder {
	if exprs != nil && len(exprs) != 0 {
		s := []string{}
		for key, value := range exprs {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["temp-dir-subpath-expr"] = &s
	}
	return b
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (b *TritonServingJobBuilder) DataDirs(volumes map[string]string) *TritonServingJobBuilder {
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
func (b *TritonServingJobBuilder) HttpPort(port int) *TritonServingJobBuilder {
	if port > 0 {
		b.args.HttpPort = port
	}
	return b
}

// RestfulPort is used to set restful port,match the option --restful-port
func (b *TritonServingJobBuilder) GrpcPort(port int) *TritonServingJobBuilder {
	if port > 0 {
		b.args.GrpcPort = port
	}
	return b
}

// MetricsPort is used to set metrics port,match the option --metric-port
func (b *TritonServingJobBuilder) MetricsPort(port int) *TritonServingJobBuilder {
	if port > 0 {
		b.args.MetricsPort = port
	}
	return b
}

// ModelRepository is used to set model store,match the option --model-repository
func (b *TritonServingJobBuilder) ModelRepository(repository string) *TritonServingJobBuilder {
	if repository != "" {
		b.args.ModelRepository = repository
	}
	return b
}

// AllowMetrics is enable metric,match the option --allow-metrics
func (b *TritonServingJobBuilder) AllowMetrics() *TritonServingJobBuilder {
	b.args.AllowMetrics = true
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *TritonServingJobBuilder) ConfigFiles(files map[string]string) *TritonServingJobBuilder {
	if files != nil && len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// Build is used to build the job
func (b *TritonServingJobBuilder) Build() (*Job, error) {
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
