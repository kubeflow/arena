package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type TFJobBuilder struct {
	args      *types.SubmitTFJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewTFJobBuilder(args *types.SubmitTFJobArgs) *TFJobBuilder {
	if args == nil {
		args = &types.SubmitTFJobArgs{
			CleanPodPolicy:        "Running",
			CommonSubmitArgs:      DefaultCommonSubmitArgs,
			SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
		}
	}

	return &TFJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitTFJobArgsBuilder(args),
	}
}

func (b *TFJobBuilder) GetArgValues() map[string]interface{} {
	return b.argValues
}

func (b *TFJobBuilder) Name(name string) *TFJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

func (b *TFJobBuilder) Command(args []string) *TFJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

func (b *TFJobBuilder) WorkingDir(dir string) *TFJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

func (b *TFJobBuilder) Envs(envs map[string]string) *TFJobBuilder {
	if envs != nil && len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

func (b *TFJobBuilder) GPUCount(count int) *TFJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

func (b *TFJobBuilder) Image(image string) *TFJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

func (b *TFJobBuilder) Tolerations(tolerations []string) *TFJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

func (b *TFJobBuilder) ConfigFiles(files map[string]string) *TFJobBuilder {
	if files != nil && len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

func (b *TFJobBuilder) NodeSelectors(selectors map[string]string) *TFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["selector"] = &selectorsSlice
	}
	return b
}

func (b *TFJobBuilder) Annotations(annotations map[string]string) *TFJobBuilder {
	if annotations != nil && len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

func (b *TFJobBuilder) Labels(labels map[string]string) *TFJobBuilder {
	if labels != nil && len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

func (b *TFJobBuilder) EnableChief() *TFJobBuilder {
	b.args.UseChief = true
	return b
}

func (b *TFJobBuilder) ChiefCPU(cpu string) *TFJobBuilder {
	if cpu != "" {
		b.args.ChiefCpu = cpu
	}
	return b
}

func (b *TFJobBuilder) ChiefMemory(mem string) *TFJobBuilder {
	if mem != "" {
		b.args.ChiefMemory = mem
	}
	return b
}

func (b *TFJobBuilder) ChiefPort(port int) *TFJobBuilder {
	if port > 0 {
		b.args.ChiefPort = port
	}
	return b
}

func (b *TFJobBuilder) ChiefSelectors(selectors map[string]string) *TFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["chief-selector"] = &s
	}
	return b
}

func (b *TFJobBuilder) Datas(volumes map[string]string) *TFJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

func (b *TFJobBuilder) DataDirs(volumes map[string]string) *TFJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

func (b *TFJobBuilder) EnableEvaluator() *TFJobBuilder {
	b.args.UseEvaluator = true
	return b
}

func (b *TFJobBuilder) EvaluatorCPU(cpu string) *TFJobBuilder {
	if cpu != "" {
		b.args.EvaluatorCpu = cpu
	}
	return b
}

func (b *TFJobBuilder) EvaluatorMemory(mem string) *TFJobBuilder {
	if mem != "" {
		b.args.EvaluatorMemory = mem
	}
	return b
}

func (b *TFJobBuilder) EvaluatorSelectors(selectors map[string]string) *TFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["evaluator-selector"] = &s
	}
	return b
}

func (b *TFJobBuilder) LogDir(dir string) *TFJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

func (b *TFJobBuilder) Priority(priority string) *TFJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

func (b *TFJobBuilder) PsCount(count int) *TFJobBuilder {
	if count > 0 {
		b.args.PSCount = count
	}
	return b
}

func (b *TFJobBuilder) PsCPU(cpu string) *TFJobBuilder {
	if cpu != "" {
		b.args.PSCpu = cpu
	}
	return b
}

func (b *TFJobBuilder) PsGPU(gpu int) *TFJobBuilder {
	if gpu > 0 {
		b.args.PSGpu = gpu
	}
	return b
}

func (b *TFJobBuilder) PsImage(image string) *TFJobBuilder {
	if image != "" {
		b.args.PSImage = image
	}
	return b
}

func (b *TFJobBuilder) PsMemory(mem string) *TFJobBuilder {
	if mem != "" {
		b.args.PSMemory = mem
	}
	return b
}

func (b *TFJobBuilder) PsPort(port int) *TFJobBuilder {
	if port > 0 {
		b.args.PSPort = port
	}
	return b
}

func (b *TFJobBuilder) PsSelectors(selectors map[string]string) *TFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["ps-selector"] = &s
	}
	return b
}

func (b *TFJobBuilder) EnableRDMA() *TFJobBuilder {
	b.args.EnableRDMA = true
	return b
}

func (b *TFJobBuilder) SyncImage(image string) *TFJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

func (b *TFJobBuilder) SyncMode(mode string) *TFJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

func (b *TFJobBuilder) SyncSource(source string) *TFJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

func (b *TFJobBuilder) EnableTensorboard() *TFJobBuilder {
	b.args.UseTensorboard = true
	return b
}

func (b *TFJobBuilder) TensorboardImage(image string) *TFJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

func (b *TFJobBuilder) WorkerCPU(cpu string) *TFJobBuilder {
	if cpu != "" {
		b.args.WorkerCpu = cpu
	}
	return b
}

func (b *TFJobBuilder) WorkerImage(image string) *TFJobBuilder {
	if image != "" {
		b.args.WorkerImage = image
	}
	return b
}

func (b *TFJobBuilder) WorkerMemory(mem string) *TFJobBuilder {
	if mem != "" {
		b.args.WorkerMemory = mem
	}
	return b
}

func (b *TFJobBuilder) WorkerPort(port int) *TFJobBuilder {
	if port > 0 {
		b.args.WorkerPort = port
	}
	return b
}

func (b *TFJobBuilder) WorkerSelectors(selectors map[string]string) *TFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["worker-selector"] = &s
	}
	return b
}

func (b *TFJobBuilder) WorkerCount(count int) *TFJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}
func (b *TFJobBuilder) ImagePullSecrets(secrets []string) *TFJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = &secrets
	}
	return b
}

func (b *TFJobBuilder) CleanPodPolicy(policy string) *TFJobBuilder {
	if policy != "" {
		b.args.CleanPodPolicy = policy
	}
	return b
}

func (b *TFJobBuilder) RoleSequence(roles []string) *TFJobBuilder {
	if roles != nil && len(roles) != 0 {
		b.argValues["role-sequence"] = strings.Join(roles, ",")
	}
	return b
}

func (b *TFJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.TFTrainingJob, b.args), nil
}
