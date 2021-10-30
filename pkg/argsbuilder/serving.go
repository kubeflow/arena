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
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServingArgsBuilder struct {
	args        *types.CommonServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewServingArgsBuilder(args *types.CommonServingArgs) ArgsBuilder {
	s := &ServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	return s
}

func (s *ServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *ServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *ServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *ServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
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
	defaultImage := ""
	item, ok := s.argValues["default-image"]
	if ok {
		defaultImage = item.(string)
	}
	command.Flags().StringVar(&s.args.Image, "image", defaultImage, "the docker image name of serving job")
	command.Flags().StringVar(&s.args.ImagePullPolicy, "imagePullPolicy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")
	command.Flags().MarkDeprecated("imagePullPolicy", "please use --image-pull-policy instead")
	command.Flags().StringVar(&s.args.ImagePullPolicy, "image-pull-policy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")

	command.Flags().IntVar(&s.args.GPUCount, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().IntVar(&s.args.GPUMemory, "gpumemory", 0, "the limit GPU memory of each replica to run the serve.")
	command.Flags().StringVar(&s.args.Cpu, "cpu", "", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&s.args.Memory, "memory", "", "the request memory of each replica to run the serve.")
	command.Flags().IntVar(&s.args.Replicas, "replicas", 1, "the replicas number of the serve job.")

	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")

	command.Flags().BoolVar(&s.args.EnableIstio, "enableIstio", false, "enable Istio for serving or not (disable Istio by default)")
	command.Flags().MarkDeprecated("enableIstio", "please use --enable-istio instead")
	command.Flags().BoolVar(&s.args.EnableIstio, "enable-istio", false, "enable Istio for serving or not (disable Istio by default)")

	command.Flags().BoolVar(&s.args.ExposeService, "exposeService", false, "expose service using Istio gateway for external access or not (not expose by default)")
	command.Flags().MarkDeprecated("exposeService", "please use --expose-service instead")
	command.Flags().BoolVar(&s.args.ExposeService, "expose-service", false, "expose service using Istio gateway for external access or not (not expose by default)")

	command.Flags().StringVar(&s.args.Name, "servingName", "", "the serving name")
	command.Flags().MarkDeprecated("servingName", "please use --name instead")
	command.Flags().StringVar(&s.args.Name, "name", "", "the serving name")

	command.Flags().StringVar(&s.args.Version, "servingVersion", "", "the serving version")
	command.Flags().MarkDeprecated("servingVersion", "please use --version instead")
	command.Flags().StringVar(&s.args.Version, "version", "", "the serving version")

	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>")
	command.Flags().StringArrayVarP(&datadir, "data-dir", "", []string{}, "specify the trained models datasource on host to mount for serving, like <host_path>:<mount_point_on_job>")
	command.MarkFlagRequired("name")

	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, "specify the annotations")
	command.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "specify the labels")
	command.Flags().StringArrayVarP(&tolerations, "toleration", "", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration taint-key" or "--toleration all" `)
	command.Flags().StringArrayVarP(&selectors, "selector", "", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)

	s.AddArgValue("annotation", &annotations).
		AddArgValue("toleration", &tolerations).
		AddArgValue("label", &labels).
		AddArgValue("selector", &selectors).
		AddArgValue("data", &dataset).
		AddArgValue("data-dir", &datadir).
		AddArgValue("env", &envs)
}

func (s *ServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	if err := s.checkNamespace(); err != nil {
		return err
	}
	if err := s.validateIstioEnablement(); err != nil {
		return err
	}
	if err := s.setDataSet(); err != nil {
		return err
	}
	if err := s.setDataDirs(); err != nil {
		return err
	}
	if err := s.setEnvs(); err != nil {
		return err
	}
	if err := s.setAnnotations(); err != nil {
		return err
	}
	if err := s.setLabels(); err != nil {
		return err
	}
	if err := s.setNodeSelectors(); err != nil {
		return err
	}
	if err := s.setTolerations(); err != nil {
		return err
	}
	if err := s.setServingVersion(); err != nil {
		return err
	}
	if err := s.setUserNameAndUserId(); err != nil {
		return err
	}
	if err := s.disabledNvidiaENVWithNoneGPURequest(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	return nil
}

func (s *ServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.check(); err != nil {
		return err
	}
	return nil
}

func (s *ServingArgsBuilder) check() error {
	return s.checkServiceExists()
}

// setDataSets is used to handle option --data
func (s *ServingArgsBuilder) setDataSet() error {
	s.args.ModelDirs = map[string]string{}
	argKey := "data"
	var dataSet *[]string
	value, ok := s.argValues[argKey]
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
	s.args.ModelDirs = transformSliceToMap(*dataSet, ":")
	return nil
}

// setDataDirs is used to handle option --data-dir
func (s *ServingArgsBuilder) setDataDirs() error {
	s.args.HostVolumes = []types.DataDirVolume{}
	argKey := "data-dir"
	var dataDirs *[]string
	value, ok := s.argValues[argKey]
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
		s.args.HostVolumes = append(s.args.HostVolumes, types.DataDirVolume{
			Name:          fmt.Sprintf("serving-data-%d", i),
			HostPath:      hostPath,
			ContainerPath: containerPath,
		})
	}
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *ServingArgsBuilder) setAnnotations() error {
	s.args.Annotations = map[string]string{}
	argKey := "annotation"
	var annotations *[]string
	item, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	annotations = item.(*[]string)
	if len(*annotations) <= 0 {
		return nil
	}
	s.args.Annotations = transformSliceToMap(*annotations, "=")
	return nil
}

// setLabels is used to handle option --label
func (s *ServingArgsBuilder) setLabels() error {
	s.args.Labels = map[string]string{}
	argKey := "label"
	var labels *[]string
	item, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	labels = item.(*[]string)
	if len(*labels) <= 0 {
		return nil
	}
	s.args.Labels = transformSliceToMap(*labels, "=")
	return nil
}

func (s *ServingArgsBuilder) setUserNameAndUserId() error {
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	arenaConfiger := config.GetArenaConfiger()
	user := arenaConfiger.GetUser()
	s.args.Labels[types.UserNameIdLabel] = user.GetId()
	s.args.Annotations[types.UserNameNameLabel] = user.GetName()
	return nil
}

// setNodeSelectors is used to handle option --selector
func (s *ServingArgsBuilder) setNodeSelectors() error {
	s.args.NodeSelectors = map[string]string{}
	argKey := "selector"
	var nodeSelectors *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	nodeSelectors = value.(*[]string)
	log.Debugf("node selectors: %v", *nodeSelectors)
	s.args.NodeSelectors = transformSliceToMap(*nodeSelectors, "=")
	return nil
}

// setTolerations is used to handle option --toleration
func (s *ServingArgsBuilder) setTolerations() error {
	s.args.Tolerations = []string{}
	argKey := "toleration"
	var tolerations *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	tolerations = value.(*[]string)
	log.Debugf("tolerations: %v", *tolerations)
	for _, taintKey := range *tolerations {
		if taintKey == "all" {
			s.args.Tolerations = []string{"all"}
			return nil
		}
		s.args.Tolerations = append(s.args.Tolerations, taintKey)
	}
	return nil
}

func (s *ServingArgsBuilder) setEnvs() error {
	argKey := "env"
	var envs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	envs = value.(*[]string)
	s.args.Envs = transformSliceToMap(*envs, "=")
	return nil
}

func (s *ServingArgsBuilder) setServingVersion() error {
	if s.args.Version == "" {
		t := time.Now()
		s.args.Version = fmt.Sprint(t.Format("200601021504"))
	}
	return nil
}

func (s *ServingArgsBuilder) validateIstioEnablement() error {
	log.Debugf("--enable-Istio=%t is specified.", s.args.EnableIstio)
	if !s.args.EnableIstio {
		return nil
	}
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.args.Name)
	if !matched {
		return fmt.Errorf("--name should be numbers, letters, dashes, and underscores ONLY")
	}
	log.Debugf("--version=%s is specified.", s.args.Version)
	if s.args.Version == "" {
		return fmt.Errorf("--version must be specified if --enable-istio=true")
	}
	return nil
}

// checkServiceExists is used to check services,must execute after function checkNamespace
func (s *ServingArgsBuilder) checkServiceExists() error {
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

func (s *ServingArgsBuilder) checkNamespace() error {
	if s.args.Namespace == "" {
		return fmt.Errorf("not set namespace,please set it")
	}
	log.Debugf("namespace is %v", s.args.Namespace)
	return nil
}

func (s *ServingArgsBuilder) disabledNvidiaENVWithNoneGPURequest() error {
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}
	if s.args.GPUCount == 0 && s.args.GPUMemory == 0 {
		s.args.Envs["NVIDIA_VISIBLE_DEVICES"] = "void"
	}
	return nil
}
