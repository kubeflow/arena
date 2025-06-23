// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package top

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTopJobCommand() *cobra.Command {
	var (
		allNamespaces bool
		format        string
		jobType       string
		notStop       bool
		instanceName  string
	)
	var command = &cobra.Command{
		Use:   "job",
		Short: "Display Resource (GPU) usage of jobs.",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			isDaemonMode := notStop

			client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
				Kubeconfig:     viper.GetString("config"),
				LogLevel:       viper.GetString("loglevel"),
				Namespace:      viper.GetString("namespace"),
				ArenaNamespace: viper.GetString("arena-namespace"),
				IsDaemonMode:   isDaemonMode,
			})
			if err != nil {
				return fmt.Errorf("failed to create arena client: %v", err)
			}
			return client.Training().Top(
				args,
				allNamespaces,
				utils.TransferTrainingJobType(jobType),
				instanceName,
				notStop,
				utils.TransferPrintFormat(format),
			)
		},
	}
	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")
	_ = command.Flags().MarkDeprecated("allNamespaces", "please use --all-namespaces instead")
	command.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "show all the namespaces")
	command.Flags().StringVarP(&format, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	command.Flags().BoolVarP(&notStop, "refresh", "r", false, "Display continuously")
	command.Flags().StringVarP(&instanceName, "instance", "i", "", "Display instance top info")
	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The training type, the possible option is [%v]. (optional)", utils.GetSupportTrainingJobTypesInfo()))
	return command
}
