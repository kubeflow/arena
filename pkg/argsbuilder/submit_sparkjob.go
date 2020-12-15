// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License
package argsbuilder

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type SubmitSparkJobArgsBuilder struct {
	args        *types.SubmitSparkJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitSparkJobArgsBuilder(args *types.SubmitSparkJobArgs) ArgsBuilder {
	args.TrainingType = types.SparkTrainingJob
	s := &SubmitSparkJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	return s
}

func (s *SubmitSparkJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitSparkJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitSparkJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitSparkJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVar(&s.args.Name, "name", "", "override name")
	command.MarkFlagRequired("name")

	command.Flags().StringVar(&s.args.Image, "image", "registry.aliyuncs.com/acs/spark:v2.4.0", "the docker image name of training job")
	command.Flags().IntVar(&s.args.Executor.Replicas, "replicas", 1, "the executor's number to run the distributed training.")
	command.Flags().StringVar(&s.args.MainClass, "main-class", "org.apache.spark.examples.SparkPi", "main class of your jar")
	command.Flags().StringVar(&s.args.Jar, "jar", "local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar", "jar path in image")

	// cpu and memory request
	command.Flags().IntVar(&s.args.Driver.CPURequest, "driver-cpu-request", 1, "cpu request for driver pod")
	command.Flags().StringVar(&s.args.Driver.MemoryRequest, "driver-memory-request", "500m", "memory request for driver pod (min is 500m)")
	command.Flags().IntVar(&s.args.Executor.CPURequest, "executor-cpu-request", 1, "cpu request for executor pod")
	command.Flags().StringVar(&s.args.Executor.MemoryRequest, "executor-memory-request", "500m", "memory request for executor pod (min is 500m)")
}

func (s *SubmitSparkJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SubmitSparkJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.isValid(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitSparkJobArgsBuilder) isValid() error {
	if s.args.Executor.Replicas == 0 {
		return errors.New("WorkersMustMoreThanOne")
	}
	return nil
}