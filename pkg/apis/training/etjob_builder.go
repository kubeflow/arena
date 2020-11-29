package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type ETJobBuilder struct {
	args      *types.SubmitETJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewETJobBuilder() *ETJobBuilder {
	args := &types.SubmitETJobArgs{
		MaxWorkers:            1000,
		MinWorkers:            1,
		CommonSubmitArgs:      defaultCommonSubmitArgs,
		SubmitTensorboardArgs: defaultSubmitTensorboardArgs,
	}
	return &ETJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitETJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *ETJobBuilder) Name(name string) *ETJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Command is used to set job command
func (b *ETJobBuilder) Command(args []string) *ETJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *ETJobBuilder) WorkingDir(dir string) *ETJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *ETJobBuilder) Envs(envs map[string]string) *ETJobBuilder {
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
func (b *ETJobBuilder) GPUCount(count int) *ETJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *ETJobBuilder) Image(image string) *ETJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *ETJobBuilder) Tolerations(tolerations []string) *ETJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *ETJobBuilder) ConfigFiles(files map[string]string) *ETJobBuilder {
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
func (b *ETJobBuilder) NodeSelectors(selectors map[string]string) *ETJobBuilder {
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
func (b *ETJobBuilder) Annotations(annotations map[string]string) *ETJobBuilder {
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
func (b *ETJobBuilder) Datas(volumes map[string]string) *ETJobBuilder {
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
func (b *ETJobBuilder) DataDirs(volumes map[string]string) *ETJobBuilder {
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
func (b *ETJobBuilder) LogDir(dir string) *ETJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *ETJobBuilder) Priority(priority string) *ETJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma,match option --rdma
func (b *ETJobBuilder) EnableRDMA() *ETJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image,match option --sync-image
func (b *ETJobBuilder) SyncImage(image string) *ETJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (b *ETJobBuilder) SyncMode(mode string) *ETJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source,match option --sync-source
func (b *ETJobBuilder) SyncSource(source string) *ETJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *ETJobBuilder) EnableTensorboard() *ETJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to enable tensorboard image
func (b *ETJobBuilder) TensorboardImage(image string) *ETJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (b *ETJobBuilder) ImagePullSecrets(secrets []string) *ETJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = secrets
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *ETJobBuilder) WorkerCount(count int) *ETJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// CPU assign cpu limts,match option --cpu
func (b *ETJobBuilder) CPU(cpu string) *ETJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *ETJobBuilder) Memory(memory string) *ETJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// MaxWorkers assign max workers,match option --max-workers
func (b *ETJobBuilder) MaxWorkers(count int) *ETJobBuilder {
	if count > 0 {
		b.args.MaxWorkers = count
	}
	return b
}

// MinWorkers assign min workers,match option --min-workers
func (b *ETJobBuilder) MinWorkers(count int) *ETJobBuilder {
	if count > 0 {
		b.args.MinWorkers = count
	}
	return b
}

// Build is used to build the job
func (b *ETJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.ETTrainingJob, b.args), nil
}
