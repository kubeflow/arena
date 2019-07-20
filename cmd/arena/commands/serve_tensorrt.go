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
	"github.com/kubeflow/arena/pkg/util/helm"
	//log "github.com/sirupsen/logrus"
	"github.com/kubeflow/arena/pkg/workflow"
	"github.com/spf13/cobra"
)

var (
	trtservingChart        = util.GetChartsFolder() + "/trtserving"
	defaultTRTServingImage = "registry.cn-beijing.aliyuncs.com/xiaozhou/tensorrt-serving:18.12-py3"
)

func NewServingTensorRTCommand() *cobra.Command {
	var (
		serveTensorRTArgs ServeTensorRTArgs
	)

	var command = &cobra.Command{
		Use:     "tensorrt",
		Short:   "Submit tensorRT inference serving job to deploy and serve machine learning models.",
		Aliases: []string{"trt"},
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			if serveTensorRTArgs.GPUMemory != 0 && serveTensorRTArgs.GPUCount != 0 {
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

			err = ensureNamespace(client, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = serveTensorRT(&serveTensorRTArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	serveTensorRTArgs.addServeCommonFlags(command)

	// TRTServingJob
	command.Flags().StringVar(&serveTensorRTArgs.Image, "image", defaultTRTServingImage, "the docker image name of serve job, and the default image is "+defaultTRTServingImage)
	command.Flags().StringVar(&serveTensorRTArgs.ModelStore, "modelStore", "", "the path of tensorRT model path")
	command.Flags().IntVar(&serveTensorRTArgs.HttpPort, "httpPort", 8000, "the port of http serving server")
	command.Flags().IntVar(&serveTensorRTArgs.GrpcPort, "grpcPort", 8001, "the port of grpc serving server")
	command.Flags().IntVar(&serveTensorRTArgs.MetricsPort, "metricPort", 8002, "the port of metrics server")
	command.Flags().BoolVar(&serveTensorRTArgs.AllowMetrics, "allowMetrics", false, "Open Metric")

	return command
}

type ServeTensorRTArgs struct {
	Image        string `yaml:"image"`        // --image
	ModelStore   string `yaml:"modelStore"`   // --modelStore
	MetricsPort  int    `yaml:"metricsPort"`  // --metricsPort
	HttpPort     int    `yaml:"httpPort"`     // --httpPort
	GrpcPort     int    `yaml:"grpcPort"`     // --grpcPort
	AllowMetrics bool   `yaml:"allowMetrics"` // --allowMetrics

	ServeArgs `yaml:",inline"`
}

func (serveTensorRTArgs *ServeTensorRTArgs) validate() (err error) {
	if serveTensorRTArgs.GPUCount == 0 {
		return fmt.Errorf("--gpus must be specific at least 1 GPU")
	}
	return nil
}

func (serveTensorRTArgs *ServeTensorRTArgs) preprocess() (err error) {
	// validate models data
	if len(dataset) > 0 {
		err := ParseMountPath(dataset)
		if err != nil {
			return fmt.Errorf("--data has wrong value: %s", err)
		}
		serveTensorRTArgs.ModelDirs = transformSliceToMap(dataset, ":")
	}
	return nil
}

func serveTensorRT(serveTensorRTArgs *ServeTensorRTArgs) (err error) {
	err = serveTensorRTArgs.preprocess()
	if err != nil {
		return err
	}
	err = serveTensorRTArgs.validate()
	if err != nil {
		return err
	}
	exist, err := helm.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it firstly. use 'arena delete %s'", name, name)
	}

	name = serveTensorRTArgs.ServingName

	//return helm.InstallRelease(name, namespace, serveTensorRTArgs, trtservingChart)
	return workflow.SubmitJob(name, "trt-serving", namespace, serveTensorRTArgs, trtservingChart)
}
