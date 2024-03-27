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

const (
	DefaultTfServingImage = "tensorflow/serving:latest"
	regexp4serviceName    = "^[a-z0-9A-Z_-]+$"
)

type TensorflowServingArgsBuilder struct {
	args        *types.TensorFlowServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewTensorflowServingArgsBuilder(args *types.TensorFlowServingArgs) ArgsBuilder {
	args.Type = types.TFServingJob
	s := &TensorflowServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewServingArgsBuilder(&s.args.CommonServingArgs),
	)
	s.AddArgValue("default-image", DefaultTfServingImage)
	return s
}

func (s *TensorflowServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *TensorflowServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *TensorflowServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *TensorflowServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().IntVar(&s.args.Port, "port", 8500, "the port of tensorflow gRPC listening port")
	command.Flags().IntVar(&s.args.RestfulPort, "restfulPort", 8501, "the port of tensorflow RESTful listening port")
	_ = command.Flags().MarkDeprecated("restfulPort", "please use --restful-port instead")
	command.Flags().IntVar(&s.args.RestfulPort, "restful-port", 8501, "the port of tensorflow RESTful listening port")

	command.Flags().StringVar(&s.args.ModelName, "modelName", "", "the model name for serving")
	_ = command.Flags().MarkDeprecated("modelName", "please use --model-name instead")

	command.Flags().StringVar(&s.args.ModelPath, "modelPath", "", "the model path for serving in the container")
	_ = command.Flags().MarkDeprecated("modelPath", "please use --model-path instead")
	command.Flags().StringVar(&s.args.ModelPath, "model-path", "", "the model path for serving in the container, ignored if --model-config-file flag is set, otherwise required")

	command.Flags().StringVar(&s.args.ModelConfigFile, "modelConfigFile", "", "corresponding with --model_config_file in tensorflow serving")
	_ = command.Flags().MarkDeprecated("modelConfigFile", "please use --model-config-file instead")
	command.Flags().StringVar(&s.args.ModelConfigFile, "model-config-file", "", "corresponding with --model_config_file in tensorflow serving")

	command.Flags().StringVar(&s.args.MonitoringConfigFile, "monitoring-config-file", "", "corresponding with --monitoring_config_file in tensorflow serving")
	command.Flags().StringVar(&s.args.VersionPolicy, "versionPolicy", "", "support latest, latest:N, specific:N, all")
	_ = command.Flags().MarkDeprecated("versionPolicy", "please use --version-policy instead")
	command.Flags().StringVar(&s.args.VersionPolicy, "version-policy", "", "support latest, latest:N, specific:N, all")
	_ = command.Flags().MarkDeprecated("version-policy", "please use --model-config-file instead")

	command.Flags().StringVar(&s.args.Command, "command", "", "the command will inject to container's command.")
}

func (s *TensorflowServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *TensorflowServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.checkPortsIsOk(); err != nil {
		return err
	}
	if err := s.preprocess(); err != nil {
		return err
	}
	return nil
}

func (s *TensorflowServingArgsBuilder) validateModelName() error {
	if s.args.ModelName == "" {
		return fmt.Errorf("model name cannot be blank")
	}
	reg := regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.args.ModelName)
	if !matched {
		return fmt.Errorf("model name should be numbers, letters, dashes, and underscores ONLY")
	}

	return nil
}

func (s *TensorflowServingArgsBuilder) preprocess() (err error) {
	//serveTensorFlowArgs.Command = strings.Join(args, " ")
	log.Debugf("command: %s", s.args.Command)
	if s.args.Image == "" {
		return fmt.Errorf("image must be specified.")
	}
	if s.args.ModelConfigFile == "" {
		// need to validate modelName, modelPath if not specify modelConfigFile
		// 1. validate modelName
		err := s.validateModelName()
		if err != nil {
			return err
		}
		//2. validate modelPath
		if s.args.ModelPath == "" {
			return fmt.Errorf("modelPath should be specified if no modelConfigFile is specified")
		}
	} else {
		//populate content from modelConfigFile
		if s.args.ModelName != "" {
			log.Infof("modelConfigFile=%s is specified, so --model-name will be ingored", s.args.ModelConfigFile)
		}
		if s.args.ModelPath != "" {
			log.Infof("modelConfigFile=%s is specified, so --model-path will be ignored", s.args.ModelConfigFile)
		}
	}
	return nil
}

func (s *TensorflowServingArgsBuilder) checkPortsIsOk() error {
	switch {
	case s.args.Port != 0:
		return nil
	case s.args.RestfulPort != 0:
		return nil
	}
	return fmt.Errorf("all  ports are 0,invalid configuration.")
}
