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

package training

import (
	"fmt"
	"os"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewAttachCommand() *cobra.Command {
	var jobType string
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	options := &podexec.ExecOptions{
		StreamOptions: podexec.StreamOptions{
			IOStreams: ioStreams,
		},
		Executor: &podexec.DefaultRemoteExecutor{},
	}
	var command = &cobra.Command{
		Use:   "attach JOB [-i INSTANCE] [-c CONTAINER]",
		Short: "Attach a training job and execute some commands",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return fmt.Errorf("not set job name,please set it")
			}
			name := args[0]
			client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
				Kubeconfig:     viper.GetString("config"),
				LogLevel:       viper.GetString("loglevel"),
				Namespace:      viper.GetString("namespace"),
				ArenaNamespace: viper.GetString("arena-namespace"),
				IsDaemonMode:   false,
			})
			if err != nil {
				return fmt.Errorf("failed to create arena client: %v", err)
			}
			job, err := client.Training().Get(name, utils.TransferTrainingJobType(jobType))
			if err != nil {
				return err
			}
			if options.PodName == "" {
				options.PodName = job.ChiefName
			}
			argsLenAtDash := cmd.ArgsLenAtDash()
			if len(args) == 1 {
				args = append(args, "sh")
			}
			if err := options.Complete(args, viper.GetString("namespace"), argsLenAtDash); err != nil {
				return err
			}
			if err := options.Validate(); err != nil {
				return err
			}
			return options.Run()
		},
	}
	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The training type to get, the possible option is %v. (optional)", utils.GetSupportTrainingJobTypesInfo()))
	command.Flags().StringVarP(&options.PodName, "instance", "i", "", "Job instance name")
	command.Flags().StringVarP(&options.ContainerName, "container", "c", "", "Container name. If omitted, the first container in the instance will be chosen")
	return command
}
