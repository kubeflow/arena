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
	"strings"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mpijob_chart = util.GetChartsFolder() + "/mpijob"
)

func NewSubmitMPIJobCommand() *cobra.Command {
	var (
		submitArgs submitMPIJobArgs
	)

	submitArgs.Mode = "mpijob"

	var command = &cobra.Command{
		Use:     "mpijob",
		Short:   "Submit MPIjob as training job.",
		Aliases: []string{"mpi", "mj"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			util.SetLogLevel(logLevel)
			_, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitMPIJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().StringVar(&submitArgs.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")
	// Tensorboard
	command.Flags().BoolVar(&submitArgs.UseTensorboard, "tensorboard", false, "enable tensorboard")

	msg := "the docker image for tensorboard"
	command.Flags().StringVar(&submitArgs.TensorboardImage, "tensorboardImage", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.12.0-devel", msg)
	command.Flags().MarkDeprecated("tensorboardImage", "please use --tensorboard-image instead")
	command.Flags().StringVar(&submitArgs.TensorboardImage, "tensorboard-image", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.12.0-devel", msg)

	command.Flags().StringVar(&submitArgs.TrainingLogdir, "logdir", "/training_logs", "the training logs dir, default is /training_logs")

	submitArgs.addCommonFlags(command)
	submitArgs.addSyncFlags(command)

	return command
}

type submitMPIJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	submitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	submitSyncCodeArgs `yaml:",inline"`
}

func (submitArgs *submitMPIJobArgs) prepare(args []string) (err error) {
	submitArgs.Command = strings.Join(args, " ")

	err = submitArgs.check()
	if err != nil {
		return err
	}

	commonArgs := &submitArgs.submitArgs
	err = commonArgs.transform()
	if err != nil {
		return nil
	}

	err = submitArgs.HandleSyncCode()
	if err != nil {
		return err
	}

	// process tensorboard
	submitArgs.processTensorboard(submitArgs.DataSet)

	if len(envs) > 0 {
		submitArgs.Envs = transformSliceToMap(envs, "=")
	}
	// add node labels,if given
	submitArgs.addMPINodeSelectors()
	// add tolerations, if given
	submitArgs.addMPITolerations()
	submitArgs.addMPIInfoToEnv()

	return nil
}

func (submitArgs submitMPIJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	if submitArgs.Image == "" {
		return fmt.Errorf("--image must be set ")
	}

	return nil
}

// add k8s nodes labels
func (submitArgs *submitMPIJobArgs) addMPINodeSelectors() {
	submitArgs.addNodeSelectors()
}

// add k8s tolerations for taints
func (submitArgs *submitMPIJobArgs) addMPITolerations() {
	submitArgs.addTolerations()
}
func (submitArgs *submitMPIJobArgs) addMPIInfoToEnv() {
	submitArgs.addJobInfoToEnv()
}

// Submit MPIJob
func submitMPIJob(args []string, submitArgs *submitMPIJobArgs) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	trainer := NewMPIJobTrainer(clientset)
	job, err := trainer.GetTrainingJob(name, namespace)
	if err != nil {
		log.Debugf("Check %s exist due to error %v", name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	// the master is also considered as a worker
	// submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	err = workflow.SubmitJob(name, submitArgs.Mode, namespace, submitArgs, "", mpijob_chart, clientset)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", name, submitArgs.Mode)
	return nil
}

// Submit MPIJob with helm
func submitMPIJobWithHelm(args []string, submitArgs *submitMPIJobArgs) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	exist, err := helm.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	return helm.InstallRelease(name, namespace, submitArgs, mpijob_chart)
}
