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

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var (
	predictChart = util.GetChartsFolder() + "/predict"
)

func NewServingPredictCommand() *cobra.Command {
	var (
		servePredictArgs ServePredictArgs
	)

	var command = &cobra.Command{
		Use:     "predict",
		Short:   "Submit common predict job to deploy and serve machine learning models.",
		Aliases: []string{"tf"},
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			if servePredictArgs.GPUMemory != 0 && servePredictArgs.GPUCount != 0 {
				fmt.Println("gpucount and gpumemory should not be used at the same time.You can only choose one mode")
				os.Exit(1)
			}
			util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
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

			err = servePredict(args, &servePredictArgs, client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	servePredictArgs.addServeCommonFlags(command)

	// TFServingJob
	// add grpc port and rest api port
	command.Flags().StringVar(&servePredictArgs.Image, "image", "", "the docker image name of serve job")
	command.Flags().StringVar(&servePredictArgs.Version, "version", "", "the version of serve job")
	command.Flags().IntVar(&servePredictArgs.Port, "port", 8500, "the port of Predict gRPC listening port")
	command.Flags().IntVar(&servePredictArgs.RestfulPort, "restful-port", 8501, "the port of Predict RESTful listening port")

	return command
}

type ServePredictArgs struct {
	Version string `yaml:"version"` // --version
	Image   string `yaml:"image"`   // --image

	ServeArgs `yaml:",inline"`
}

func (servePredictArgs *ServePredictArgs) preprocess(client *kubernetes.Clientset, args []string) (err error) {
	//servePredictArgs.Command = strings.Join(args, " ")
	log.Debugf("command: %s", servePredictArgs.Command)

	if servePredictArgs.Image == "" {
		return fmt.Errorf("image must be specified.")
	}

	// validate models data
	if len(dataset) > 0 {
		err := ParseMountPath(dataset)
		if err != nil {
			return fmt.Errorf("--data has wrong value: %s", err)
		}
		servePredictArgs.ModelDirs = transformSliceToMap(dataset, ":")
	}

	log.Debugf("models:%s", servePredictArgs.ModelDirs)

	//validate Istio enablement
	err = servePredictArgs.ServeArgs.validateIstioEnablement()
	if err != nil {
		return err
	}

	// populate environment variables
	if len(envs) > 0 {
		servePredictArgs.Envs = transformSliceToMap(envs, "=")
	}

	modelServiceExists, err := checkServiceExists(client, namespace, servePredictArgs.ServingName)
	if err != nil {
		return err
	}
	servePredictArgs.ModelServiceExists = modelServiceExists

	return nil
}

func servePredict(args []string, servePredictArgs *ServePredictArgs, client *kubernetes.Clientset) (err error) {
	err = servePredictArgs.preprocess(client, args)
	if err != nil {
		return err
	}

	name = servePredictArgs.ServingName
	if servePredictArgs.Version != "" {
		name += "-" + servePredictArgs.Version
	}
	return workflow.SubmitJob(name, "predict", namespace, servePredictArgs, predictChart)
}
