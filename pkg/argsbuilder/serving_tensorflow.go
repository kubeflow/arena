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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultTfServingImage = "tensorflow/serving:latest"
	modelPathSeparator    = ":"
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

	command.Flags().StringVar(&s.args.ModelName, "modelName", "", "the model name for serving")
	command.Flags().MarkDeprecated("modelName", "please use --model-name instead")
	command.Flags().StringVar(&s.args.ModelName, "model-name", "", "the model name for serving")

	command.Flags().StringVar(&s.args.ModelPath, "modelPath", "", "the model path for serving in the container")
	command.Flags().MarkDeprecated("modelPath", "please use --model-path instead")
	command.Flags().StringVar(&s.args.ModelPath, "model-path", "", "the model path for serving in the container")

	command.Flags().StringVar(&s.args.ModelConfigFile, "modelConfigFile", "", "Corresponding with --model_config_file in tensorflow serving")
	command.Flags().StringVar(&s.args.VersionPolicy, "versionPolicy", "", "support latest, latest:N, specific:N, all")
	command.Flags().MarkDeprecated("versionPolicy", "please use --version-policy instead")
	command.Flags().StringVar(&s.args.VersionPolicy, "version-policy", "", "support latest, latest:N, specific:N, all")
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
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
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
		// need to validate modelName, modelPath and versionPolicy if not specify modelConfigFile
		// 1. validate modelName
		err := s.validateModelName()
		if err != nil {
			return err
		}
		//2. validate modelPath
		if s.args.ModelPath == "" {
			return fmt.Errorf("modelPath should be specified if no modelConfigFile is specified")
		}

		//3. validate versionPolicy
		err = s.validateVersionPolicy()
		if err != nil {
			return err
		}
		//populate content according to CLI parameters
		s.args.ModelConfigFileContent = s.generateModelConfigFileContent()

	} else {
		//populate content from modelConfigFile
		if s.args.ModelName != "" {
			return fmt.Errorf("modelConfigFile=%s is specified, so --model-name cannot be used", s.args.ModelConfigFile)
		}
		if s.args.ModelPath != "" {
			return fmt.Errorf("modelConfigFile=%s is specified, so --model-path cannot be used", s.args.ModelConfigFile)
		}

		modelConfigFileContentBytes, err := ioutil.ReadFile(s.args.ModelConfigFile)
		if err != nil {
			return fmt.Errorf("cannot read the modelConfigFile[%s]: %s", s.args.ModelConfigFile, err)
		}
		modelConfigString := string(modelConfigFileContentBytes)
		log.Debugf("The content of modelConfigFile[%s] is: %s", s.args.ModelConfigFile, modelConfigString)
		s.args.ModelConfigFileContent = modelConfigString
	}
	return nil
}

func (s *TensorflowServingArgsBuilder) checkServiceExists() error {
	client := config.GetArenaConfiger().GetClientSet()
	_, err := client.CoreV1().Services(s.args.Namespace).Get(context.TODO(), s.args.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		s.args.ModelServiceExists = false
	} else {
		s.args.ModelServiceExists = true
	}
	return nil
}

func (s *TensorflowServingArgsBuilder) validateVersionPolicy() error {
	// validate version policy
	if s.args.VersionPolicy == "" {
		s.args.VersionPolicy = "latest"
	}
	versionPolicyName := strings.Split(s.args.VersionPolicy, ":")
	switch versionPolicyName[0] {
	case "latest", "specific", "all":
		log.Debug("Support TensorFlow Serving Version Policy: latest, specific, all.")
	default:
		return fmt.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
	}
	return nil
}

func (s *TensorflowServingArgsBuilder) generateModelConfigFileContent() string {
	modelName := s.args.ModelName
	versionPolicy := s.args.VersionPolicy
	mountPath := s.args.ModelPath
	versionPolicyName := strings.Split(versionPolicy, ":")

	var buffer bytes.Buffer
	buffer.WriteString("model_config_list: { config: { name: ")
	buffer.WriteString("\"" + modelName + "\" base_path: \"")
	buffer.WriteString(mountPath + "\" model_platform: \"")
	buffer.WriteString("tensorflow\" model_version_policy: { ")
	switch versionPolicyName[0] {
	case "all":
		buffer.WriteString(versionPolicyName[0] + ": {} } } }")
	case "specific":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": { " + "versions: " + versionPolicyName[1] + " } } } }")
		} else {
			log.Errorf("[specific] version policy scheme should be specific:N")
		}
	case "latest":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": { " + "num_versions: " + versionPolicyName[1] + " } } } }")
		} else {
			buffer.WriteString(versionPolicyName[0] + ": { " + "num_versions: 1 } } } }")
		}
	default:
		log.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
		buffer.Reset()
	}

	result := buffer.String()
	log.Debugf("generateModelConfigFileContent: \n%s", result)
	return result
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
