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

	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/arena/util/helm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	tfjob_chart = "/charts/tfjob"
)

func NewSubmitTFJobCommand() *cobra.Command {
	var (
		submitArgs submitTFJobArgs
	)

	submitArgs.Mode = "tfjob"

	var command = &cobra.Command{
		Use:     "tfjob",
		Short:   "Submit TFJob as training job.",
		Aliases: []string{"tf"},
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

			err = ensureNamespace(clientset, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitTFJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	submitArgs.addCommonFlags(command)
	submitArgs.addSyncFlags(command)

	// TFJob
	command.Flags().StringVar(&submitArgs.WorkerImage, "workerImage", "", "the docker image for tensorflow workers")
	command.Flags().StringVar(&submitArgs.PSImage, "psImage", "", "the docker image for tensorflow workers")
	command.Flags().IntVar(&submitArgs.PSCount, "ps", 0, "the number of the parameter servers.")
	command.Flags().IntVar(&submitArgs.PSPort, "psPort", 22223, "the port of the parameter server.")
	command.Flags().IntVar(&submitArgs.WorkerPort, "workerPort", 22222, "the port of the worker.")
	command.Flags().StringVar(&submitArgs.WorkerCpu, "workerCpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.WorkerMemory, "workerMemory", "", "the memory resource to use for the worker, like 1Gi.")
	command.Flags().StringVar(&submitArgs.PSCpu, "psCpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.PSMemory, "psMemory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	// How to clean up Task
	command.Flags().StringVar(&submitArgs.CleanPodPolicy, "cleanTaskPolicy", "Running", "How to clean tasks after Training is done, only support Running, None.")

	// Tensorboard
	command.Flags().BoolVar(&submitArgs.UseTensorboard, "tensorboard", false, "enable tensorboard")
	command.Flags().StringVar(&submitArgs.TensorboardImage, "tensorboardImage", "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.5.0-devel", "the docker image for tensorboard")
	command.Flags().StringVar(&submitArgs.TrainingLogdir, "logdir", "/training_logs", "the training logs dir, default is /training_logs")

	// command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}

type submitTFJobArgs struct {
	Port           int    // --port, it's used set workerPort and PSPort if they are not set
	WorkerImage    string `yaml:"workerImage"`    // --workerImage
	WorkerPort     int    `yaml:"workerPort"`     // --workerPort
	PSPort         int    `yaml:"psPort"`         // --psPort
	PSCount        int    `yaml:"ps"`             // --ps
	PSImage        string `yaml:"psImage"`        // --psImage
	WorkerCpu      string `yaml:"workerCPU"`      // --workerCpu
	WorkerMemory   string `yaml:"workerMemory"`   // --workerMemory
	PSCpu          string `yaml:"psCPU"`          // --psCpu
	PSMemory       string `yaml:"psMemory"`       // --psMemory
	CleanPodPolicy string `yaml:"cleanPodPolicy"` // --cleanTaskPolicy
	// determine if it has gang scheduler
	HasGangScheduler bool `yaml:"hasGangScheduler"`

	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	submitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	submitSyncCodeArgs `yaml:",inline"`
}

func (submitArgs *submitTFJobArgs) prepare(args []string) (err error) {
	submitArgs.Command = strings.Join(args, " ")

	err = submitArgs.transform()
	if err != nil {
		return err
	}

	err = submitArgs.check()
	if err != nil {
		return err
	}

	err = submitArgs.HandleSyncCode()
	if err != nil {
		return err
	}

	// process tensorboard
	submitArgs.processTensorboad()

	commonArgs := &submitArgs.submitArgs
	err = commonArgs.transform()
	if err != nil {
		return nil
	}

	if len(envs) > 0 {
		submitArgs.Envs = transformSliceToMap(envs, "=")
	}
	// pass the workers, gpu to environment variables
	// addTFJobInfoToEnv(submitArgs)
	submitArgs.addTFJobInfoToEnv()
	return nil
}

func (submitArgs submitTFJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	switch submitArgs.CleanPodPolicy {
	case "None", "Running":
		log.Debugf("Supported cleanTaskPolicy: %s", submitArgs.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", submitArgs.CleanPodPolicy)
	}

	if submitArgs.WorkerCount == 0 {
		return fmt.Errorf("--workers must be greater than 0")
	}

	if submitArgs.WorkerImage == "" {
		return fmt.Errorf("--image or --workerImage must be set")
	}

	// distributed tensorflow should enable workerPort
	if submitArgs.WorkerCount+submitArgs.PSCount > 1 {
		if submitArgs.WorkerPort <= 0 {
			return fmt.Errorf("--port or --workerPort must be set")
		}
	}

	if submitArgs.PSCount > 0 {
		if submitArgs.PSImage == "" {
			return fmt.Errorf("--image or --psImage must be set")
		}

		if submitArgs.PSPort <= 0 {
			return fmt.Errorf("--port or --psPort must be set")
		}
	}

	return nil
}

func (submitArgs *submitTFJobArgs) transform() error {
	if submitArgs.WorkerPort == 0 {
		submitArgs.WorkerPort = submitArgs.Port
	}

	if submitArgs.WorkerImage == "" {
		submitArgs.WorkerImage = submitArgs.Image
	}

	if submitArgs.PSCount > 0 {
		if submitArgs.PSPort == 0 {
			submitArgs.PSPort = submitArgs.Port
		}

		if submitArgs.PSImage == "" {
			submitArgs.PSImage = submitArgs.Image
		}
	}

	// check Gang scheduler
	submitArgs.checkGangCapablitiesInCluster()

	return nil
}

func (submitArgs *submitTFJobArgs) addTFJobInfoToEnv() {
	submitArgs.addJobInfoToEnv()
}

func (submitArgs *submitTFJobArgs) checkGangCapablitiesInCluster() {
	gangCapablity := false
	if clientset != nil {
		_, err := clientset.AppsV1beta1().Deployments(metav1.NamespaceSystem).Get(gangSchdName, metav1.GetOptions{})
		if err != nil {
			log.Debugf("Failed to find %s due to %v", gangSchdName, err)
		} else {
			log.Debugf("Found %s successfully, the gang scheduler is enabled in the cluster.", gangSchdName)
			gangCapablity = true
		}
	}

	submitArgs.HasGangScheduler = gangCapablity
}

func submitTFJob(args []string, submitArgs *submitTFJobArgs) (err error) {
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

	// the master is also considered as a worker
	// submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	return helm.InstallRelease(name, namespace, submitArgs, tfjob_chart)
}
