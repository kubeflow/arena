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
	"time"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var (
	customChart = util.GetChartsFolder() + "/custom-serving"
)

func NewServingCustomCommand() *cobra.Command {
	var (
		serveCustomArgs ServeCustomArgs
	)

	var command = &cobra.Command{
		Use:     "custom",
		Short:   "Submit custom serving to deploy and serve machine learning models.",
		Aliases: []string{"tf"},
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			if serveCustomArgs.GPUMemory != 0 && serveCustomArgs.GPUCount != 0 {
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

			err = servePredict(args, &serveCustomArgs, client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	serveCustomArgs.addServeCommonFlags(command)

	// TFServingJob
	// add grpc port and rest api port
	command.Flags().StringVar(&serveCustomArgs.Image, "image", "", "the docker image name of serve job")
	command.Flags().IntVar(&serveCustomArgs.Port, "port", 0, "the port of gRPC listening port,default is 0 represents that don't create service listening on this port")
	command.Flags().IntVar(&serveCustomArgs.RestfulPort, "restful-port", 0, "the port of RESTful listening port,default is 0 represents that don't create service listening on this port")

	return command
}

type ServeCustomArgs struct {
	// Version string `yaml:"version"` // --version
	Image     string `yaml:"image"` // --image
	ServeArgs `yaml:",inline"`
}

func (serveCustomArgs *ServeCustomArgs) preprocess(client *kubernetes.Clientset, args []string) (err error) {
	//serveCustomArgs.Command = strings.Join(args, " ")
	if err := serveCustomArgs.PreCheck(); err != nil {
		return err
	}
	serveCustomArgs.Command = strings.Join(args, " ")
	log.Debugf("command: %s", serveCustomArgs.Command)

	if serveCustomArgs.Image == "" {
		return fmt.Errorf("image must be specified.")
	}

	// validate models data
	if len(dataset) > 0 {
		err := ParseMountPath(dataset)
		if err != nil {
			return fmt.Errorf("--data has wrong value: %s", err)
		}
		serveCustomArgs.ModelDirs = transformSliceToMap(dataset, ":")
	}

	log.Debugf("models:%s", serveCustomArgs.ModelDirs)

	//validate Istio enablement
	err = serveCustomArgs.ServeArgs.validateIstioEnablement()
	if err != nil {
		return err
	}

	// populate environment variables
	if len(envs) > 0 {
		serveCustomArgs.Envs = transformSliceToMap(envs, "=")
	}

	modelServiceExists, err := checkServiceExists(client, namespace, serveCustomArgs.ServingName)
	if err != nil {
		return err
	}
	serveCustomArgs.ModelServiceExists = modelServiceExists

	serveCustomArgs.addNodeSelectors()
	serveCustomArgs.addTolerations()
	serveCustomArgs.addAnnotations()

	return nil
}

func servePredict(args []string, serveCustomArgs *ServeCustomArgs, client *kubernetes.Clientset) (err error) {
	err = serveCustomArgs.preprocess(client, args)
	if err != nil {
		return err
	}

	name = serveCustomArgs.ServingName
	if serveCustomArgs.ServingVersion == "" {
		t := time.Now()
		serveCustomArgs.ServingVersion = fmt.Sprint(t.Format("200601021504"))
	}

	name += "-" + serveCustomArgs.ServingVersion

	servingTypes := getServingTypes(name, namespace)
	if len(servingTypes) > 1 {
		return fmt.Errorf("The serving job with the name %s and version %s, please delete it first. `arena serve delete %s --version %s --type custom`",
			serveCustomArgs.ServingName,
			serveCustomArgs.ServingVersion,
			serveCustomArgs.ServingName,
			serveCustomArgs.ServingVersion)
	}

	return workflow.SubmitJob(name, "custom-serving", namespace, serveCustomArgs, "", customChart, clientset)
}
