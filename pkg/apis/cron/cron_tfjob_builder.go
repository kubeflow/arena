package cron

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
	"strings"
)

type CronCronTFJobBuilder struct {
	args      *types.CronTFJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewCronTFJobBuilder() *CronCronTFJobBuilder {
	args := &types.CronTFJobArgs{}
	return &CronCronTFJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewCronTFJobArgsBuilder(args),
	}
}

func (c *CronCronTFJobBuilder) Name(name string) *CronCronTFJobBuilder {
	if name != "" {
		c.args.Name = name
	}
	return c
}

func (c *CronCronTFJobBuilder) Schedule(schedule string) *CronCronTFJobBuilder {
	if schedule != "" {
		c.args.Schedule = schedule
	}
	return c
}

func (c *CronCronTFJobBuilder) ConcurrencyPolicy(concurrencyPolicy string) *CronCronTFJobBuilder {
	if concurrencyPolicy != "" {
		c.args.ConcurrencyPolicy = concurrencyPolicy
	}
	return c
}

func (c *CronCronTFJobBuilder) Deadline(deadline string) *CronCronTFJobBuilder {
	if deadline != "" {
		c.args.Deadline = deadline
	}
	return c
}

func (c *CronCronTFJobBuilder) HistoryLimit(historyLimit int) *CronCronTFJobBuilder {
	if historyLimit > 0 {
		c.args.HistoryLimit = historyLimit
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkingDir(dir string) *CronCronTFJobBuilder {
	if dir != "" {
		c.args.WorkingDir = dir
	}
	return c
}

func (c *CronCronTFJobBuilder) Envs(envs map[string]string) *CronCronTFJobBuilder {
	if envs != nil && len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["env"] = &envSlice
	}
	return c
}

func (c *CronCronTFJobBuilder) GPUCount(count int) *CronCronTFJobBuilder {
	if count > 0 {
		c.args.GPUCount = count
	}
	return c
}

func (c *CronCronTFJobBuilder) Image(image string) *CronCronTFJobBuilder {
	if image != "" {
		c.args.Image = image
	}
	return c
}

func (c *CronCronTFJobBuilder) Tolerations(tolerations []string) *CronCronTFJobBuilder {
	c.argValues["toleration"] = &tolerations
	return c
}

func (c *CronCronTFJobBuilder) ConfigFiles(files map[string]string) *CronCronTFJobBuilder {
	if files != nil && len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		c.argValues["config-file"] = &filesSlice
	}
	return c
}

func (c *CronCronTFJobBuilder) NodeSelectors(selectors map[string]string) *CronCronTFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["selector"] = &selectorsSlice
	}
	return c
}

