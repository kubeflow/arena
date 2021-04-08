package cron

import (
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
	"strings"
)

type CronTFJobBuilder struct {
	args      *types.CronTFJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewCronTFJobBuilder() *CronTFJobBuilder {
	args := &types.CronTFJobArgs{
		//TODO
	}
	return &CronTFJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewCronTFJobArgsBuilder(args),
	}
}

func (c *CronTFJobBuilder) Name(name string) *CronTFJobBuilder {
	if name != "" {
		c.args.Name = name
	}
	return c
}

func (c *CronTFJobBuilder) Command(args []string) *CronTFJobBuilder {
	c.args.Command = strings.Join(args, " ")
	return c
}

func (c *CronTFJobBuilder) Build() (*Job, error) {
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
