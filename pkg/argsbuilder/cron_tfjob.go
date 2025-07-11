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

package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type CronTFJobArgsBuilder struct {
	args        *types.CronTFJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewCronTFJobArgsBuilder(args *types.CronTFJobArgs) ArgsBuilder {
	args.TrainingType = types.TFTrainingJob
	c := &CronTFJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	c.AddSubBuilder(
		NewSubmitTFJobArgsBuilder(&c.args.SubmitTFJobArgs),
	)
	return c
}

func (c *CronTFJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*c)), ".")
	return items[len(items)-1]
}

func (c *CronTFJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		c.subBuilders[b.GetName()] = b
	}
	return c
}

func (c *CronTFJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range c.subBuilders {
		c.subBuilders[name].AddArgValue(key, value)
	}
	c.argValues[key] = value
	return c
}

func (c *CronTFJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range c.subBuilders {
		c.subBuilders[name].AddCommandFlags(command)
	}
	// cron task arguments
	command.Flags().StringVar(&c.args.Schedule, "schedule", "", "the schedule of cron task")
	command.Flags().StringVar(&c.args.ConcurrencyPolicy, "concurrency-policy", "Allow", "specifies how to treat concurrent executions of a task")
	command.Flags().BoolVar(&c.args.Suspend, "suspend", false, "if suspend the cron task")
	command.Flags().StringVar(&c.args.Deadline, "deadline", "", "the timestamp that a cron job can keep scheduling util then")
	command.Flags().IntVar(&c.args.HistoryLimit, "history-limit", 10, "the number of finished job history to retain")
}

func (c *CronTFJobArgsBuilder) PreBuild() error {
	for name := range c.subBuilders {
		if err := c.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CronTFJobArgsBuilder) Build() error {
	for name := range c.subBuilders {
		if err := c.subBuilders[name].Build(); err != nil {
			return err
		}
	}

	if err := c.check(); err != nil {
		return err
	}

	return nil
}

func (c *CronTFJobArgsBuilder) check() error {
	if c.args.Schedule == "" {
		return fmt.Errorf("--schedule must be set ")
	}
	return nil
}
