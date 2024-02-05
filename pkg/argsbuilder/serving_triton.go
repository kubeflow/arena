// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License
package argsbuilder

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type TritonServingArgsBuilder struct {
	args        *types.TritonServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewTritonServingArgsBuilder(args *types.TritonServingArgs) ArgsBuilder {
	args.Type = types.TritonServingJob
	s := &TritonServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewServingArgsBuilder(&s.args.CommonServingArgs),
	)
	return s
}

func (s *TritonServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *TritonServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *TritonServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *TritonServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var loadModels []string
	command.Flags().StringVar(&s.args.Backend, "backend", "", "the backend type of triton server. Valid values: [vllm|trt-llm]")
	command.Flags().StringVar(&s.args.ModelRepository, "model-repository", "", "the path of triton model path")
	command.Flags().IntVar(&s.args.HttpPort, "http-port", 8000, "the port of http serving server")
	command.Flags().IntVar(&s.args.GrpcPort, "grpc-port", 8001, "the port of grpc serving server")
	command.Flags().IntVar(&s.args.MetricsPort, "metrics-port", 8002, "the port of metrics server")
	command.Flags().BoolVar(&s.args.AllowMetrics, "allow-metrics", false, "open metrics")
	command.Flags().StringVar(&s.args.Command, "command", "", "the command will inject to container's command.")
	command.Flags().StringVar(&s.args.ExtendCommand, "extend-command", "", "the command will attach to server's command.")
	command.Flags().StringArrayVar(&loadModels, "load-model", []string{}, `giving names of model to load, usage:"--load-model <model-name>"`)

	s.AddArgValue("load-model", &loadModels)
}

func (s *TritonServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	if err := s.setLoadModels(); err != nil {
		return err
	}
	return nil
}

func (s *TritonServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.checkPortsIsOk(); err != nil {
		return err
	}
	if err := s.validate(); err != nil {
		return err
	}
	return nil
}

func (s *TritonServingArgsBuilder) validate() (err error) {
	if s.args.Backend != "" {
		if s.args.Backend != "vllm" && s.args.Backend != "trt-llm" {
			return fmt.Errorf("backend %s is Invalid. Triton backend only supports vllm or trt-llm", s.args.Backend)
		}
		if s.args.GPUCount == 0 {
			return fmt.Errorf("--gpus must be specific at least 1 GPU")
		}
	}
	return nil
}

func (s *TritonServingArgsBuilder) checkPortsIsOk() error {
	switch {
	case s.args.HttpPort != 0:
		return nil
	case s.args.GrpcPort != 0:
		return nil
	}
	return fmt.Errorf("all ports are 0, invalid configuration")
}

func (s *TritonServingArgsBuilder) setLoadModels() error {
	argKey := "load-model"
	var loadModels *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	loadModels = value.(*[]string)

	if len(*loadModels) > 0 {
		s.args.LoadModels = *loadModels
	} else {
		s.args.LoadModels = []string{}
	}

	log.Debugf("Load Models: %v", s.args.LoadModels)
	return nil
}
