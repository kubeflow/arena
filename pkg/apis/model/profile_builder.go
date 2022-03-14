package model

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
	"strings"
)

type ModelProfileJobBuilder struct {
	args      *types.ModelProfileArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewModelProfileJobBuilder() *ModelProfileJobBuilder {
	args := &types.ModelProfileArgs{
		CommonModelArgs: types.CommonModelArgs{
			Image:     argsbuilder.DefaultModelJobImage,
			Namespace: "default",
		},
	}
	return &ModelProfileJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewModelProfileArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (m *ModelProfileJobBuilder) Name(name string) *ModelProfileJobBuilder {
	if name != "" {
		m.args.Name = name
	}
	return m
}

// Namespace is used to set job namespace,match option --namespace
func (m *ModelProfileJobBuilder) Namespace(namespace string) *ModelProfileJobBuilder {
	if namespace != "" {
		m.args.Namespace = namespace
	}
	return m
}

// Shell is used to specify linux shell type
func (m *ModelProfileJobBuilder) Shell(shell string) *ModelProfileJobBuilder {
	if shell != "" {
		m.args.Shell = shell
	}
	return m
}

// Command is used to set job command
func (m *ModelProfileJobBuilder) Command(args []string) *ModelProfileJobBuilder {
	m.args.Command = strings.Join(args, " ")
	return m
}

// Image is used to set job image,match the option --image
func (m *ModelProfileJobBuilder) Image(image string) *ModelProfileJobBuilder {
	if image != "" {
		m.args.Image = image
	}
	return m
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (m *ModelProfileJobBuilder) ImagePullPolicy(policy string) *ModelProfileJobBuilder {
	if policy != "" {
		m.args.ImagePullPolicy = policy
	}
	return m
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (m *ModelProfileJobBuilder) ImagePullSecrets(secrets []string) *ModelProfileJobBuilder {
	if secrets != nil {
		m.argValues["image-pull-secret"] = &secrets
	}
	return m
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (m *ModelProfileJobBuilder) GPUCount(count int) *ModelProfileJobBuilder {
	if count > 0 {
		m.args.GPUCount = count
	}
	return m
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (m *ModelProfileJobBuilder) GPUMemory(memory int) *ModelProfileJobBuilder {
	if memory > 0 {
		m.args.GPUMemory = memory
	}
	return m
}

// CPU assign cpu limits,match the option --cpu
func (m *ModelProfileJobBuilder) CPU(cpu string) *ModelProfileJobBuilder {
	if cpu != "" {
		m.args.Cpu = cpu
	}
	return m
}

// Memory assign memory limits,match option --memory
func (m *ModelProfileJobBuilder) Memory(memory string) *ModelProfileJobBuilder {
	if memory != "" {
		m.args.Memory = memory
	}
	return m
}

// Envs is used to set env of job containers,match option --env
func (m *ModelProfileJobBuilder) Envs(envs map[string]string) *ModelProfileJobBuilder {
	if envs != nil && len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["env"] = &envSlice
	}
	return m
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (m *ModelProfileJobBuilder) Tolerations(tolerations []string) *ModelProfileJobBuilder {
	m.argValues["toleration"] = &tolerations
	return m
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (m *ModelProfileJobBuilder) NodeSelectors(selectors map[string]string) *ModelProfileJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["selector"] = &selectorsSlice
	}
	return m
}

// Annotations is used to add annotations for job pods,match option --annotation
func (m *ModelProfileJobBuilder) Annotations(annotations map[string]string) *ModelProfileJobBuilder {
	if annotations != nil && len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["annotation"] = &s
	}
	return m
}

// Labels is used to add labels for job
func (m *ModelProfileJobBuilder) Labels(labels map[string]string) *ModelProfileJobBuilder {
	if labels != nil && len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		m.argValues["label"] = &s
	}
	return m
}

// Datas is used to mount k8s pvc to job pods,match option --data
func (m *ModelProfileJobBuilder) Datas(volumes map[string]string) *ModelProfileJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		m.argValues["data"] = &s
	}
	return m
}

// DataDirs is used to mount host files to job containers,match option --data-dir
func (m *ModelProfileJobBuilder) DataDirs(volumes map[string]string) *ModelProfileJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		m.argValues["data-dir"] = &s
	}
	return m
}

// ModelConfigFile is used to set model config file,match the option --model-config-file
func (m *ModelProfileJobBuilder) ModelConfigFile(filePath string) *ModelProfileJobBuilder {
	if filePath != "" {
		m.args.ModelConfigFile = filePath
	}
	return m
}

// ModelName is used to set model name,match the option --model-name
func (m *ModelProfileJobBuilder) ModelName(name string) *ModelProfileJobBuilder {
	if name != "" {
		m.args.ModelName = name
	}
	return m
}

// ModelPath is used to set model path,match the option --model-path
func (m *ModelProfileJobBuilder) ModelPath(path string) *ModelProfileJobBuilder {
	if path != "" {
		m.args.ModelPath = path
	}
	return m
}

// Inputs is used to specify model inputs
func (m *ModelProfileJobBuilder) Inputs(inputs string) *ModelProfileJobBuilder {
	if inputs != "" {
		m.args.Inputs = inputs
	}
	return m
}

// Outputs is used to specify model outputs
func (m *ModelProfileJobBuilder) Outputs(outputs string) *ModelProfileJobBuilder {
	if outputs != "" {
		m.args.Outputs = outputs
	}
	return m
}

// ReportPath is used to specify profile result path
func (m *ModelProfileJobBuilder) ReportPath(reportPath string) *ModelProfileJobBuilder {
	if reportPath != "" {
		m.args.ReportPath = reportPath
	}
	return m
}

// Build is used to build the job
func (m *ModelProfileJobBuilder) Build() (*Job, error) {
	for key, value := range m.argValues {
		m.AddArgValue(key, value)
	}
	if err := m.PreBuild(); err != nil {
		return nil, err
	}
	if err := m.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(m.args.Name, types.ModelProfileJob, m.args), nil
}
