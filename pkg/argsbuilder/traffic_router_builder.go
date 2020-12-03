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
	"regexp"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
)

type TrafficRouterArgsBuilder struct {
	args        *types.TrafficRouterSplitArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewTrafficRouterArgsBuilder(args *types.TrafficRouterSplitArgs) ArgsBuilder {
	s := &TrafficRouterArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	return s
}

func (s *TrafficRouterArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *TrafficRouterArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *TrafficRouterArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *TrafficRouterArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		versions []string
	)
	command.Flags().StringVar(&s.args.ServingName, "name", "", "the serving name")
	command.Flags().StringSliceVarP(&versions, "version-weight", "v", []string{}, "set the version and weight,format is: version:weight, e.g. --version-weight version1:20 --version-weight version2:40")
	//command.Flags().StringVar(&s.args.Versions, "versions", "", "Model versions which the traffic will be routed to, e.g. 1,2,3")
	//command.Flags().StringVar(&s.args.Weights, "weights", "", "Weight percentage values for each model version which the traffic will be routed to,e.g. 70,20,10")
	command.MarkFlagRequired("name")
	command.MarkFlagRequired("version-weight")
	s.AddArgValue("version-weight", &versions)
}

func (s *TrafficRouterArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *TrafficRouterArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.checkModelName(); err != nil {
		return err
	}
	if err := s.setVersionWeights(); err != nil {
		return err
	}
	return nil
}

func (s *TrafficRouterArgsBuilder) checkModelName() error {
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.args.ServingName)
	if !matched {
		return fmt.Errorf("parameter model name should be numbers, letters, dashes, and underscores ONLY")
	}
	return nil
}

func (s *TrafficRouterArgsBuilder) setVersionWeights() error {
	s.args.VersionWeights = []types.ServingVersionWeight{}
	obj, ok := s.argValues["version-weight"]
	if !ok {
		return fmt.Errorf("versions and weights must be set,use '--version-weight' or '-v' to set")
	}
	total := 0
	versions := obj.(*[]string)
	for _, vw := range *versions {
		item := strings.Split(vw, ":")
		versionWeight := types.ServingVersionWeight{
			Version: item[0],
			Weight:  100,
		}
		if len(item) >= 2 {
			weight, err := strconv.Atoi(item[1])
			if err != nil {
				return fmt.Errorf("invalid format for version and weight,should like version:weight and weight must be int")
			}
			versionWeight.Weight = weight
		}
		total += versionWeight.Weight
		s.args.VersionWeights = append(s.args.VersionWeights, versionWeight)
	}
	if total != 100 {
		return fmt.Errorf("invalid version weight format,total of all version weights is not equal to 100")
	}
	return nil
}
