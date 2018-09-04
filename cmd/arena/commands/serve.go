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

	validate "github.com/kubeflow/arena/util"
	"github.com/spf13/cobra"
)

var (
	modelPathSeparator = "|"
)

type ServeArgs struct {
	Image     string            `yaml:"image"`     // --image
	Gpus      int               `yaml:"gpus"`      // --gpus
	Cpu       string            `yaml:"cpu"`       // --cpu
	Memory    string            `yaml:"memory"`    // --memory
	Envs      map[string]string `yaml:"envs"`      // --envs
	Command   string            `yaml:"command"`   // --command
	Replicas  int               `yaml:"replicas"`  // --replicas
	Port      int               `yaml:"port"`      // --port
	ModelName string            `yaml:"modelName"` // --modelName
	ModelPath string            `yaml:"modelPath"` // --modelPath
	PvcName   string            `yaml:"pvcName"`
	MountPath string            `yaml:"mountPath"`
}

func (s ServeArgs) check() error {
	if name == "" {
		return fmt.Errorf("--name must be set")
	}

	err := validate.ValidateJobName(name)
	if err != nil {
		return err
	}

	return nil
}

// transform common parts of submitArgs
func (s *ServeArgs) transform() (err error) {
	// parse ModelPath, if ModePath string include ':'，should split it into PvcName and MountPath
	if strings.Index(s.ModelPath, modelPathSeparator) > 0 {
		modelPathSplitArray := strings.Split(s.ModelPath, modelPathSeparator)
		s.PvcName = modelPathSplitArray[0]
		s.MountPath = modelPathSplitArray[1]
	} else {
		s.MountPath = s.ModelPath
	}

	return nil
}

func (serveArgs *ServeArgs) addServeCommonFlags(command *cobra.Command) {

	// create subcommands
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")
	command.Flags().StringVar(&serveArgs.Image, "image", "", "the docker image name of serve job.")
	command.Flags().IntVar(&serveArgs.Port, "port", 9000, "the port of serve pod exposed.")
	command.Flags().StringVar(&serveArgs.Command, "command", "", "the command will inject to container's command.")
	command.Flags().IntVar(&serveArgs.Gpus, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Cpu, "cpu", "0", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&serveArgs.Memory, "memory", "0", "the request memory of each replica to run the serve.")
	command.Flags().IntVar(&serveArgs.Replicas, "replicas", 1, "the replicas number of the serve job.")
	command.Flags().StringVar(&serveArgs.ModelPath, "modelPath", "", "")
	command.Flags().StringArrayVarP(&envs, "envs", "e", []string{}, "the environment variables")
	command.Flags().StringVar(&serveArgs.ModelName, "modelName", "", "the model name to serve")
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

	command.AddCommand(NewServeTensorFlowCommand())

	return command
}
