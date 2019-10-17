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
	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/kubeflow/arena/pkg/util"
	servingexec "github.com/kubeflow/arena/pkg/cmd/serving/attach"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

func NewServeExecCommand() *cobra.Command {
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	options := &podexec.ExecOptions{
		StreamOptions: podexec.StreamOptions{
			IOStreams: ioStreams,
		},
		Executor: &podexec.DefaultRemoteExecutor{},
	}
	var command = &cobra.Command{
		Use:   "attach JOB [-c CONTAINER] [-i INSTANCE]",
		Short: "Attach a job instance",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			setupKubeconfig()
			conf, err := clientConfig.ClientConfig()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				os.Exit(1)
			}
			acceptExecArgs := servingexec.AcceptExecArgs {
				Options: options,
				Config: conf,
				Namespace: namespace,
				ArgsIn: args,
				ArgsLenAtDash: cmd.ArgsLenAtDash(),
				Version: version,
				Type: stype,
			}
			if err := servingexec.ServingJobExecCommand(acceptExecArgs); err != nil {
				fmt.Println(err)
			}
		},
	}
	command.Flags().StringVar(&version, "version", "", "assign the serving job version")
	command.Flags().StringVar(&stype, "type", "", `assign the serving job type,type can be "tf"("tensorflow"),"trt"("tensorrt"),"custom"`)
	command.Flags().StringVarP(&options.PodName, "instance", "i", options.PodName, `job instance name,use "arena serve get" to get.`)
	command.Flags().StringVarP(&options.ContainerName, "container", "c", options.ContainerName, "Container name. If omitted, the first container in the instance will be chosen")
	return command
}