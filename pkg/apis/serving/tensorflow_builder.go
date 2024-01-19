package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type TFServingJobBuilder struct {
	args      *types.TensorFlowServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewTFServingJobBuilder() *TFServingJobBuilder {
	args := &types.TensorFlowServingArgs{
		Port:        8500,
		RestfulPort: 8501,
		CommonServingArgs: types.CommonServingArgs{
			Image:           argsbuilder.DefaultTfServingImage,
			ImagePullPolicy: "IfNotPresent",
			Replicas:        1,
			Namespace:       "default",
			Shell:           "sh",
		},
	}
	return &TFServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewTensorflowServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *TFServingJobBuilder) Name(name string) *TFServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *TFServingJobBuilder) Namespace(namespace string) *TFServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Shell is used to set bash or sh
func (b *TFServingJobBuilder) Shell(shell string) *TFServingJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *TFServingJobBuilder) Command(args []string) *TFServingJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *TFServingJobBuilder) GPUCount(count int) *TFServingJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (b *TFServingJobBuilder) GPUMemory(memory int) *TFServingJobBuilder {
	if memory > 0 {
		b.args.GPUMemory = memory
	}
	return b
}

// GPUCore is used to set gpu core for the job, match the option --gpucore
func (b *TFServingJobBuilder) GPUCore(core int) *TFServingJobBuilder {
	if core > 0 {
		b.args.GPUCore = core
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *TFServingJobBuilder) Image(image string) *TFServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (b *TFServingJobBuilder) ImagePullPolicy(policy string) *TFServingJobBuilder {
	if policy != "" {
		b.args.ImagePullPolicy = policy
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (b *TFServingJobBuilder) CPU(cpu string) *TFServingJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *TFServingJobBuilder) Memory(memory string) *TFServingJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *TFServingJobBuilder) Envs(envs map[string]string) *TFServingJobBuilder {
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
func (b *TFServingJobBuilder) Replicas(count int) *TFServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// Port is used to set port,match the option --port
func (b *TFServingJobBuilder) Port(port int) *TFServingJobBuilder {
	if port > 0 {
		b.args.Port = port
	}
	return b
}

// RestfulPort is used to set restful port,match the option --restful-port
func (b *TFServingJobBuilder) RestfulPort(port int) *TFServingJobBuilder {
	if port > 0 {
		b.args.RestfulPort = port
	}
	return b
}

// EnableIstio is used to enable istio,match the option --enable-istio
func (b *TFServingJobBuilder) EnableIstio() *TFServingJobBuilder {
	b.args.EnableIstio = true
	return b
}

// ExposeService is used to expose service,match the option --expose-service
func (b *TFServingJobBuilder) ExposeService() *TFServingJobBuilder {
	b.args.ExposeService = true
	return b
}

// Version is used to set serving job version,match the option --version
func (b *TFServingJobBuilder) Version(version string) *TFServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *TFServingJobBuilder) Tolerations(tolerations []string) *TFServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *TFServingJobBuilder) NodeSelectors(selectors map[string]string) *TFServingJobBuilder {
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
func (b *TFServingJobBuilder) Annotations(annotations map[string]string) *TFServingJobBuilder {
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
func (b *TFServingJobBuilder) Labels(labels map[string]string) *TFServingJobBuilder {
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
func (b *TFServingJobBuilder) Datas(volumes map[string]string) *TFServingJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

// DataSubPathExprs is used to mount k8s pvc subpath to job pods,match option data-subpath-expr
func (b *TFServingJobBuilder) DataSubPathExprs(exprs map[string]string) *TFServingJobBuilder {
	if len(exprs) != 0 {
		s := []string{}
		for key, value := range exprs {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-subpath-expr"] = &s
	}
	return b
}

// TempDirs specify the deployment empty dir
func (b *TFServingJobBuilder) TempDirs(volumes map[string]string) *TFServingJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["temp-dir"] = &s
	}
	return b
}

// EmptyDirSubPathExprs specify the datasource subpath to mount to the pod by expression
func (b *TFServingJobBuilder) EmptyDirSubPathExprs(exprs map[string]string) *TFServingJobBuilder {
	if len(exprs) != 0 {
		s := []string{}
		for key, value := range exprs {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["temp-dir-subpath-expr"] = &s
	}
	return b
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (b *TFServingJobBuilder) DataDirs(volumes map[string]string) *TFServingJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// VersionPolicy is used to set version policy,match the option --version-policy
func (b *TFServingJobBuilder) VersionPolicy(policy string) *TFServingJobBuilder {
	if policy != "" {
		b.args.VersionPolicy = policy
	}
	return b
}

// ModelConfigFile is used to set model config file,match the option --model-config-file
func (b *TFServingJobBuilder) ModelConfigFile(filePath string) *TFServingJobBuilder {
	if filePath != "" {
		b.args.ModelConfigFile = filePath
	}
	return b
}

// MonitoringConfigFile is used to set monitoring config file,match the option --monitoring-config-file
func (b *TFServingJobBuilder) MonitoringConfigFile(filePath string) *TFServingJobBuilder {
	if filePath != "" {
		b.args.MonitoringConfigFile = filePath
	}
	return b
}

// ModelName is used to set model name,match the option --model-name
func (b *TFServingJobBuilder) ModelName(name string) *TFServingJobBuilder {
	if name != "" {
		b.args.ModelName = name
	}
	return b
}

// ModelPath is used to set model path,match the option --model-path
func (b *TFServingJobBuilder) ModelPath(path string) *TFServingJobBuilder {
	if path != "" {
		b.args.ModelPath = path
	}
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *TFServingJobBuilder) ConfigFiles(files map[string]string) *TFServingJobBuilder {
	if len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// Build is used to build the job
func (b *TFServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.TFServingJob, b.args), nil
}
