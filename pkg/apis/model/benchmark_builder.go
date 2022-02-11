package model

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
	"strings"
)

type ModelBenchmarkArgsBuilder struct {
	args      *types.ModelBenchmarkArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewModelBenchmarkArgsBuilder() *ModelBenchmarkArgsBuilder {
	args := &types.ModelBenchmarkArgs{
		CommonModelArgs: types.CommonModelArgs{
			Image:     argsbuilder.DefaultModelJobImage,
			Namespace: "default",
		},
	}
	return &ModelBenchmarkArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewModelBenchmarkArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (m *ModelBenchmarkArgsBuilder) Name(name string) *ModelBenchmarkArgsBuilder {
	if name != "" {
		m.args.Name = name
	}
	return m
}

// Namespace is used to set job namespace,match option --namespace
func (m *ModelBenchmarkArgsBuilder) Namespace(namespace string) *ModelBenchmarkArgsBuilder {
	if namespace != "" {
		m.args.Namespace = namespace
	}
	return m
}

// Shell is used to specify linux shell type
func (m *ModelBenchmarkArgsBuilder) Shell(shell string) *ModelBenchmarkArgsBuilder {
	if shell != "" {
		m.args.Shell = shell
	}
	return m
}

// Command is used to set job command
func (m *ModelBenchmarkArgsBuilder) Command(args []string) *ModelBenchmarkArgsBuilder {
	m.args.Command = strings.Join(args, " ")
	return m
}

// Image is used to set job image,match the option --image
func (m *ModelBenchmarkArgsBuilder) Image(image string) *ModelBenchmarkArgsBuilder {
	if image != "" {
		m.args.Image = image
	}
	return m
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (m *ModelBenchmarkArgsBuilder) ImagePullPolicy(policy string) *ModelBenchmarkArgsBuilder {
	if policy != "" {
		m.args.ImagePullPolicy = policy
	}
	return m
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (m *ModelBenchmarkArgsBuilder) ImagePullSecrets(secrets []string) *ModelBenchmarkArgsBuilder {
	if secrets != nil {
		m.argValues["image-pull-secret"] = &secrets
	}
	return m
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (m *ModelBenchmarkArgsBuilder) GPUCount(count int) *ModelBenchmarkArgsBuilder {
	if count > 0 {
		m.args.GPUCount = count
	}
	return m
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (m *ModelBenchmarkArgsBuilder) GPUMemory(memory int) *ModelBenchmarkArgsBuilder {
	if memory > 0 {
		m.args.GPUMemory = memory
	}
	return m
}

// CPU assign cpu limits,match the option --cpu
func (m *ModelBenchmarkArgsBuilder) CPU(cpu string) *ModelBenchmarkArgsBuilder {
	if cpu != "" {
		m.args.Cpu = cpu
	}
	return m
}

// Memory assign memory limits,match option --memory
func (m *ModelBenchmarkArgsBuilder) Memory(memory string) *ModelBenchmarkArgsBuilder {
	if memory != "" {
		m.args.Memory = memory
	}
	return m
}

// Envs is used to set env of job containers,match option --env
func (m *ModelBenchmarkArgsBuilder) Envs(envs map[string]string) *ModelBenchmarkArgsBuilder {
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
func (m *ModelBenchmarkArgsBuilder) Tolerations(tolerations []string) *ModelBenchmarkArgsBuilder {
	m.argValues["toleration"] = &tolerations
	return m
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (m *ModelBenchmarkArgsBuilder) NodeSelectors(selectors map[string]string) *ModelBenchmarkArgsBuilder {
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
func (m *ModelBenchmarkArgsBuilder) Annotations(annotations map[string]string) *ModelBenchmarkArgsBuilder {
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
func (m *ModelBenchmarkArgsBuilder) Labels(labels map[string]string) *ModelBenchmarkArgsBuilder {
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
func (m *ModelBenchmarkArgsBuilder) Datas(volumes map[string]string) *ModelBenchmarkArgsBuilder {
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
func (m *ModelBenchmarkArgsBuilder) DataDirs(volumes map[string]string) *ModelBenchmarkArgsBuilder {
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
func (m *ModelBenchmarkArgsBuilder) ModelConfigFile(filePath string) *ModelBenchmarkArgsBuilder {
	if filePath != "" {
		m.args.ModelConfigFile = filePath
	}
	return m
}

// ModelName is used to set model name,match the option --model-name
func (m *ModelBenchmarkArgsBuilder) ModelName(name string) *ModelBenchmarkArgsBuilder {
	if name != "" {
		m.args.ModelName = name
	}
	return m
}

// ModelPath is used to set model path,match the option --model-path
func (m *ModelBenchmarkArgsBuilder) ModelPath(path string) *ModelBenchmarkArgsBuilder {
	if path != "" {
		m.args.ModelPath = path
	}
	return m
}

// Inputs is used to specify model inputs
func (m *ModelBenchmarkArgsBuilder) Inputs(inputs string) *ModelBenchmarkArgsBuilder {
	if inputs != "" {
		m.args.Inputs = inputs
	}
	return m
}

// Outputs is used to specify model outputs
func (m *ModelBenchmarkArgsBuilder) Outputs(outputs string) *ModelBenchmarkArgsBuilder {
	if outputs != "" {
		m.args.Outputs = outputs
	}
	return m
}

// Concurrency is used to specify number of concurrently to run
func (m *ModelBenchmarkArgsBuilder) Concurrency(concurrency int) *ModelBenchmarkArgsBuilder {
	if concurrency > 0 {
		m.args.Concurrency = concurrency
	}
	return m
}

// Requests is used to specify number of requests to run
func (m *ModelBenchmarkArgsBuilder) Requests(requests int) *ModelBenchmarkArgsBuilder {
	if requests > 0 {
		m.args.Requests = requests
	}
	return m
}

// Duration is used to specify benchmark duration
func (m *ModelBenchmarkArgsBuilder) Duration(duration int) *ModelBenchmarkArgsBuilder {
	if duration > 0 {
		m.args.Duration = duration
	}
	return m
}

// ReportPath is used to specify benchmark result saved path
func (m *ModelBenchmarkArgsBuilder) ReportPath(reportPath string) *ModelBenchmarkArgsBuilder {
	if reportPath != "" {
		m.args.ReportPath = reportPath
	}
	return m
}

// Build is used to build the job
func (m *ModelBenchmarkArgsBuilder) Build() (*Job, error) {
	for key, value := range m.argValues {
		m.AddArgValue(key, value)
	}
	if err := m.PreBuild(); err != nil {
		return nil, err
	}
	if err := m.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(m.args.Name, types.ModelBenchmarkJob, m.args), nil
}
