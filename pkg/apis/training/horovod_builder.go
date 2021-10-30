package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type HorovodJobBuilder struct {
	args      *types.SubmitHorovodJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewHorovodJobBuilder() *HorovodJobBuilder {
	args := &types.SubmitHorovodJobArgs{
		CommonSubmitArgs:      DefaultCommonSubmitArgs,
		SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
	}
	return &HorovodJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitHorovodJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *HorovodJobBuilder) Name(name string) *HorovodJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Shell is used to set bash or sh
func (b *HorovodJobBuilder) Shell(shell string) *HorovodJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *HorovodJobBuilder) Command(args []string) *HorovodJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *HorovodJobBuilder) WorkingDir(dir string) *HorovodJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *HorovodJobBuilder) Envs(envs map[string]string) *HorovodJobBuilder {
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
func (b *HorovodJobBuilder) GPUCount(count int) *HorovodJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *HorovodJobBuilder) Image(image string) *HorovodJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *HorovodJobBuilder) Tolerations(tolerations []string) *HorovodJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers,match option --config-file
func (b *HorovodJobBuilder) ConfigFiles(files map[string]string) *HorovodJobBuilder {
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
func (b *HorovodJobBuilder) NodeSelectors(selectors map[string]string) *HorovodJobBuilder {
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
func (b *HorovodJobBuilder) Annotations(annotations map[string]string) *HorovodJobBuilder {
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
func (b *HorovodJobBuilder) Datas(volumes map[string]string) *HorovodJobBuilder {
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
func (b *HorovodJobBuilder) DataDirs(volumes map[string]string) *HorovodJobBuilder {
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
func (b *HorovodJobBuilder) LogDir(dir string) *HorovodJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *HorovodJobBuilder) Priority(priority string) *HorovodJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma,match option --rdma
func (b *HorovodJobBuilder) EnableRDMA() *HorovodJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image,match option --sync-image
func (b *HorovodJobBuilder) SyncImage(image string) *HorovodJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode,match option --sync-mode
func (b *HorovodJobBuilder) SyncMode(mode string) *HorovodJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source,match option --sync-source
func (b *HorovodJobBuilder) SyncSource(source string) *HorovodJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *HorovodJobBuilder) EnableTensorboard() *HorovodJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to enable tensorboard image
func (b *HorovodJobBuilder) TensorboardImage(image string) *HorovodJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrests,match option --image-pull-secret
func (b *HorovodJobBuilder) ImagePullSecrets(secrets []string) *HorovodJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = secrets
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *HorovodJobBuilder) WorkerCount(count int) *HorovodJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// CPU assign cpu limts,match option --cpu
func (b *HorovodJobBuilder) CPU(cpu string) *HorovodJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits,match option --memory
func (b *HorovodJobBuilder) Memory(memory string) *HorovodJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// SSHPort set the ssh port,match option --ssh-port
func (b *HorovodJobBuilder) SSHPort(port int) *HorovodJobBuilder {
	if port != 0 {
		b.args.SSHPort = port
	}
	return b
}

// Build is used to build the job
func (b *HorovodJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.HorovodTrainingJob, b.args), nil
}
