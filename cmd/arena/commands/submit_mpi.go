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
	"path"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	SubmitMpiCommand = "submitmpi"
)

var (
	mpijob_chart = path.Join(util.GetChartsFolder(), "mpijob")
)

func NewRunaiSubmitMPIJobCommand() *cobra.Command {
	var (
		submitArgs submitMPIJobArgs
	)

	submitArgs.Mode = "mpijob"

	var command = &cobra.Command{
		Use:     SubmitMpiCommand + " [NAME]",
		Short:   "Submit a Runai MPI job.",
		Aliases: []string{"mpi", "mj"},
		Hidden:  true,
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			submitArgs.setCommonRun(cmd, args, kubeClient)

			err = submitMPIJob(args, &submitArgs, kubeClient)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	// Tensorboard
	// command.Flags().BoolVar(&submitArgs.UseTensorboard, "tensorboard", false, "enable tensorboard")

	// msg := "the docker image for tensorboard"
	// command.Flags().StringVar(&submitArgs.TensorboardImage, "tensorboardImage", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.12.0-devel", msg)
	// command.Flags().MarkDeprecated("tensorboardImage", "please use --tensorboard-image instead")
	// command.Flags().StringVar(&submitArgs.TensorboardImage, "tensorboard-image", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.12.0-devel", msg)

	// command.Flags().StringVar(&submitArgs.TrainingLogdir, "logdir", "/training_logs", "the training logs dir, default is /training_logs")
	// command.Flags().StringVar(&(submitArgs.NodeName), "nodename", "", "Enforce node affinity by setting a nodeName label")
	command.Flags().StringVar(&submitArgs.Command, "command", "", "Run this command on container start. Use together with --args.")
	command.Flags().IntVar(&submitArgs.NumberProcesses, "num-processes", 1, "the number of processes to run the distributed training.")

	submitArgs.addCommonFlags(command)
	submitArgs.addSyncFlags(command)

	return command

}

type submitMPIJobArgs struct {
	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	submitTensorboardArgs `yaml:",inline"`
	Command               string `yaml:"command"`
	NodeName              string `yaml:"nodeName,omitempty"`
	NumberProcesses       int    `yaml:"numProcesses"` // --workers
	// for sync up source code
	submitSyncCodeArgs `yaml:",inline"`
}

func (submitArgs *submitMPIJobArgs) prepare(args []string) (err error) {
	err = submitArgs.check()
	if err != nil {
		return err
	}
	return nil
}

func (submitArgs submitMPIJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	if submitArgs.Image == "" {
		return fmt.Errorf("--image must be set")
	}

	if float64(int(*submitArgs.GPU)) != *submitArgs.GPU {
		return fmt.Errorf("--gpu must be an integer")
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

// Submit MPIJob
func submitMPIJob(args []string, submitArgs *submitMPIJobArgs, client *client.Client) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	trainer := NewMPIJobTrainer(*client)
	job, err := trainer.GetTrainingJob(name, submitArgs.Namespace)
	if err != nil {
		log.Debugf("Check %s exist due to error %v", name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s already exists, please delete it first. use 'runai delete %s'", name, name)
	}

	// the master is also considered as a worker
	// submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	err = workflow.SubmitJob(name, submitArgs.Mode, submitArgs.Namespace, submitArgs, "", mpijob_chart, client.GetClientset(), dryRun)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can run `runai get %s --type %s` to check the job status", name, submitArgs.Mode)
	return nil
}
