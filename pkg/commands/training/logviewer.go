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

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewLogViewerCommand() *cobra.Command {
	var jobType string
	var command = &cobra.Command{
		Use:   "logviewer JOB [-T JOB_TYPE]",
		Short: "Display Log Viewer URL of a training job",
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
			urls, err := client.Training().LogViewer(name, utils.TransferTrainingJobType(jobType))
			if err != nil {
				return fmt.Errorf("failed to get LogViewer: %v", err)
			}
			if len(urls) == 0 {
				fmt.Printf("No LogViewer Installed\n")
				return nil
			}
			fmt.Printf("Your LogViewer will be available on:\n")
			for _, url := range urls {
				fmt.Println(url)
			}
			return nil
		},
	}
	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The training type to get, the possible option is %v. (optional)", utils.GetSupportTrainingJobTypesInfo()))
	return command
}
