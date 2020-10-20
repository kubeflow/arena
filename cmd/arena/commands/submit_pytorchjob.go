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
	"strconv"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	pytorchjobChart = util.GetChartsFolder() + "/pytorchjob"
)

func NewSubmitPyTorchJobCommand() *cobra.Command {
	var (
		submitArgs submitPyTorchJobArgs
	)

	submitArgs.Mode = "pytorchjob"

	var command = &cobra.Command{
		Use:     "pytorchjob",
		Short:   "Submit PyTorchJob as training job.",
		Aliases: []string{"pytorch"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			util.SetLogLevel(logLevel)
			setupKubeconfig()
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

			err = submitPyTorchJob(args, &submitArgs)
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
	command.Flags().BoolVar(&submitArgs.Conscheduling, "gang", false, "enable gang scheduling")
	// TODO jiaqianjing: user can replace "alpine:3.10" with custom image addr
	//command.Flags().StringVar(&submitArgs.WorkerInitPytorchImage,
	//	"worker-init-pytorch-image",
	//	"alpine:3.10",
	//	"the docker image for workers init container image wait for master ready, default 'alpine:3.10'")

	command.Flags().StringVar(&submitArgs.CleanPodPolicy, "clean-task-policy", "None", "How to clean tasks after Training is done, support None, Running, All.")

	submitArgs.addCommonFlags(command)
	submitArgs.addSyncFlags(command)
	log.Debugf("pytorchjob command: %v", command)

	return command
}

type submitPyTorchJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	submitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	submitSyncCodeArgs `yaml:",inline"`

	// worker init pytorch image, default "alpine:3.10";
	// TODO jiaqianjing: user can set init-pytorch container image by param "--worker-init-pytorch-image"
	// WorkerInitPytorchImage string `yaml: workerInitPytorchImage`

	// clean-task-policy
	CleanPodPolicy string `yaml:"cleanPodPolicy"`
}

func (submitArgs *submitPyTorchJobArgs) prepare(args []string) (err error) {
	submitArgs.Command = strings.Join(args, " ")

	// check name、priority、image
	err = submitArgs.check()
	if err != nil {
		return err
	}

	// check clean-task-policy
	switch submitArgs.CleanPodPolicy {
	case "None", "Running", "All":
		log.Debugf("Supported cleanTaskPolicy: %s", submitArgs.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", submitArgs.CleanPodPolicy)
	}

	commonArgs := &submitArgs.submitArgs

	// e.g. process --data-dir、--data、--annotation、PodSecurityContext
	err = commonArgs.transform()
	if err != nil {
		return err
	}

	err = submitArgs.HandleSyncCode()
	if err != nil {
		return err
	}
	if err := submitArgs.addConfigFiles(); err != nil {
		return err
	}
	// process tensorboard about storage, local or not local
	submitArgs.processTensorboard(submitArgs.DataSet)

	if len(envs) > 0 {
		submitArgs.Envs = transformSliceToMap(envs, "=")
	}

	submitArgs.processCommonFlags()

	if submitArgs.Conscheduling {
		submitArgs.addPodGroupLabel()
	}

	return nil
}

// check name length、priorityClass、image
func (submitArgs submitPyTorchJobArgs) check() error {
	// check name length、priorityClass
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	if submitArgs.Image == "" {
		return fmt.Errorf("--image must be set ")
	}

	return nil
}

func (submitArgs *submitPyTorchJobArgs) addConfigFiles() error {
	return submitArgs.addJobConfigFiles()
}

func (submitArgs *submitPyTorchJobArgs) addPodGroupLabel() {
	//submitArgs.PodGroupName = name
	//submitArgs.PodGroupMinAvailable = strconv.Itoa(submitArgs.WorkerCount)
	submitArgs.PodGroupName = yinlei-test
	submitArgs.PodGroupMinAvailable = 10
}

// Submit PyTorchJob
func submitPyTorchJob(args []string, submitArgs *submitPyTorchJobArgs) (err error) {
	// param check, set tensorboard storage mode and add env、selector、toleration、config-file
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	trainer := NewPyTorchJobTrainer(clientset)

	// check pytorch job has exist
	job, err := trainer.GetTrainingJob(name, namespace)
	if err != nil {
		log.Debugf("Check %s exist due to error %v", name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	// the master is also considered as a worker
	submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	err = workflow.SubmitJob(name, submitArgs.Mode, namespace, submitArgs, pytorchjobChart, submitArgs.addHelmOptions()...)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", name, submitArgs.Mode)
	return nil
}
