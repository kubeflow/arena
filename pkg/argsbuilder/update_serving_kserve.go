// Copyright 2023 The Kubeflow Authors
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

type UpdateKServeArgsBuilder struct {
	args        *types.UpdateKServeArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewUpdateKServeArgsBuilder(args *types.UpdateKServeArgs) ArgsBuilder {
	args.Type = types.KServeJob
	s := &UpdateKServeArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewUpdateServingArgsBuilder(&s.args.CommonUpdateServingArgs),
	)
	return s
}

func (s *UpdateKServeArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *UpdateKServeArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *UpdateKServeArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *UpdateKServeArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}

	var (
		modelFormat string
	)

	command.Flags().StringVar(&modelFormat, "model-format", "", `the ModelFormat being served. usage: "--model-format=name" or "--model-format=name:version"`)
	command.Flags().StringVar(&s.args.Runtime, "runtime", "", "the ClusterServingRuntime or ServingRuntime name to use for deployment.")
	command.Flags().StringVar(&s.args.StorageUri, "storage-uri", "", "the uri direct to the model file")
	command.Flags().IntVar(&s.args.Port, "port", 0, "the port of tcp listening port, default is 8080 in kserve")
	command.Flags().StringVar(&s.args.RuntimeVersion, "runtime-version", "", "the predictor docker image")
	command.Flags().StringVar(&s.args.ProtocolVersion, "protocol-version", "", "the protocol version to use by the predictor (i.e. v1 or v2 or grpc-v1 or grpc-v2)")

	// ComponentExtension defines the deployment configuration for a given InferenceService component
	command.Flags().IntVar(&s.args.MinReplicas, "min-replicas", 1, "minimum number of replicas, defaults to 1 but can be set to 0 to enable scale-to-zero")
	command.Flags().IntVar(&s.args.MaxReplicas, "max-replicas", 0, "maximum number of replicas for autoscaling")
	command.Flags().IntVar(&s.args.ScaleTarget, "scale-target", 0, "specifies the integer target value of the metric type the Autoscaler watches for")
	command.Flags().StringVar(&s.args.ScaleMetric, "scale-metric", "", "the scaling metric type watched by autoscaler. possible values are concurrency, rps, cpu, memory. concurrency, rps are supported via KPA")
	command.Flags().Int64Var(&s.args.ContainerConcurrency, "container-concurrency", 0, "the requests can be processed concurrently, this sets the hard limit of the container concurrency")
	command.Flags().Int64Var(&s.args.TimeoutSeconds, "timeout", 0, "the number of seconds to wait before timing out a request to the component.")
	command.Flags().Int64Var(&s.args.CanaryTrafficPercent, "canary-traffic-percent", -1, "the traffic split percentage between the candidate revision and the last ready revision")

	s.AddArgValue(KServeModelFormat, &modelFormat)
}

func (s *UpdateKServeArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}

	return nil
}

func (s *UpdateKServeArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.setModelFormat(); err != nil {
		return err
	}

	return nil
}

func (s *UpdateKServeArgsBuilder) setModelFormat() error {
	item, ok := s.argValues[KServeModelFormat]
	if !ok {
		return nil
	}

	modelFormat := item.(*string)
	if *modelFormat == "" {
		return nil
	}

	mfs := strings.Split(*modelFormat, ":")
	if len(mfs) > 2 {
		return fmt.Errorf("model format is invalid: %s", *modelFormat)
	}
	if len(mfs) == 1 {
		s.args.ModelFormat = &types.ModelFormat{
			Name: mfs[0],
		}
	}
	if len(mfs) == 2 {
		s.args.ModelFormat = &types.ModelFormat{
			Name:    mfs[0],
			Version: &mfs[1],
		}
	}
	return nil
}
