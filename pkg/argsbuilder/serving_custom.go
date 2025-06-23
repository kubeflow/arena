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
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	LivenessProbeActionOption  = "livenessProbeActionOption"
	ReadinessProbeActionOption = "readinessProbeActionOption"
	StartupProbeActionOption   = "StartupProbeActionOption"
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

	var (
		livenessProbeActionOption  []string
		readinessProbeActionOption []string
		startupProbeActionOption   []string
	)

	command.Flags().IntVar(&s.args.Port, "port", 0, "the port of gRPC listening port, default is 0 represents that don't create service listening on this port")
	command.Flags().IntVar(&s.args.RestfulPort, "restful-port", 0, "the port of RESTful listening port, default is 0 represents that don't create service listening on this port")
	command.Flags().IntVar(&s.args.MetricsPort, "metrics-port", 0, "the port of metrics, default is 0 represents that don't create service listening on this port")
	command.Flags().StringVar(&s.args.MaxSurge, "max-surge", "", "the maximum number of pods that can be created over the desired number of pods")
	command.Flags().StringVar(&s.args.MaxUnavailable, "max-unavailable", "", "the maximum number of Pods that can be unavailable during the update process")
	command.Flags().StringVar(&s.args.LivenessProbeAction, "liveness-probe-action", "", "the liveness probe action, support httpGet,exec,grpc,tcpSocket")
	command.Flags().StringArrayVar(&livenessProbeActionOption, "liveness-probe-action-option", []string{}, `the liveness probe action option, usage: --liveness-probe-action-option="path: /healthz" or --liveness-probe-action-option="command=cat /tmp/healthy"`)
	command.Flags().StringArrayVar(&s.args.LivenessProbeOption, "liveness-probe-option", []string{}, `the liveness probe option, usage: --liveness-probe-option="initialDelaySeconds: 3" or --liveness-probe-option="periodSeconds: 3"`)
	command.Flags().StringVar(&s.args.ReadinessProbeAction, "readiness-probe-action", "", "the readiness probe action, support httpGet,exec,grpc,tcpSocket")
	command.Flags().StringArrayVar(&readinessProbeActionOption, "readiness-probe-action-option", []string{}, `the readiness probe action option, usage: --readiness-probe-action-option="path: /healthz" or --readiness-probe-action-option="command=cat /tmp/healthy"`)
	command.Flags().StringArrayVar(&s.args.ReadinessProbeOption, "readiness-probe-option", []string{}, `the readiness probe option, usage: --readiness-probe-option="initialDelaySeconds: 3" or --readiness-probe-option="periodSeconds: 3"`)
	command.Flags().StringVar(&s.args.StartupProbeAction, "startup-probe-action", "", "the startup probe action, support httpGet,exec,grpc,tcpSocket")
	command.Flags().StringArrayVar(&startupProbeActionOption, "startup-probe-action-option", []string{}, `the startup probe action option, usage: --startup-probe-action-option="path: /healthz" or --startup-probe-action-option="command=cat /tmp/healthy"`)
	command.Flags().StringArrayVar(&s.args.StartupProbeOption, "startup-probe-option", []string{}, `the startup probe option, usage: --startup-probe-option="initialDelaySeconds: 3" or --startup-probe-option="periodSeconds: 3"`)

	s.AddArgValue(LivenessProbeActionOption, &livenessProbeActionOption).
		AddArgValue(ReadinessProbeActionOption, &readinessProbeActionOption).
		AddArgValue(StartupProbeActionOption, &startupProbeActionOption)
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
	if err := s.setLivenessProbeActionOption(); err != nil {
		return err
	}
	if err := s.setReadinessProbeActionOption(); err != nil {
		return err
	}
	if err := s.setStartupProbeActionOption(); err != nil {
		return err
	}
	return nil
}

func (s *CustomServingArgsBuilder) check() error {
	if s.args.Image == "" {
		return fmt.Errorf("image must be specified")
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
	return fmt.Errorf("all  ports are 0,invalid configuration")
}

func (s *CustomServingArgsBuilder) setLivenessProbeActionOption() error {
	value, ok := s.argValues[LivenessProbeActionOption]
	if !ok {
		return nil
	}
	livenessProbeActionOptions := value.(*[]string)
	log.Debugf("livenessProbeActionOptions: %v", *livenessProbeActionOptions)
	for _, option := range *livenessProbeActionOptions {
		if strings.HasPrefix(option, "command=") {
			s.args.LivenessProbeActionOption = append(s.args.LivenessProbeActionOption, "command:")
			commands := strings.Split(option[8:], ` `)
			for _, command := range commands {
				cmd := fmt.Sprintf("- %s", command)
				s.args.LivenessProbeActionOption = append(s.args.LivenessProbeActionOption, cmd)
			}
		} else {
			s.args.LivenessProbeActionOption = append(s.args.LivenessProbeActionOption, option)
		}
	}

	return nil
}

func (s *CustomServingArgsBuilder) setReadinessProbeActionOption() error {
	value, ok := s.argValues[ReadinessProbeActionOption]
	if !ok {
		return nil
	}
	readinessProbeActionOptions := value.(*[]string)
	log.Debugf("readinessProbeActionOptions: %v", *readinessProbeActionOptions)
	for _, option := range *readinessProbeActionOptions {
		if strings.HasPrefix(option, "command=") {
			s.args.ReadinessProbeActionOption = append(s.args.ReadinessProbeActionOption, "command:")
			commands := strings.Split(option[8:], ` `)
			for _, command := range commands {
				cmd := fmt.Sprintf("- %s", command)
				s.args.ReadinessProbeActionOption = append(s.args.ReadinessProbeActionOption, cmd)
			}
		} else {
			s.args.ReadinessProbeActionOption = append(s.args.ReadinessProbeActionOption, option)
		}
	}

	return nil
}

func (s *CustomServingArgsBuilder) setStartupProbeActionOption() error {
	value, ok := s.argValues[StartupProbeActionOption]
	if !ok {
		return nil
	}
	startupProbeActionOptions := value.(*[]string)
	log.Debugf("startupProbeActionOptions: %v", *startupProbeActionOptions)
	for _, lpo := range *startupProbeActionOptions {
		if strings.HasPrefix(lpo, "command=") {
			s.args.StartupProbeActionOption = append(s.args.StartupProbeActionOption, "command:")
			commands := strings.Split(lpo[8:], ` `)
			for _, command := range commands {
				cmd := fmt.Sprintf("- %s", command)
				s.args.StartupProbeActionOption = append(s.args.StartupProbeActionOption, cmd)
			}
		} else {
			s.args.StartupProbeActionOption = append(s.args.StartupProbeActionOption, lpo)
		}
	}

	return nil
}
