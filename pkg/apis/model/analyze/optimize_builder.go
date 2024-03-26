package analyze

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type ModelOptimizeJobBuilder struct {
	args      *types.ModelOptimizeArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewModelOptimizeJobBuilder() *ModelOptimizeJobBuilder {
	args := &types.ModelOptimizeArgs{
		CommonModelArgs: types.CommonModelArgs{
			Image:     argsbuilder.DefaultModelJobImage,
			Namespace: "default",
		},
	}
	return &ModelOptimizeJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewModelOptimizeArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (m *ModelOptimizeJobBuilder) Name(name string) *ModelOptimizeJobBuilder {
	if name != "" {
		m.args.Name = name
	}
	return m
}

// Namespace is used to set job namespace,match option --namespace
func (m *ModelOptimizeJobBuilder) Namespace(namespace string) *ModelOptimizeJobBuilder {
	if namespace != "" {
		m.args.Namespace = namespace
	}
	return m
}

// Shell is used to specify linux shell type
func (m *ModelOptimizeJobBuilder) Shell(shell string) *ModelOptimizeJobBuilder {
	if shell != "" {
		m.args.Shell = shell
	}
	return m
}

// Command is used to set job command
func (m *ModelOptimizeJobBuilder) Command(args []string) *ModelOptimizeJobBuilder {
	m.args.Command = strings.Join(args, " ")
	return m
}

// Image is used to set job image,match the option --image
func (m *ModelOptimizeJobBuilder) Image(image string) *ModelOptimizeJobBuilder {
	if image != "" {
		m.args.Image = image
	}
	return m
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (m *ModelOptimizeJobBuilder) ImagePullPolicy(policy string) *ModelOptimizeJobBuilder {
	if policy != "" {
		m.args.ImagePullPolicy = policy
	}
	return m
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (m *ModelOptimizeJobBuilder) ImagePullSecrets(secrets []string) *ModelOptimizeJobBuilder {
	if secrets != nil {
		m.argValues["image-pull-secret"] = &secrets
	}
	return m
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (m *ModelOptimizeJobBuilder) GPUCount(count int) *ModelOptimizeJobBuilder {
	if count > 0 {
		m.args.GPUCount = count
	}
	return m
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (m *ModelOptimizeJobBuilder) GPUMemory(memory int) *ModelOptimizeJobBuilder {
	if memory > 0 {
		m.args.GPUMemory = memory
	}
	return m
}

// GPUCore is used to set gpu core for the job, match the option --gpucore
func (b *ModelOptimizeJobBuilder) GPUCore(core int) *ModelOptimizeJobBuilder {
	if core > 0 {
		b.args.GPUCore = core
	}
	return b
}

// CPU assign cpu limits,match the option --cpu
func (m *ModelOptimizeJobBuilder) CPU(cpu string) *ModelOptimizeJobBuilder {
	if cpu != "" {
		m.args.Cpu = cpu
	}
	return m
}

// Memory assign memory limits,match option --memory
func (m *ModelOptimizeJobBuilder) Memory(memory string) *ModelOptimizeJobBuilder {
	if memory != "" {
		m.args.Memory = memory
	}
	return m
}

// Envs is used to set env of job containers,match option --env
func (m *ModelOptimizeJobBuilder) Envs(envs map[string]string) *ModelOptimizeJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["env"] = &envSlice
	}
	return m
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (m *ModelOptimizeJobBuilder) Tolerations(tolerations []string) *ModelOptimizeJobBuilder {
	m.argValues["toleration"] = &tolerations
	return m
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (m *ModelOptimizeJobBuilder) NodeSelectors(selectors map[string]string) *ModelOptimizeJobBuilder {
	if len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["selector"] = &selectorsSlice
	}
	return m
}

// Annotations is used to add annotations for job pods,match option --annotation
func (m *ModelOptimizeJobBuilder) Annotations(annotations map[string]string) *ModelOptimizeJobBuilder {
	if len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["annotation"] = &s
	}
	return m
}

// Labels is used to add labels for job
func (m *ModelOptimizeJobBuilder) Labels(labels map[string]string) *ModelOptimizeJobBuilder {
	if len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["label"] = &s
	}
	return m
}

// Datas is used to mount k8s pvc to job pods,match option --data
func (m *ModelOptimizeJobBuilder) Datas(volumes map[string]string) *ModelOptimizeJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		m.argValues["data"] = &s
	}
	return m
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (m *ModelOptimizeJobBuilder) DataDirs(volumes map[string]string) *ModelOptimizeJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		m.argValues["data-dir"] = &s
	}
	return m
}

// ModelConfigFile is used to set model config file,match the option --model-config-file
func (m *ModelOptimizeJobBuilder) ModelConfigFile(filePath string) *ModelOptimizeJobBuilder {
	if filePath != "" {
		m.args.ModelConfigFile = filePath
	}
	return m
}

// ModelName is used to set model name,match the option --model-name
func (m *ModelOptimizeJobBuilder) ModelName(name string) *ModelOptimizeJobBuilder {
	if name != "" {
		m.args.ModelName = name
	}
	return m
}

// ModelPath is used to set model path,match the option --model-path
func (m *ModelOptimizeJobBuilder) ModelPath(path string) *ModelOptimizeJobBuilder {
	if path != "" {
		m.args.ModelPath = path
	}
	return m
}

// Inputs is used to specify model inputs
func (m *ModelOptimizeJobBuilder) Inputs(inputs string) *ModelOptimizeJobBuilder {
	if inputs != "" {
		m.args.Inputs = inputs
	}
	return m
}

// Outputs is used to specify model outputs
func (m *ModelOptimizeJobBuilder) Outputs(outputs string) *ModelOptimizeJobBuilder {
	if outputs != "" {
		m.args.Outputs = outputs
	}
	return m
}

// Optimizer is used to specify optimized model save path
func (m *ModelOptimizeJobBuilder) Optimizer(optimizer string) *ModelOptimizeJobBuilder {
	if optimizer != "" {
		m.args.Optimizer = optimizer
	}
	return m
}

// TargetDevice is used to specify model deploy device
func (m *ModelOptimizeJobBuilder) TargetDevice(targetDevice string) *ModelOptimizeJobBuilder {
	if targetDevice != "" {
		m.args.TargetDevice = targetDevice
	}
	return m
}

// ExportPath is used to specify optimized model save path
func (m *ModelOptimizeJobBuilder) ExportPath(exportPath string) *ModelOptimizeJobBuilder {
	if exportPath != "" {
		m.args.ExportPath = exportPath
	}
	return m
}

// Build is used to build the job
func (m *ModelOptimizeJobBuilder) Build() (*Job, error) {
	for key, value := range m.argValues {
		m.AddArgValue(key, value)
	}
	if err := m.PreBuild(); err != nil {
		return nil, err
	}
	if err := m.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(m.args.Name, types.ModelOptimizeJob, m.args), nil
}
