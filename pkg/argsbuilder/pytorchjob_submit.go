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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type SubmitPytorchJobArgsBuilder struct {
	args        *types.SubmitPyTorchJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitPytorchJobArgsBuilder(args *types.SubmitPyTorchJobArgs) ArgsBuilder {
	s := &SubmitPytorchJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewSubmitArgsBuilder(&s.args.CommonSubmitArgs),
		NewSubmitSyncCodeArgsBuilder(&s.args.SubmitSyncCodeArgs),
		NewSubmitTensorboardArgsBuilder(&s.args.SubmitTensorboardArgs),
	)
	return s
}

func (s *SubmitPytorchJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitPytorchJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitPytorchJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitPytorchJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVar(&s.args.CleanPodPolicy, "clean-task-policy", "None", "How to clean tasks after Training is done, support None, Running, All.")
}

func (s *SubmitPytorchJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	s.AddArgValue(ShareDataPrefix+"dataset", s.args.DataSet)
	return nil
}

func (s *SubmitPytorchJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.addPodGroupLabel(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitPytorchJobArgsBuilder) check() error {
	if s.args.Image == "" {
		return fmt.Errorf("--image must be set ")
	}
	// check clean-task-policy
	switch s.args.CleanPodPolicy {
	case "None", "Running", "All":
		log.Debugf("Supported cleanTaskPolicy: %s", s.args.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", s.args.CleanPodPolicy)
	}
	return nil
}

func (s *SubmitPytorchJobArgsBuilder) addPodGroupLabel() error {
	s.args.CommonSubmitArgs.PodGroupName = s.args.CommonSubmitArgs.Name
	s.args.CommonSubmitArgs.PodGroupMinAvailable = fmt.Sprintf("%v", s.args.CommonSubmitArgs.WorkerCount)
	return nil
}
