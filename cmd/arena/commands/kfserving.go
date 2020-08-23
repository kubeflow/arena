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
	kfservingChart = util.GetChartsFolder() + "/kfserving"
)

func NewKFServingCommand() *cobra.Command {
	var (
		kfservingArgs KFServingArgs
	)

	var command = &cobra.Command{
		Use:     "kfserving",
		Short:   "Submit kfserving to deploy and serve machine learning models.",
		Aliases: []string{"kfs"},
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			if kfservingArgs.GPUMemory != 0 && kfservingArgs.GPUCount != 0 {
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

			err = kfServingSummit(args, &kfservingArgs, client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	kfservingArgs.addServeCommonFlags(command)

	command.Flags().StringVar(&kfservingArgs.ModelType, "model-type", "custom", "the type of serving model,default to custom type")
	command.Flags().StringVar(&kfservingArgs.StorageUri, "storage-uri", "", "the uri direct to the model file")
	command.Flags().IntVar(&kfservingArgs.CanaryPercent, "canary-percent", 0, "the percent of the desired canary")
	command.Flags().StringVar(&kfservingArgs.Image, "image", "", "the docker image name of serve job")
	command.Flags().IntVar(&kfservingArgs.Port, "port", 0, "the port of the application listens in the custom image")

	return command
}

type KFServingArgs struct {
	// Version string `yaml:"version"` // --version
	Image     string `yaml:"image"` // --image
	ModelType     string `yaml:"modelType"` // --modelType
	CanaryPercent            int               `yaml:"canaryPercent"`            // --canaryTrafficPercent
	StorageUri     string `yaml:"storageUri"` // --storageUri
	ServeArgs `yaml:",inline"`
}

func (kfservingArgs *KFServingArgs) preprocess(client *kubernetes.Clientset, args []string) (err error) {
	kfservingArgs.Command = strings.Join(args, " ")
	log.Debugf("command: %s", kfservingArgs.Command)

	if kfservingArgs.StorageUri == "" && kfservingArgs.Image == "" {
		return fmt.Errorf("storage uri and image can not be empty at the same time.")
	}

	// validate models data
	if len(dataset) > 0 {
		err := ParseMountPath(dataset)
		if err != nil {
			return fmt.Errorf("--data has wrong value: %s", err)
		}
		kfservingArgs.ModelDirs = transformSliceToMap(dataset, ":")
	}

	log.Debugf("models:%s", kfservingArgs.StorageUri)

	// populate environment variables
	if len(envs) > 0 {
		kfservingArgs.Envs = transformSliceToMap(envs, "=")
	}

	modelServiceExists, err := checkServiceExists(client, namespace, kfservingArgs.ServingName)
	if err != nil {
		return err
	}
	kfservingArgs.ModelServiceExists = modelServiceExists

	kfservingArgs.addNodeSelectors()
	kfservingArgs.addTolerations()
	kfservingArgs.addAnnotations()

	return nil
}

func kfServingSummit(args []string, kfservingArgs *KFServingArgs, client *kubernetes.Clientset) (err error) {
	err = kfservingArgs.preprocess(client, args)
	if err != nil {
		return err
	}

	name = kfservingArgs.ServingName
	if kfservingArgs.ServingVersion == "" {
		t := time.Now()
		kfservingArgs.ServingVersion = fmt.Sprint(t.Format("200601021504"))
	}
	name += "-" + kfservingArgs.ServingVersion
	servingTypes := getServingTypes(name, namespace)
	if len(servingTypes) > 1 {
		return fmt.Errorf("The serving job with the name %s and version %s, please delete it first. `arena serve delete %s --version %s --type custom`",
			kfservingArgs.ServingName,
			kfservingArgs.ServingVersion,
			kfservingArgs.ServingName,
			kfservingArgs.ServingVersion)
	}

	return workflow.SubmitJob(name, "kfserving", namespace, kfservingArgs, kfservingChart)
}