func (c *CronCronTFJobBuilder) Annotations(annotations map[string]string) *CronCronTFJobBuilder {
	if annotations != nil && len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["annotation"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) EnableChief() *CronCronTFJobBuilder {
	c.args.UseChief = true
	return c
}

func (c *CronCronTFJobBuilder) ChiefCPU(cpu string) *CronCronTFJobBuilder {
	if cpu != "" {
		c.args.ChiefCpu = cpu
	}
	return c
}

func (c *CronCronTFJobBuilder) ChiefMemory(mem string) *CronCronTFJobBuilder {
	if mem != "" {
		c.args.ChiefMemory = mem
	}
	return c
}

func (c *CronCronTFJobBuilder) ChiefPort(port int) *CronCronTFJobBuilder {
	if port > 0 {
		c.args.ChiefPort = port
	}
	return c
}

func (c *CronCronTFJobBuilder) ChiefSelectors(selectors map[string]string) *CronCronTFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["chief-selector"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) Datas(volumes map[string]string) *CronCronTFJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		c.argValues["data"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) DataDirs(volumes map[string]string) *CronCronTFJobBuilder {
	if volumes != nil && len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		c.argValues["data-dir"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) EnableEvaluator() *CronCronTFJobBuilder {
	c.args.UseEvaluator = true
	return c
}

func (c *CronCronTFJobBuilder) EvaluatorCPU(cpu string) *CronCronTFJobBuilder {
	if cpu != "" {
		c.args.EvaluatorCpu = cpu
	}
	return c
}

func (c *CronCronTFJobBuilder) EvaluatorMemory(mem string) *CronCronTFJobBuilder {
	if mem != "" {
		c.args.EvaluatorMemory = mem
	}
	return c
}

func (c *CronCronTFJobBuilder) EvaluatorSelectors(selectors map[string]string) *CronCronTFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["evaluator-selector"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) LogDir(dir string) *CronCronTFJobBuilder {
	if dir != "" {
		c.args.TrainingLogdir = dir
	}
	return c
}

func (c *CronCronTFJobBuilder) Priority(priority string) *CronCronTFJobBuilder {
	if priority != "" {
		c.args.PriorityClassName = priority
	}
	return c
}

func (c *CronCronTFJobBuilder) PsCount(count int) *CronCronTFJobBuilder {
	if count > 0 {
		c.args.PSCount = count
	}
	return c
}

func (c *CronCronTFJobBuilder) PsCPU(cpu string) *CronCronTFJobBuilder {
	if cpu != "" {
		c.args.PSCpu = cpu
	}
	return c
}

func (c *CronCronTFJobBuilder) PsImage(image string) *CronCronTFJobBuilder {
	if image != "" {
		c.args.PSImage = image
	}
	return c
}

func (c *CronCronTFJobBuilder) PsMemory(mem string) *CronCronTFJobBuilder {
	if mem != "" {
		c.args.PSMemory = mem
	}
	return c
}

func (c *CronCronTFJobBuilder) PsPort(port int) *CronCronTFJobBuilder {
	if port > 0 {
		c.args.PSPort = port
	}
	return c
}

func (c *CronCronTFJobBuilder) PsSelectors(selectors map[string]string) *CronCronTFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["ps-selector"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) EnableRDMA() *CronCronTFJobBuilder {
	c.args.EnableRDMA = true
	return c
}

func (c *CronCronTFJobBuilder) SyncImage(image string) *CronCronTFJobBuilder {
	if image != "" {
		c.args.SyncImage = image
	}
	return c
}

func (c *CronCronTFJobBuilder) SyncMode(mode string) *CronCronTFJobBuilder {
	if mode != "" {
		c.args.SyncMode = mode
	}
	return c
}

func (c *CronCronTFJobBuilder) SyncSource(source string) *CronCronTFJobBuilder {
	if source != "" {
		c.args.SyncSource = source
	}
	return c
}

func (c *CronCronTFJobBuilder) EnableTensorboard() *CronCronTFJobBuilder {
	c.args.UseTensorboard = true
	return c
}

func (c *CronCronTFJobBuilder) TensorboardImage(image string) *CronCronTFJobBuilder {
	if image != "" {
		c.args.TensorboardImage = image
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkerCPU(cpu string) *CronCronTFJobBuilder {
	if cpu != "" {
		c.args.WorkerCpu = cpu
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkerImage(image string) *CronCronTFJobBuilder {
	if image != "" {
		c.args.WorkerImage = image
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkerMemory(mem string) *CronCronTFJobBuilder {
	if mem != "" {
		c.args.WorkerMemory = mem
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkerPort(port int) *CronCronTFJobBuilder {
	if port > 0 {
		c.args.WorkerPort = port
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkerSelectors(selectors map[string]string) *CronCronTFJobBuilder {
	if selectors != nil && len(selectors) != 0 {
		s := []string{}
		for key, value := range selectors {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		c.argValues["worker-selector"] = &s
	}
	return c
}

func (c *CronCronTFJobBuilder) WorkerCount(count int) *CronCronTFJobBuilder {
	if count > 0 {
		c.args.WorkerCount = count
	}
	return c
}
func (c *CronCronTFJobBuilder) ImagePullSecrets(secrets []string) *CronCronTFJobBuilder {
	if secrets != nil {
		c.argValues["image-pull-secret"] = secrets
	}
	return c
}

func (c *CronCronTFJobBuilder) CleanPodPolicy(policy string) *CronCronTFJobBuilder {
	if policy != "" {
		c.args.CleanPodPolicy = policy
	}
	return c
}

func (c *CronCronTFJobBuilder) RoleSequence(roles []string) *CronCronTFJobBuilder {
	if roles != nil && len(roles) != 0 {
		c.argValues["role-sequence"] = strings.Join(roles, ",")
	}
	return c
}

func (c *CronCronTFJobBuilder) Command(args []string) *CronCronTFJobBuilder {
	c.args.Command = strings.Join(args, " ")
	return c
}

func (c *CronCronTFJobBuilder) Build() (*Job, error) {
	for key, value := range c.argValues {
		c.AddArgValue(key, value)
	}
	if err := c.PreBuild(); err != nil {
		return nil, err
	}
	if err := c.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(c.args.Name, types.CronTFTrainingJob, c.args), nil
}
