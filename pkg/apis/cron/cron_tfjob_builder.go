// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cron

import (
	"strings"

	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type CronTFJobBuilder struct {
	args      *types.CronTFJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
	*training.TFJobBuilder
}

func NewCronTFJobBuilder() *CronTFJobBuilder {
	args := &types.CronTFJobArgs{
		SubmitTFJobArgs: types.SubmitTFJobArgs{
			CleanPodPolicy:        "Running",
			CommonSubmitArgs:      training.DefaultCommonSubmitArgs,
			SubmitTensorboardArgs: training.DefaultSubmitTensorboardArgs,
		},
	}
	return &CronTFJobBuilder{
		args:         args,
		argValues:    map[string]interface{}{},
		ArgsBuilder:  argsbuilder.NewCronTFJobArgsBuilder(args),
		TFJobBuilder: training.NewTFJobBuilder(&args.SubmitTFJobArgs),
	}
}

func (c *CronTFJobBuilder) Name(name string) *CronTFJobBuilder {
	c.TFJobBuilder.Name(name)
	return c
}

func (c *CronTFJobBuilder) Schedule(schedule string) *CronTFJobBuilder {
	if schedule != "" {
		c.args.Schedule = schedule
	}
	return c
}

func (c *CronTFJobBuilder) ConcurrencyPolicy(concurrencyPolicy string) *CronTFJobBuilder {
	if concurrencyPolicy != "" {
		c.args.ConcurrencyPolicy = concurrencyPolicy
	}
	return c
}

func (c *CronTFJobBuilder) Deadline(deadline string) *CronTFJobBuilder {
	if deadline != "" {
		c.args.Deadline = deadline
	}
	return c
}

func (c *CronTFJobBuilder) HistoryLimit(historyLimit int) *CronTFJobBuilder {
	if historyLimit > 0 {
		c.args.HistoryLimit = historyLimit
	}
	return c
}

func (c *CronTFJobBuilder) WorkingDir(dir string) *CronTFJobBuilder {
	c.TFJobBuilder.WorkingDir(dir)
	return c
}

func (c *CronTFJobBuilder) Envs(envs map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.Envs(envs)
	return c
}

func (c *CronTFJobBuilder) GPUCount(count int) *CronTFJobBuilder {
	c.TFJobBuilder.GPUCount(count)
	return c
}

func (c *CronTFJobBuilder) Image(image string) *CronTFJobBuilder {
	c.TFJobBuilder.Image(image)
	return c
}

func (c *CronTFJobBuilder) Tolerations(tolerations []string) *CronTFJobBuilder {
	c.TFJobBuilder.Tolerations(tolerations)
	return c
}

func (c *CronTFJobBuilder) ConfigFiles(files map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.ConfigFiles(files)
	return c
}

func (c *CronTFJobBuilder) NodeSelectors(selectors map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.NodeSelectors(selectors)
	return c
}

func (c *CronTFJobBuilder) Annotations(annotations map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.Annotations(annotations)
	return c
}

func (c *CronTFJobBuilder) Labels(labels map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.Labels(labels)
	return c
}

func (c *CronTFJobBuilder) EnableChief() *CronTFJobBuilder {
	c.TFJobBuilder.EnableChief()
	return c
}

func (c *CronTFJobBuilder) ChiefCPU(cpu string) *CronTFJobBuilder {
	c.TFJobBuilder.ChiefCPU(cpu)
	return c
}

func (c *CronTFJobBuilder) ChiefMemory(mem string) *CronTFJobBuilder {
	c.TFJobBuilder.ChiefMemory(mem)
	return c
}

func (c *CronTFJobBuilder) ChiefPort(port int) *CronTFJobBuilder {
	c.TFJobBuilder.ChiefPort(port)
	return c
}

func (c *CronTFJobBuilder) ChiefSelectors(selectors map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.ChiefSelectors(selectors)
	return c
}

func (c *CronTFJobBuilder) Datas(volumes map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.Datas(volumes)
	return c
}

func (c *CronTFJobBuilder) DataDirs(volumes map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.DataDirs(volumes)
	return c
}

func (c *CronTFJobBuilder) EnableEvaluator() *CronTFJobBuilder {
	c.TFJobBuilder.EnableEvaluator()
	return c
}

func (c *CronTFJobBuilder) EvaluatorCPU(cpu string) *CronTFJobBuilder {
	c.TFJobBuilder.EvaluatorCPU(cpu)
	return c
}

func (c *CronTFJobBuilder) EvaluatorMemory(mem string) *CronTFJobBuilder {
	c.TFJobBuilder.EvaluatorMemory(mem)
	return c
}

func (c *CronTFJobBuilder) EvaluatorSelectors(selectors map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.EvaluatorSelectors(selectors)
	return c
}

func (c *CronTFJobBuilder) LogDir(dir string) *CronTFJobBuilder {
	c.TFJobBuilder.LogDir(dir)
	return c
}

func (c *CronTFJobBuilder) Priority(priority string) *CronTFJobBuilder {
	c.TFJobBuilder.Priority(priority)
	return c
}

func (c *CronTFJobBuilder) PsCount(count int) *CronTFJobBuilder {
	c.TFJobBuilder.PsCount(count)
	return c
}

func (c *CronTFJobBuilder) PsGPU(gpu int) *CronTFJobBuilder {
	c.TFJobBuilder.PsGPU(gpu)
	return c
}

func (c *CronTFJobBuilder) PsCPU(cpu string) *CronTFJobBuilder {
	c.TFJobBuilder.PsCPU(cpu)
	return c
}

func (c *CronTFJobBuilder) PsImage(image string) *CronTFJobBuilder {
	c.TFJobBuilder.PsImage(image)
	return c
}

func (c *CronTFJobBuilder) PsMemory(mem string) *CronTFJobBuilder {
	c.TFJobBuilder.PsMemory(mem)
	return c
}

func (c *CronTFJobBuilder) PsPort(port int) *CronTFJobBuilder {
	c.TFJobBuilder.PsPort(port)
	return c
}

func (c *CronTFJobBuilder) PsSelectors(selectors map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.PsSelectors(selectors)
	return c
}

func (c *CronTFJobBuilder) EnableRDMA() *CronTFJobBuilder {
	c.TFJobBuilder.EnableRDMA()
	return c
}

func (c *CronTFJobBuilder) SyncImage(image string) *CronTFJobBuilder {
	c.TFJobBuilder.SyncImage(image)
	return c
}

func (c *CronTFJobBuilder) SyncMode(mode string) *CronTFJobBuilder {
	c.TFJobBuilder.SyncMode(mode)
	return c
}

func (c *CronTFJobBuilder) SyncSource(source string) *CronTFJobBuilder {
	c.TFJobBuilder.SyncSource(source)
	return c
}

func (c *CronTFJobBuilder) EnableTensorboard() *CronTFJobBuilder {
	c.TFJobBuilder.EnableTensorboard()
	return c
}

func (c *CronTFJobBuilder) TensorboardImage(image string) *CronTFJobBuilder {
	c.TFJobBuilder.TensorboardImage(image)
	return c
}

func (c *CronTFJobBuilder) WorkerCPU(cpu string) *CronTFJobBuilder {
	c.TFJobBuilder.WorkerCPU(cpu)
	return c
}

func (c *CronTFJobBuilder) WorkerImage(image string) *CronTFJobBuilder {
	c.TFJobBuilder.WorkerImage(image)
	return c
}

func (c *CronTFJobBuilder) WorkerMemory(mem string) *CronTFJobBuilder {
	c.TFJobBuilder.WorkerMemory(mem)
	return c
}

func (c *CronTFJobBuilder) WorkerPort(port int) *CronTFJobBuilder {
	c.TFJobBuilder.WorkerPort(port)
	return c
}

func (c *CronTFJobBuilder) WorkerSelectors(selectors map[string]string) *CronTFJobBuilder {
	c.TFJobBuilder.WorkerSelectors(selectors)
	return c
}

func (c *CronTFJobBuilder) WorkerCount(count int) *CronTFJobBuilder {
	c.TFJobBuilder.WorkerCount(count)
	return c
}
func (c *CronTFJobBuilder) ImagePullSecrets(secrets []string) *CronTFJobBuilder {
	c.TFJobBuilder.ImagePullSecrets(secrets)
	return c
}

func (c *CronTFJobBuilder) CleanPodPolicy(policy string) *CronTFJobBuilder {
	c.TFJobBuilder.CleanPodPolicy(policy)
	return c
}

func (c *CronTFJobBuilder) RoleSequence(roles []string) *CronTFJobBuilder {
	c.TFJobBuilder.RoleSequence(roles)
	return c
}

func (c *CronTFJobBuilder) ActiveDeadlineSeconds(ttl int64) *CronTFJobBuilder {
	c.TFJobBuilder.ActiveDeadlineSeconds(ttl)
	return c
}

func (c *CronTFJobBuilder) TTLSecondsAfterFinished(ttl int32) *CronTFJobBuilder {
	c.TFJobBuilder.TTLSecondsAfterFinished(ttl)
	return c
}

func (c *CronTFJobBuilder) Shell(shell string) *CronTFJobBuilder {
	if shell != "" {
		c.args.Shell = shell
	}
	return c
}

func (c *CronTFJobBuilder) Command(args []string) *CronTFJobBuilder {
	if c.args.Command == "" {
		c.args.Command = strings.Join(args, " ")
	}
	return c
}

func (c *CronTFJobBuilder) Build() (*Job, error) {
	for key, value := range c.argValues {
		c.AddArgValue(key, value)
	}

	for key, value := range c.GetArgValues() {
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
