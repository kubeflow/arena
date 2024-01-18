package model

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
	"strings"
)

type ModelEvaluateJobBuilder struct {
	args      *types.ModelEvaluateArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewModelEvaluateJobBuilder() *ModelEvaluateJobBuilder {
	args := &types.ModelEvaluateArgs{
		CommonModelArgs: types.CommonModelArgs{
			Image:     argsbuilder.DefaultModelJobImage,
			Namespace: "default",
		},
	}
	return &ModelEvaluateJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewModelEvaluateArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (m *ModelEvaluateJobBuilder) Name(name string) *ModelEvaluateJobBuilder {
	if name != "" {
		m.args.Name = name
	}
	return m
}

// Namespace is used to set job namespace,match option --namespace
func (m *ModelEvaluateJobBuilder) Namespace(namespace string) *ModelEvaluateJobBuilder {
	if namespace != "" {
		m.args.Namespace = namespace
	}
	return m
}

// Shell is used to specify linux shell type
func (m *ModelEvaluateJobBuilder) Shell(shell string) *ModelEvaluateJobBuilder {
	if shell != "" {
		m.args.Shell = shell
	}
	return m
}

// Command is used to set job command
func (m *ModelEvaluateJobBuilder) Command(args []string) *ModelEvaluateJobBuilder {
	m.args.Command = strings.Join(args, " ")
	return m
}

// Image is used to set job image,match the option --image
func (m *ModelEvaluateJobBuilder) Image(image string) *ModelEvaluateJobBuilder {
	if image != "" {
		m.args.Image = image
	}
	return m
}

// ImagePullPolicy is used to set image pull policy,match the option --image-pull-policy
func (m *ModelEvaluateJobBuilder) ImagePullPolicy(policy string) *ModelEvaluateJobBuilder {
	if policy != "" {
		m.args.ImagePullPolicy = policy
	}
	return m
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (m *ModelEvaluateJobBuilder) ImagePullSecrets(secrets []string) *ModelEvaluateJobBuilder {
	if secrets != nil {
		m.argValues["image-pull-secret"] = &secrets
	}
	return m
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (m *ModelEvaluateJobBuilder) GPUCount(count int) *ModelEvaluateJobBuilder {
	if count > 0 {
		m.args.GPUCount = count
	}
	return m
}

// GPUMemory is used to set gpu memory for the job,match the option --gpumemory
func (m *ModelEvaluateJobBuilder) GPUMemory(memory int) *ModelEvaluateJobBuilder {
	if memory > 0 {
		m.args.GPUMemory = memory
	}
	return m
}

// GPUCore is used to set gpu core for the job,match the option --gpumemory
func (m *ModelEvaluateJobBuilder) GPUCore(core int) *ModelEvaluateJobBuilder {
	if core > 0 {
		m.args.GPUCore = core
	}
	return m
}

// CPU assign cpu limits,match the option --cpu
func (m *ModelEvaluateJobBuilder) CPU(cpu string) *ModelEvaluateJobBuilder {
	if cpu != "" {
		m.args.Cpu = cpu
	}
	return m
}

// Memory assign memory limits,match option --memory
func (m *ModelEvaluateJobBuilder) Memory(memory string) *ModelEvaluateJobBuilder {
	if memory != "" {
		m.args.Memory = memory
	}
	return m
}

// Envs is used to set env of job containers,match option --env
func (m *ModelEvaluateJobBuilder) Envs(envs map[string]string) *ModelEvaluateJobBuilder {
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
func (m *ModelEvaluateJobBuilder) Tolerations(tolerations []string) *ModelEvaluateJobBuilder {
	m.argValues["toleration"] = &tolerations
	return m
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (m *ModelEvaluateJobBuilder) NodeSelectors(selectors map[string]string) *ModelEvaluateJobBuilder {
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
func (m *ModelEvaluateJobBuilder) Annotations(annotations map[string]string) *ModelEvaluateJobBuilder {
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
func (m *ModelEvaluateJobBuilder) Labels(labels map[string]string) *ModelEvaluateJobBuilder {
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
func (m *ModelEvaluateJobBuilder) Datas(volumes map[string]string) *ModelEvaluateJobBuilder {
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
func (m *ModelEvaluateJobBuilder) DataDirs(volumes map[string]string) *ModelEvaluateJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		m.argValues["data-dir"] = &s
	}
	return m
}

// SyncImage is used to set syncing image,match option --sync-image
func (m *ModelEvaluateJobBuilder) SyncImage(image string) *ModelEvaluateJobBuilder {
	if image != "" {
		m.args.SyncImage = image
	}
	return m
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (m *ModelEvaluateJobBuilder) SyncMode(mode string) *ModelEvaluateJobBuilder {
	if mode != "" {
		m.args.SyncMode = mode
	}
	return m
}

// SyncSource is used to set syncing source,match option --sync-source
func (m *ModelEvaluateJobBuilder) SyncSource(source string) *ModelEvaluateJobBuilder {
	if source != "" {
		m.args.SyncSource = source
	}
	return m
}

// ModelPath is used to set model path,match the option --model-path
func (m *ModelEvaluateJobBuilder) ModelPath(path string) *ModelEvaluateJobBuilder {
	if path != "" {
		m.args.ModelPath = path
	}
	return m
}

// ModelPlatform specify the model platform, such as torchscript/tensorflow
func (m *ModelEvaluateJobBuilder) ModelPlatform(modelPlatform string) *ModelEvaluateJobBuilder {
	if modelPlatform != "" {
		m.args.ModelPlatform = modelPlatform
	}
	return m
}

// DatasetPath is the dataset to evaluate model
func (m *ModelEvaluateJobBuilder) DatasetPath(datasetPath string) *ModelEvaluateJobBuilder {
	if datasetPath != "" {
		m.args.DatasetPath = datasetPath
	}
	return m
}

// BatchSize is the batch size of evaluate
func (m *ModelEvaluateJobBuilder) BatchSize(batchSize int) *ModelEvaluateJobBuilder {
	if batchSize > 0 {
		m.args.BatchSize = batchSize
	}
	return m
}

// ReportPath is used to specify evaluate result path
func (m *ModelEvaluateJobBuilder) ReportPath(reportPath string) *ModelEvaluateJobBuilder {
	if reportPath != "" {
		m.args.ReportPath = reportPath
	}
	return m
}

// Build is used to build the job
func (m *ModelEvaluateJobBuilder) Build() (*Job, error) {
	for key, value := range m.argValues {
		m.AddArgValue(key, value)
	}
	if err := m.PreBuild(); err != nil {
		return nil, err
	}
	if err := m.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(m.args.Name, types.ModelEvaluateJob, m.args), nil
}
