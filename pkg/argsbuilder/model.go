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
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"strings"
)

const (
	DefaultModelJobImage = "tensorflow/serving:latest"
)

type ModelArgsBuilder struct {
	args        *types.CommonModelArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewModelArgsBuilder(args *types.CommonModelArgs) ArgsBuilder {
	return &ModelArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
}

func (m *ModelArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*m)), ".")
	return items[len(items)-1]
}

func (m *ModelArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		m.subBuilders[b.GetName()] = b
	}
	return m
}

func (m *ModelArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range m.subBuilders {
		m.subBuilders[name].AddArgValue(key, value)
	}
	m.argValues[key] = value
	return m
}

func (m *ModelArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range m.subBuilders {
		m.subBuilders[name].AddCommandFlags(command)
	}

	var (
		envs        []string
		dataset     []string
		datadir     []string
		annotations []string
		tolerations []string
		labels      []string
		selectors   []string
	)

	command.Flags().StringVar(&m.args.Name, "name", "", "the serving name")

	command.Flags().StringVar(&m.args.ModelConfigFile, "model-config-file", "", "model config file")
	command.Flags().StringVar(&m.args.ModelName, "model-name", "", "model name")
	command.Flags().StringVar(&m.args.ModelPath, "model-path", "", "model path")
	command.Flags().StringVar(&m.args.Inputs, "inputs", "", "model inputs")
	command.Flags().StringVar(&m.args.Outputs, "outputs", "", "model outputs")

	command.Flags().StringVar(&m.args.Shell, "shell", "sh", "linux shell type, bash or sh")
	command.Flags().StringVar(&m.args.Command, "command", "", "the command will inject to container's command.")

	command.Flags().StringVar(&m.args.Image, "image", "", "the docker image name of model job")
	command.Flags().StringVar(&m.args.ImagePullPolicy, "image-pull-policy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")

	command.Flags().IntVar(&m.args.GPUCount, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().IntVar(&m.args.GPUMemory, "gpumemory", 0, "the limit GPU memory of each replica to run the serve.")
	command.Flags().IntVar(&m.args.GPUCore, "gpucore", 0, "the limit GPU core of each replica to run the serve.")
	command.Flags().StringVar(&m.args.Cpu, "cpu", "", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&m.args.Memory, "memory", "", "the request memory of each replica to run the serve.")

	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>")
	command.Flags().StringArrayVarP(&datadir, "data-dir", "", []string{}, "specify the trained models datasource on host to mount for serving, like <host_path>:<mount_point_on_job>")
	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")
	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, `specify the annotations, usage: "--annotation=key=value" or "--annotation key=value"`)
	command.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "specify the labels")
	command.Flags().StringArrayVarP(&tolerations, "toleration", "", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration key=value:effect,operator" or "--toleration all" `)
	command.Flags().StringArrayVarP(&selectors, "selector", "", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)

	m.AddArgValue("annotation", &annotations).
		AddArgValue("toleration", &tolerations).
		AddArgValue("label", &labels).
		AddArgValue("selector", &selectors).
		AddArgValue("data", &dataset).
		AddArgValue("data-dir", &datadir).
		AddArgValue("env", &envs)
}

func (m *ModelArgsBuilder) PreBuild() error {
	for name := range m.subBuilders {
		if err := m.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}

	return nil
}

// Build builds the common submit args
func (m *ModelArgsBuilder) Build() error {
	for name := range m.subBuilders {
		if err := m.subBuilders[name].Build(); err != nil {
			return err
		}
	}

	// set data
	if err := m.setDataSet(); err != nil {
		return err
	}
	// set data dir
	if err := m.setDataDirs(); err != nil {
		return err
	}
	// set annotation
	if err := m.setAnnotations(); err != nil {
		return err
	}
	// set label
	if err := m.setLabels(); err != nil {
		return err
	}
	// set environment
	if err := m.setEnvs(); err != nil {
		return err
	}
	// set node selectors
	if err := m.setNodeSelectors(); err != nil {
		return err
	}
	// set toleration
	if err := m.setTolerations(); err != nil {
		return err
	}

	if err := m.preprocess(); err != nil {
		return err
	}

	if err := m.checkGPUCore(); err != nil {
		return err
	}

	// check resource
	if err := m.check(); err != nil {
		return err
	}
	return nil
}

func (m *ModelArgsBuilder) check() error {

	if m.args.GPUCount < 0 {
		return fmt.Errorf("--gpus is invalid")
	}
	if m.args.GPUMemory < 0 {
		return fmt.Errorf("--gpumemory is invalid")
	}
	if m.args.Cpu != "" {
		_, err := resource.ParseQuantity(m.args.Cpu)
		if err != nil {
			return fmt.Errorf("--cpu is invalid")
		}
	}
	if m.args.Memory != "" {
		_, err := resource.ParseQuantity(m.args.Memory)
		if err != nil {
			return fmt.Errorf("--memory is invalid")
		}
	}
	return nil
}

func (m *ModelArgsBuilder) setEnvs() error {
	argKey := "env"
	var envs *[]string
	value, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	envs = value.(*[]string)
	m.args.Envs = transformSliceToMap(*envs, "=")
	return nil
}

