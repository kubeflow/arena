// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ModelEvaluateArgsBuilder struct {
	args        *types.ModelEvaluateArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewModelEvaluateArgsBuilder(args *types.ModelEvaluateArgs) ArgsBuilder {
	args.Type = types.ModelEvaluateJob
	m := &ModelEvaluateArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	m.AddSubBuilder(
		NewModelArgsBuilder(&m.args.CommonModelArgs),
		NewSubmitSyncCodeArgsBuilder(&m.args.SubmitSyncCodeArgs),
	)
	m.AddArgValue("default-image", DefaultModelJobImage)
	return m
}

func (m *ModelEvaluateArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*m)), ".")
	return items[len(items)-1]
}

func (m *ModelEvaluateArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		m.subBuilders[b.GetName()] = b
	}
	return m
}

func (m *ModelEvaluateArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range m.subBuilders {
		m.subBuilders[name].AddArgValue(key, value)
	}
	m.argValues[key] = value
	return m
}

func (m *ModelEvaluateArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range m.subBuilders {
		m.subBuilders[name].AddCommandFlags(command)
	}

	var (
		imagePullSecrets []string
	)

	command.Flags().StringVar(&m.args.ModelPlatform, "model-platform", "", "model platform")
	command.Flags().StringVar(&m.args.DatasetPath, "dataset-path", "", "evaluate dataset path")
	command.Flags().IntVar(&m.args.BatchSize, "batch-size", 1, "evaluate dataset path")
	command.Flags().StringVar(&m.args.ReportPath, "report-path", "", "evaluate result path")

	command.Flags().StringArrayVar(&imagePullSecrets, "image-pull-secret", []string{}, `giving names of imagePullSecret when you want to use a private registry, usage:"--image-pull-secret <name1>"`)

	m.AddArgValue("image-pull-secret", &imagePullSecrets)
}

func (m *ModelEvaluateArgsBuilder) PreBuild() error {
	for name := range m.subBuilders {
		if err := m.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelEvaluateArgsBuilder) Build() error {
	for name := range m.subBuilders {
		if err := m.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := m.preprocess(); err != nil {
		return err
	}
	return nil
}

func (m *ModelEvaluateArgsBuilder) preprocess() (err error) {
	log.Debugf("command: %s", m.args.Command)
	if m.args.Image == "" {
		return fmt.Errorf("image must be specified")
	}

	return nil
}
