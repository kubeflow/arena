package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type MPIJobBuilder struct {
	args      *types.SubmitMPIJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewMPIJobBuilder() *MPIJobBuilder {
	args := &types.SubmitMPIJobArgs{
		CommonSubmitArgs:      DefaultCommonSubmitArgs,
		SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
	}
	return &MPIJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitMPIJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *MPIJobBuilder) Name(name string) *MPIJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Command is used to set job command
func (b *MPIJobBuilder) Command(args []string) *MPIJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *MPIJobBuilder) WorkingDir(dir string) *MPIJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *MPIJobBuilder) Envs(envs map[string]string) *MPIJobBuilder {
	if envs != nil && len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *MPIJobBuilder) GPUCount(count int) *MPIJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *MPIJobBuilder) Image(image string) *MPIJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *MPIJobBuilder) Tolerations(tolerations []string) *MPIJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *MPIJobBuilder) ConfigFiles(files map[string]string) *MPIJobBuilder {
	if files != nil && len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// NodeSelectors is used to set node selectors for scheduling job,match option --selector
func (b *MPIJobBuilder) NodeSelectors(selectors map[string]string) *MPIJobBuilder {
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
func (b *MPIJobBuilder) Annotations(annotations map[string]string) *MPIJobBuilder {
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
func (b *MPIJobBuilder) Datas(volumes map[string]string) *MPIJobBuilder {
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
func (b *MPIJobBuilder) DataDirs(volumes map[string]string) *MPIJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// LogDir is used to set log directory,match option --logdir
func (b *MPIJobBuilder) LogDir(dir string) *MPIJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *MPIJobBuilder) Priority(priority string) *MPIJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma,match option --rdma
func (b *MPIJobBuilder) EnableRDMA() *MPIJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image,match option --sync-image
func (b *MPIJobBuilder) SyncImage(image string) *MPIJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (b *MPIJobBuilder) SyncMode(mode string) *MPIJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source,match option --sync-source
func (b *MPIJobBuilder) SyncSource(source string) *MPIJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *MPIJobBuilder) EnableTensorboard() *MPIJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to enable tensorboard image
func (b *MPIJobBuilder) TensorboardImage(image string) *MPIJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (b *MPIJobBuilder) ImagePullSecrets(secrets []string) *MPIJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = secrets
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *MPIJobBuilder) WorkerCount(count int) *MPIJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// CPU assign cpu limts,match option --cpu
func (b *MPIJobBuilder) CPU(cpu string) *MPIJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *MPIJobBuilder) Memory(memory string) *MPIJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// EnableGPUTopology is used to enable gpu topology scheduling
func (b *MPIJobBuilder) EnableGPUTopology() *MPIJobBuilder {
	b.args.GPUTopology = true
	return b
}

// Build is used to build the job
func (b *MPIJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.MPITrainingJob, b.args), nil
}
