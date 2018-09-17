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
	"strconv"
	"strings"

	"github.com/kubeflow/arena/util"
	validate "github.com/kubeflow/arena/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	standalone_training_chart = "/charts/training"
	horovod_training_chart    = "/charts/tf-horovod"
	envs                      []string
	dataset                   []string
	dataDirs                  []string
)

// The common parts of the submitAthd
type submitArgs struct {
	// Name       string   `yaml:"name"`       // --name
	Image      string            `yaml:"image"`      // --image
	GPUCount   int               `yaml:"gpuCount"`   // --gpuCount
	Envs       map[string]string `yaml:"envs"`       // --envs
	WorkingDir string            `yaml:"workingDir"` // --workingDir
	Command    string            `yaml:"command"`
	// for horovod
	Mode        string `yaml:"mode"`    // --mode
	WorkerCount int    `yaml:"workers"` // --workers
	// SSHPort     int               `yaml:"sshPort"`  // --sshPort
	Retry int `yaml:"retry"` // --retry
	// DataDir  string            `yaml:"dataDir"`  // --dataDir
	DataSet  map[string]string `yaml:"dataset"`
	DataDirs []dataDirVolume   `yaml:"dataDirs"`

	EnableRDMA bool `yaml:"enableRDMA"` // --rdma
}

type dataDirVolume struct {
	HostPath      string `yaml:"hostPath"`
	ContainerPath string `yaml:"containerPath"`
	Name          string `yaml:"name"`
}

func (s submitArgs) check() error {
	if name == "" {
		return fmt.Errorf("--name must be set")
	}

	// return fmt.Errorf("must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	err := validate.ValidateJobName(name)
	if err != nil {
		return err
	}

	// if s.DataDir == "" {
	// 	return fmt.Errorf("--dataDir must be set")
	// }

	return nil
}

// transform common parts of submitArgs
func (s *submitArgs) transform() (err error) {
	// 1. handle data dirs
	log.Debugf("dataDir: %v", dataDirs)
	if len(dataDirs) > 0 {
		s.DataDirs = []dataDirVolume{}
		for i, dataDir := range dataDirs {
			hostPath, containerPath, err := util.ParseDataDirRaw(dataDir)
			if err != nil {
				return err
			}
			s.DataDirs = append(s.DataDirs, dataDirVolume{
				Name:          fmt.Sprintf("training-data-%d", i),
				HostPath:      hostPath,
				ContainerPath: containerPath,
			})
		}
	}
	// 2. handle data sets
	log.Debugf("dataset: %v", dataset)
	if len(dataset) > 0 {
		err = util.ValidateDatasets(dataset)
		if err != nil {
			return err
		}
		s.DataSet = transformSliceToMap(dataset, ":")
	}

	return nil
}

func (submitArgs *submitArgs) addJobInfoToEnv() {
	if len(submitArgs.Envs) == 0 {
		submitArgs.Envs = map[string]string{}
	}
	submitArgs.Envs["workers"] = strconv.Itoa(submitArgs.WorkerCount)
	submitArgs.Envs["gpus"] = strconv.Itoa(submitArgs.GPUCount)
}

func (submitArgs *submitArgs) addCommonFlags(command *cobra.Command) {

	// create subcommands
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")
	command.Flags().StringVar(&submitArgs.Image, "image", "", "the docker image name of training job")
	// command.MarkFlagRequired("image")
	command.Flags().IntVar(&submitArgs.GPUCount, "gpus", 0,
		"the GPU count of each worker to run the training.")
	// command.Flags().StringVar(&submitArgs.DataDir, "dataDir", "", "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	command.Flags().IntVar(&submitArgs.WorkerCount, "workers", 1,
		"the worker number to run the distributed training.")
	command.Flags().IntVar(&submitArgs.Retry, "retry", 0,
		"retry times.")
	// command.MarkFlagRequired("syncSource")
	command.Flags().StringVar(&submitArgs.WorkingDir, "workingDir", "/root", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")
	// command.MarkFlagRequired("workingDir")
	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")
	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the datasource to mount to the job, like <name_of_datasource>:<mount_point_on_job>")
	command.Flags().StringArrayVar(&dataDirs, "dataDir", []string{}, "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	// enable RDMA or not, support hostnetwork for now
	command.Flags().BoolVar(&submitArgs.EnableRDMA, "rdma", false, "enable RDMA")
}

func init() {
	if os.Getenv(CHART_PKG_LOC) != "" {
		standalone_training_chart = filepath.Join(os.Getenv(CHART_PKG_LOC), "training")
	}
}

var (
	submitLong = `Submit a job.

Available Commands:
  tfjob,tf             Submit a TFJob.
  horovod,hj           Submit a Horovod Job.
  mpijob,mpi           Submit a MPIJob.
  standalonejob,sj     Submit a standalone Job.
  tfserving,tfserving  Submit a Serving Job.
    `
)

func NewSubmitCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "submit",
		Short: "Submit a job.",
		Long:  submitLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewSubmitTFJobCommand())
	command.AddCommand(NewSubmitMPIJobCommand())
	command.AddCommand(NewSubmitHorovodJobCommand())
	// This will be deprecated soon.
	command.AddCommand(NewSubmitStandaloneJobCommand())

	return command
}

func transformSliceToMap(sets []string, split string) (valuesMap map[string]string) {
	valuesMap = map[string]string{}
	for _, member := range sets {
		splits := strings.SplitN(member, split, 2)
		if len(splits) == 2 {
			valuesMap[splits[0]] = splits[1]
		}
	}

	return valuesMap
}
