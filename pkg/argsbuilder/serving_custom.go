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

type CustomServingArgsBuilder struct {
	args        *types.CustomServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewCustomServingArgsBuilder(args *types.CustomServingArgs) ArgsBuilder {
	args.Type = types.CustomServingJob
	s := &CustomServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewServingArgsBuilder(&s.args.CommonServingArgs),
	)
	return s
}

func (s *CustomServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *CustomServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *CustomServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *CustomServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().IntVar(&s.args.Port, "port", 0, "the port of gRPC listening port,default is 0 represents that don't create service listening on this port")
	command.Flags().IntVar(&s.args.RestfulPort, "restful-port", 0, "the port of RESTful listening port,default is 0 represents that don't create service listening on this port")
}

func (s *CustomServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *CustomServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.checkPortsIsOk(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	return nil
}

func (s *CustomServingArgsBuilder) check() error {
	if s.args.Image == "" {
		return fmt.Errorf("image must be specified.")
	}
	return nil
}

func (s *CustomServingArgsBuilder) checkPortsIsOk() error {
	switch {
	case s.args.Port != 0:
		return nil
	case s.args.RestfulPort != 0:
		return nil
	}
	return fmt.Errorf("all  ports are 0,invalid configuration.")
}