// setAnnotations is used to handle option --annotation
func (m *ModelArgsBuilder) setAnnotations() error {
	m.args.Annotations = map[string]string{}
	argKey := "annotation"
	var annotations *[]string
	item, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	annotations = item.(*[]string)
	if len(*annotations) <= 0 {
		return nil
	}
	m.args.Annotations = transformSliceToMap(*annotations, "=")
	return nil
}

// setLabels is used to handle option --label
func (m *ModelArgsBuilder) setLabels() error {
	m.args.Labels = map[string]string{}
	argKey := "label"
	var labels *[]string
	item, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	labels = item.(*[]string)
	if len(*labels) <= 0 {
		return nil
	}
	m.args.Labels = transformSliceToMap(*labels, "=")
	return nil
}

// setNodeSelectors is used to handle option --selector
func (m *ModelArgsBuilder) setNodeSelectors() error {
	m.args.NodeSelectors = map[string]string{}
	argKey := "selector"
	var nodeSelectors *[]string
	value, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	nodeSelectors = value.(*[]string)
	log.Debugf("node selectors: %v", *nodeSelectors)
	m.args.NodeSelectors = transformSliceToMap(*nodeSelectors, "=")
	return nil
}

// setTolerations is used to handle option --toleration
func (m *ModelArgsBuilder) setTolerations() error {
	if m.args.Tolerations == nil {
		m.args.Tolerations = []types.TolerationArgs{}
	}
	argKey := "toleration"
	var tolerations *[]string
	value, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	tolerations = value.(*[]string)
	log.Debugf("tolerations: %v", *tolerations)
	for _, taintKey := range *tolerations {
		if taintKey == "all" {
			m.args.Tolerations = append(m.args.Tolerations, types.TolerationArgs{
				Operator: "Exists",
			})
			return nil
		}
		tolerationArg, err := parseTolerationString(taintKey)
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		m.args.Tolerations = append(m.args.Tolerations, *tolerationArg)
	}
	return nil
}

// setDataDirs is used to handle option --data-dir
func (m *ModelArgsBuilder) setDataDirs() error {
	m.args.DataDirs = []types.DataDirVolume{}
	argKey := "data-dir"
	var dataDirs *[]string
	value, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	dataDirs = value.(*[]string)
	log.Debugf("dataDir: %v", *dataDirs)
	for i, dataDir := range *dataDirs {
		hostPath, containerPath, err := util.ParseDataDirRaw(dataDir)
		if err != nil {
			return err
		}
		m.args.DataDirs = append(m.args.DataDirs, types.DataDirVolume{
			Name:          fmt.Sprintf("training-data-%d", i),
			HostPath:      hostPath,
			ContainerPath: containerPath,
		})
	}
	return nil
}

// setDataSets is used to handle option --data
func (m *ModelArgsBuilder) setDataSet() error {
	m.args.DataSet = map[string]string{}
	argKey := "data"
	var dataSet *[]string
	value, ok := m.argValues[argKey]
	if !ok {
		return nil
	}
	dataSet = value.(*[]string)
	log.Debugf("dataset: %v", *dataSet)
	if len(*dataSet) <= 0 {
		return nil
	}
	err := util.ValidateDatasets(*dataSet)
	if err != nil {
		return err
	}
	m.args.DataSet = transformSliceToMap(*dataSet, ":")
	return nil
}

func (m *ModelArgsBuilder) preprocess() (err error) {
	log.Debugf("command: %s", m.args.Command)
	if m.args.Image == "" {
		return fmt.Errorf("image must be specified")
	}
	if m.args.ModelConfigFile == "" {
		// need to validate modelName, modelPath if not specify modelConfigFile
		//if m.args.ModelName == "" {
		//	return fmt.Errorf("model name must be specified")
		//}
		//if m.args.ModelPath == "" {
		//	return fmt.Errorf("model path must be specified")
		//}
		//if m.args.Inputs == "" {
		//	return fmt.Errorf("model inputs must be specified")
		//}
		//if m.args.Outputs == "" {
		//	return fmt.Errorf("model outputs must be specified")
		//}
	} else {
		//populate content from modelConfigFile
		if m.args.ModelName != "" {
			log.Infof("modelConfigFile=%s is specified, so --model-name will be ingored", m.args.ModelConfigFile)
		}
		if m.args.ModelPath != "" {
			log.Infof("modelConfigFile=%s is specified, so --model-path will be ignored", m.args.ModelConfigFile)
		}
		if m.args.Inputs != "" {
			log.Infof("modelConfigFile=%s is specified, so --inputs will be ignored", m.args.ModelConfigFile)
		}
		if m.args.Inputs != "" {
			log.Infof("modelConfigFile=%s is specified, so --outputs will be ignored", m.args.ModelConfigFile)
		}
	}
	return nil
}

func (m *ModelArgsBuilder) checkGPUCore() error {
	if m.args.GPUCore%5 != 0 {
		return fmt.Errorf("GPUCore should be the multiple of 5")
	}
	return nil
}
