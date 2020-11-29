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
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type ScaleETJobArgsBuilder struct {
	args        *types.ScaleETJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewScaleETJobArgsBuilder(args *types.ScaleETJobArgs) ArgsBuilder {
	s := &ScaleETJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	return s
}

func (s *ScaleETJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *ScaleETJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *ScaleETJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *ScaleETJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		scaleDuration time.Duration
		scaleinEnvs   []string
		script        string
	)
	if s, ok := s.argValues["script"]; ok {
		script = s.(string)
	}
	command.Flags().StringVar(&s.args.Name, "name", "", "required, et job name")
	command.MarkFlagRequired("name")
	command.Flags().DurationVarP(&scaleDuration, "timeout", "t", 60*time.Second, "timeout of callback scaler script, like 5s, 2m, or 3h.")
	command.Flags().IntVar(&s.args.Retry, "retry", 0, "retry times.")
	command.Flags().IntVar(&s.args.Count, "count", 1, "the nums of you want to add or delete worker.")
	command.Flags().StringVar(&s.args.Script, "script", script, "script of scaling.")
	command.Flags().StringArrayVarP(&scaleinEnvs, "env", "e", []string{}, "the environment variables.")
	s.AddArgValue("env", &scaleinEnvs).AddArgValue("timeout", &scaleDuration)
}

func (s *ScaleETJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	if err := s.transferEnvs(); err != nil {
		return err
	}
	if err := s.transferTimeout(); err != nil {
		return err
	}
	return nil
}

func (s *ScaleETJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	return nil
}

func (s *ScaleETJobArgsBuilder) transferEnvs() error {
	// transfer envs
	s.args.Envs = map[string]string{}
	item, ok := s.argValues["env"]
	if !ok {
		return nil
	}
	envs := item.(*[]string)
	s.args.Envs = transformSliceToMap(*envs, "=")
	return nil
}

func (s *ScaleETJobArgsBuilder) transferTimeout() error {
	// transfer timeout
	s.args.Timeout = 0
	item, ok := s.argValues["timeout"]
	if !ok {
		return nil
	}
	timeout := item.(*time.Duration)
	s.args.Timeout = int((*timeout).Seconds())
	return nil
}
