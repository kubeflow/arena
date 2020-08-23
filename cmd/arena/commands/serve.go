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
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	validate "github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	modelPathSeparator = ":"
	regexp4serviceName = "^[a-z0-9A-Z_-]+$"
)

type ServeArgs struct {
	ImagePullPolicy string            `yaml:"imagePullPolicy"` // --imagePullPolicy
	GPUCount        int               `yaml:"gpuCount"`        // --gpus
	GPUMemory       int               `yaml:"gpuMemory"`       // --gpumemory
	Cpu             string            `yaml:"cpu"`             // --cpu
	Memory          string            `yaml:"memory"`          // --memory
	Envs            map[string]string `yaml:"envs"`            // --envs
	Command         string            `yaml:"command"`         // --command
	Replicas        int               `yaml:"replicas"`        // --replicas
	Port            int               `yaml:"port"`            // --port
	RestfulPort     int               `yaml:"restApiPort"`     // --restfulPort
	EnableIstio     bool              `yaml:"enableIstio"`     // --enableIstio
	ExposeService   bool              `yaml:"exposeService"`   // --exposeService
	ServingName     string            `yaml:"servingName"`     // --servingName
	ServingVersion  string            `yaml:"servingVersion"`  // --servingVersion
	ModelDirs       map[string]string `yaml:"modelDirs"`
	NodeSelectors   map[string]string `yaml:"nodeSelectors"` // --selector
	Tolerations     []string          `yaml:"tolerations"`   // --toleration
	Annotations     map[string]string `yaml:"annotations"`

	ModelServiceExists bool `yaml:"modelServiceExists"` // --modelServiceExists
}

func (s ServeArgs) validateIstioEnablement() error {
	log.Debugf("--enableIstio=%t is specified.", s.EnableIstio)
	if !s.EnableIstio {
		return nil
	}

	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.ServingName)
	if !matched {
		return fmt.Errorf("--serviceName should be numbers, letters, dashes, and underscores ONLY")
	}
	log.Debugf("--servingVersion=%s is specified.", s.ServingVersion)
	if s.ServingVersion == "" {
		return fmt.Errorf("--servingVersion must be specified if enableIstio=true")
	}

	return nil
}

// PreCheck gives some checking for args.
func (s ServeArgs) PreCheck() error {
	return s.checkPortsIsOk()
}

// check all ports is ok
func (s ServeArgs) checkPortsIsOk() error {
	switch {
	case s.Port != 0:
		return nil
	case s.RestfulPort != 0:
		return nil
	default:
		return fmt.Errorf("all  ports are 0,invalid configuration.")
	}
}

// get node selectors
func (s *ServeArgs) addNodeSelectors() {
	log.Debugf("node selectors: %v", selectors)
	if len(selectors) == 0 {
		s.NodeSelectors = map[string]string{}
		return
	}
	s.NodeSelectors = transformSliceToMap(selectors, "=")
}

// get tolerations labels
func (s *ServeArgs) addTolerations() {
	log.Debugf("tolerations: %v", tolerations)
	if len(tolerations) == 0 {
		s.Tolerations = []string{}
		return
	}
	s.Tolerations = []string{}
	for _, taintKey := range tolerations {
		if taintKey == "all" {
			s.Tolerations = []string{"all"}
			return
		}
		s.Tolerations = append(s.Tolerations, taintKey)
	}
}

// get annotations
func (s *ServeArgs) addAnnotations() {
	log.Debugf("annotations: %v", annotations)
	if len(annotations) > 0 {
		s.Annotations = transformSliceToMap(annotations, "=")
	}
}

func ParseMountPath(dataset []string) (err error) {
	err = validate.ValidateDatasets(dataset)
	return err
}

func (serveArgs *ServeArgs) addServeCommonFlags(command *cobra.Command) {

	// create subcommands
	command.Flags().StringVar(&serveArgs.ImagePullPolicy, "imagePullPolicy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")
	command.Flags().MarkDeprecated("imagePullPolicy", "please use --image-pull-policy instead")
	command.Flags().StringVar(&serveArgs.ImagePullPolicy, "image-pull-policy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")

	command.Flags().IntVar(&serveArgs.GPUCount, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().IntVar(&serveArgs.GPUMemory, "gpumemory", 0, "the limit GPU memory of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Cpu, "cpu", "", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Memory, "memory", "", "the request memory of each replica to run the serve.")
	command.Flags().IntVar(&serveArgs.Replicas, "replicas", 1, "the replicas number of the serve job.")

	command.Flags().StringArrayVarP(&envs, "envs", "e", []string{}, "the environment variables")

	command.Flags().BoolVar(&serveArgs.EnableIstio, "enableIstio", false, "enable Istio for serving or not (disable Istio by default)")
	command.Flags().MarkDeprecated("enableIstio", "please use --enable-istio instead")
	command.Flags().BoolVar(&serveArgs.EnableIstio, "enable-istio", false, "enable Istio for serving or not (disable Istio by default)")

	command.Flags().BoolVar(&serveArgs.ExposeService, "exposeService", false, "expose service using Istio gateway for external access or not (not expose by default)")
	command.Flags().MarkDeprecated("exposeService", "please use --expose-service instead")
	command.Flags().BoolVar(&serveArgs.ExposeService, "expose-service", false, "expose service using Istio gateway for external access or not (not expose by default)")

	command.Flags().StringVar(&serveArgs.ServingName, "servingName", "", "the serving name")
	command.Flags().MarkDeprecated("servingName", "please use --name instead")
	command.Flags().StringVar(&serveArgs.ServingName, "name", "", "the serving name")

	command.Flags().StringVar(&serveArgs.ServingVersion, "servingVersion", "", "the serving version")
	command.Flags().MarkDeprecated("servingVersion", "please use --version instead")
	command.Flags().StringVar(&serveArgs.ServingVersion, "version", "", "the serving version")

	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>")
	command.MarkFlagRequired("name")

	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, "the annotations")
	command.Flags().StringArrayVarP(&tolerations, "toleration", "", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration taint-key" or "--toleration all" `)
	command.Flags().StringArrayVarP(&selectors, "selector", "", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)
}

func init() {
	if os.Getenv(CHART_PKG_LOC) != "" {
		standalone_training_chart = filepath.Join(os.Getenv(CHART_PKG_LOC), "training")
	}
}

var (
	serveLong = `serve a job.

Available Commands:
  tensorflow,tf  Submit a TensorFlow Serving Job.
  tensorrt,trt   Submit a TensorRT Job  
  kfserving,kfs  Submit a kfserving Job  `
)

func NewServeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "serve",
		Short: "Serve a job.",
		Long:  serveLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewServingTensorRTCommand())
	command.AddCommand(NewServingTensorFlowCommand())
	command.AddCommand(NewServingListCommand())
	command.AddCommand(NewServingDeleteCommand())
	command.AddCommand(NewServingCustomCommand())
	command.AddCommand(NewTrafficRouterSplitCommand())
	command.AddCommand(NewServingGetCommand())
	command.AddCommand(NewServingLogCommand())
	command.AddCommand(NewKFServingCommand())
	return command
}
