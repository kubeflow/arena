package training

import (
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type ScaleOutETJobBuilder struct {
	args      *types.ScaleOutETJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewScaleOutETJobBuilder() *ScaleOutETJobBuilder {
	args := &types.ScaleOutETJobArgs{
		ScaleETJobArgs: types.ScaleETJobArgs{
			Timeout: 60,
			Retry:   0,
			Count:   1,
		},
	}
	return &ScaleOutETJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewScaleOutETJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *ScaleOutETJobBuilder) Name(name string) *ScaleOutETJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Retry is used to set retry times
func (b *ScaleOutETJobBuilder) Retry(count int) *ScaleOutETJobBuilder {
	if count > 0 {
		b.args.Retry = count
	}
	return b
}

// Retry is used to set retry times
func (b *ScaleOutETJobBuilder) Count(count int) *ScaleOutETJobBuilder {
	if count > 0 {
		b.args.Count = count
	}
	return b
}

// Timeout is used to set timeout seconds
func (b *ScaleOutETJobBuilder) Timeout(timeout time.Duration) *ScaleOutETJobBuilder {
	b.argValues["timeout"] = &timeout
	return b
}

// Script is used to set scale script
func (b *ScaleOutETJobBuilder) Script(s string) *ScaleOutETJobBuilder {
	if s != "" {
		b.args.Script = s
	}
	return b
}

// Envs is used to set envs
func (b *ScaleOutETJobBuilder) Envs(envs map[string]string) *ScaleOutETJobBuilder {
	items := []string{}
	for key, value := range envs {
		items = append(items, fmt.Sprintf("%v=%v", key, value))
	}
	b.argValues["env"] = items
	return b
}

// Build is used to build the job
func (b *ScaleOutETJobBuilder) Build() (*Job, error) {
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
