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

	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

func NewExecCommand() *cobra.Command {
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
			job, err := searchTrainingJob(args[0], trainingType, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if options.PodName == "" {
				chiefPod := job.ChiefPod()
				options.PodName = path.Base(chiefPod.ObjectMeta.SelfLink)
			}
			argsLenAtDash := cmd.ArgsLenAtDash()
			if err := options.Complete(conf, namespace, args, argsLenAtDash); err != nil {
				if err == types.ErrInvalidUsage {
					fmt.Println(podexec.ExecUsageStr)
					os.Exit(1)
				}
				log.Errorf("complete exec options failed,reason: %v", err)
				os.Exit(1)
			}
			if err := options.Validate(); err != nil {
				log.Errorf("validate failed,reason: %v", err)
				os.Exit(1)
			}
			if err := options.Run(); err != nil {
				//log.Errorf("exec in container failed,reason: %v", err)
				fmt.Printf("exec in container failed,reason: %v\n", err)
				os.Exit(1)
			}
		},
	}
	command.Flags().StringVar(&trainingType, "type", "", "The type of the training job, the possible option is tfjob, mpijob, horovodjob or standalonejob. (optional)")
	command.Flags().StringVarP(&options.PodName, "instance", "i", options.PodName, "Job instance name")
	command.Flags().StringVarP(&options.ContainerName, "container", "c", options.ContainerName, "Container name. If omitted, the first container in the instance will be chosen")
	return command
}
