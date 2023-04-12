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
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type KFServingArgsBuilder struct {
	args        *types.KFServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewKFServingArgsBuilder(args *types.KFServingArgs) ArgsBuilder {
	args.Type = types.KFServingJob
	s := &KFServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewServingArgsBuilder(&s.args.CommonServingArgs),
	)
	return s
}

func (s *KFServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *KFServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *KFServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *KFServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVar(&s.args.ModelType, "model-type", "custom", "the type of serving model,default to custom type")
	command.Flags().StringVar(&s.args.StorageUri, "storage-uri", "", "the uri direct to the model file")
	command.Flags().UintVar(&s.args.CanaryPercent, "canary-percent", 0, "the percent of the desired canary")
	command.Flags().UintVar(&s.args.Port, "port", 0, "the port of the application listens in the custom image")
	command.Flags().UintVar(&s.args.MinReplicas, "min-replicas", 1, "the minimal replica for the server")
}

func (s *KFServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *KFServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.validate(); err != nil {
		return err
	}
	return nil
}

func (s *KFServingArgsBuilder) validate() (err error) {
	if s.args.StorageUri == "" && s.args.Image == "" {
		return fmt.Errorf("storage uri and image can not be empty at the same time.")
	}
	return nil
}
