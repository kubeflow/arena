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
	"strings"

	"github.com/spf13/cobra"
	"regexp"
	log "github.com/sirupsen/logrus"
	validate "github.com/kubeflow/arena/util"
)

var (
	modelPathSeparator = ":"
	regexp4serviceName = "^[a-z0-9A-Z_-]+$"
)

type ServeArgs struct {
	Image          string            `yaml:"image"`     // --image
	Gpus           int               `yaml:"gpus"`      // --gpus
	Cpu            string            `yaml:"cpu"`       // --cpu
	Memory         string            `yaml:"memory"`    // --memory
	Envs           map[string]string `yaml:"envs"`      // --envs
	Command        string            `yaml:"command"`   // --command
	Replicas       int               `yaml:"replicas"`  // --replicas
	Port           int               `yaml:"port"`      // --port
	ModelName      string            `yaml:"modelName"` // --modelName
	ModelPath      string            `yaml:"modelPath"` // --modelPath
	PvcName        string            `yaml:"pvcName"`
	MountPath      string            `yaml:"mountPath"`
	EnableIstio    bool              `yaml:"enableIstio"`    // --enableIstio
	ServiceName    string            `yaml:"serviceName"`    //--serviceName
	ServiceVersion string            `yaml:"serviceVersion"` //--serviceVersion
	DataDirs       []dataDirVolume   `yaml:"dataDirs"`
}

func (s ServeArgs) validateIstioEnablement() error {
	log.Debugf("--enableIstio=%t is specified.", s.EnableIstio)
	if !s.EnableIstio {
		return nil
	}

	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.ServiceName)
	if !matched {
		return fmt.Errorf("--serviceName should be numbers, letters, dashes, and underscores ONLY")
	}
	log.Debugf("--serviceVersion=%s is specified.", s.ServiceVersion)
	if s.ServiceVersion == "" {
		return fmt.Errorf("--serviceVersion must be specified if enableIstio=true")
	}

	return nil
}

func (s ServeArgs) validateModelName() error {
	if s.ModelName == "" {
		return fmt.Errorf("--modelName cannot be blank")
	}

	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.ModelName)
	if !matched {
		return fmt.Errorf("--modelName should be numbers, letters, dashes, and underscores ONLY")
	}

	return nil
}

func ParseBasePath(basePath string) (dataDir dataDirVolume, err error) {
	// parse basePath, if basePath string include ':'，should split it into PvcName, ModelPathInPVC and MountPath
	dataDir = dataDirVolume{}
	if strings.Index(basePath, modelPathSeparator) > 0 {
		modelPathSplitArray := strings.Split(basePath, modelPathSeparator)
		dataDir.Name = modelPathSplitArray[0]
		if len(modelPathSplitArray) >= 3 {
			dataDir.HostPath = modelPathSplitArray[1]
			dataDir.ContainerPath = modelPathSplitArray[2]
			err := validate.ValidateMountDestination(dataDir.HostPath)
			if err != nil {
				return dataDir, err
			}
			err = validate.ValidateMountDestination(dataDir.ContainerPath)
			if err != nil {
				return dataDir, err
			}
			dataDir.ContainerPath = strings.Trim(dataDir.ContainerPath, "")
			dataDir.ContainerPath = strings.TrimRight(dataDir.ContainerPath, "/")
		} else {
			return dataDir, fmt.Errorf("the modelPath should be specified as pvc:modelPathInPVC:mountPathInContainer")
		}
	} else {
		//no pvc, use local path
		//s.ModelPath is the local model path
		err := validate.ValidateMountDestination(basePath)
		if err != nil {
			return dataDir, err
		}
		dataDir.Name = ""
		dataDir.HostPath = ""
		dataDir.ContainerPath = strings.Trim(basePath, "")
		dataDir.ContainerPath = strings.TrimRight(dataDir.ContainerPath, "/")
	}
	log.Debugf("dataDir: %s", dataDir)
	return dataDir, nil
}

func (s *ServeArgs) validateModelPath() (err error) {
	log.Debugf("ModelPath: %s", s.ModelPath)
	//hostPath, containerPath, err := util.ParseDataDirRaw(s.ModelPath)
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}
	//log.Debugf("hostpath: %s", hostPath)
	//log.Debugf("containerPath: %s", containerPath)
	//s.DataDirs = append(s.DataDirs, dataDirVolume{
	//	Name:          "serving-model",
	//	HostPath:      hostPath,
	//	ContainerPath: containerPath,
	//})

	//log.Debugf("s.DataDirs: %v", s.DataDirs)
	// parse ModelPath, if ModePath string include ':'，should split it into PvcName, ModelPathInPVC and MountPath
	if strings.Index(s.ModelPath, modelPathSeparator) > 0 {
		modelPathSplitArray := strings.Split(s.ModelPath, modelPathSeparator)
		s.PvcName = modelPathSplitArray[0]
		if len(modelPathSplitArray) == 3 {
			//MountPath == ModelPathInPVC
			s.ModelPath = modelPathSplitArray[1]
			s.MountPath = modelPathSplitArray[2]
			err := validate.ValidateMountDestination(s.ModelPath)
			if err != nil {
				return err
			}
			err = validate.ValidateMountDestination(s.MountPath)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("the modelPath should be specified as pvc:modelPathInPVC:mountPathInContainer")
		}
	} else {
		//no pvc, use local path
		//s.ModelPath is the local model path
		s.MountPath = ""
		err := validate.ValidateMountDestination(s.ModelPath)
		if err != nil {
			return err
		}
	}
	log.Debugf("PvcName: %s", s.PvcName)
	log.Debugf("ModelPath: %s", s.ModelPath)
	log.Debugf("MountPath: %s", s.MountPath)
	return nil
}

func (serveArgs *ServeArgs) addServeCommonFlags(command *cobra.Command) {

	// create subcommands
	//command.Flags().StringVar(&name, "name", "", "override name")
	command.Flags().StringVar(&serveArgs.Image, "image", defaultTfServingImage, "the docker image name of serve job, default image is "+defaultTfServingImage)
	command.Flags().IntVar(&serveArgs.Port, "port", 8500, "the port of serve pod exposed.")
	command.Flags().StringVar(&serveArgs.Command, "command", "", "the command will inject to container's command.")
	command.Flags().IntVar(&serveArgs.Gpus, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Cpu, "cpu", "", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Memory, "memory", "", "the request memory of each replica to run the serve.")
	command.Flags().IntVar(&serveArgs.Replicas, "replicas", 1, "the replicas number of the serve job.")
	command.Flags().StringVar(&serveArgs.ModelPath, "modelPath", "", "the model path for serving following the format: pvc-name:/root/model")
	command.Flags().StringArrayVarP(&envs, "envs", "e", []string{}, "the environment variables")
	command.Flags().StringVar(&serveArgs.ModelName, "modelName", "", "the model name for serving")
	command.Flags().BoolVar(&serveArgs.EnableIstio, "enableIstio", false, "enable Istio for serving or not (disable Istio by default)")
	command.Flags().StringVar(&serveArgs.ServiceName, "serviceName", "", "the serving name")
	command.Flags().StringVar(&serveArgs.ServiceVersion, "serviceVersion", "", "the serving version")

	command.MarkFlagRequired("serviceName")
	//command.MarkFlagRequired("modelName")
	//command.MarkFlagRequired("modelPath")
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
    `
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

	command.AddCommand(NewServingTensorFlowCommand())
	command.AddCommand(NewServingListCommand())
	command.AddCommand(NewServingDeleteCommand())

	return command
}
