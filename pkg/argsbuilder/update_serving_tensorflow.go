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
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/apis/types"
)

type UpdateTensorflowServingArgsBuilder struct {
	args        *types.UpdateTensorFlowServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewUpdateTensorflowServingArgsBuilder(args *types.UpdateTensorFlowServingArgs) ArgsBuilder {
	args.Type = types.TFServingJob
	s := &UpdateTensorflowServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewUpdateServingArgsBuilder(&s.args.CommonUpdateServingArgs),
	)
	s.AddArgValue("default-image", DefaultTfServingImage)
	return s
}

func (s *UpdateTensorflowServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *UpdateTensorflowServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *UpdateTensorflowServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *UpdateTensorflowServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVar(&s.args.ModelName, "model-name", "", "the model name for serving, ignored if --model-config-file flag is set")
	command.Flags().StringVar(&s.args.ModelPath, "model-path", "", "the model path for serving in the container, ignored if --model-config-file flag is set, otherwise required")
	command.Flags().StringVar(&s.args.ModelConfigFile, "model-config-file", "", "corresponding with --model_config_file in tensorflow serving")
	command.Flags().StringVar(&s.args.MonitoringConfigFile, "monitoring-config-file", "", "corresponding with --monitoring_config_file in tensorflow serving")
}

func (s *UpdateTensorflowServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}

	return nil
}

func (s *UpdateTensorflowServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}

	if err := s.preprocess(); err != nil {
		return err
	}
	return nil
}

func (s *UpdateTensorflowServingArgsBuilder) validateModelName() error {
	if s.args.ModelName != "" {
		reg := regexp.MustCompile(regexp4serviceName)
		matched := reg.MatchString(s.args.ModelName)
		if !matched {
			return fmt.Errorf("model name should be numbers, letters, dashes, and underscores ONLY")
		}
	}

	return nil
}

func (s *UpdateTensorflowServingArgsBuilder) preprocess() error {
	log.Debugf("command: %s", s.args.Command)
	err := s.validateModelName()
	if err != nil {
		return err
	}
	return nil
}
