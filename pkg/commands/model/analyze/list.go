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

package analyze

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewListModelJobsCommand() *cobra.Command {
	var allNamespaces bool
	var format string
	var jobType string
	var command = &cobra.Command{
		Use:     "list",
		Short:   "List all model analyze jobs",
		Aliases: []string{"ls"},
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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
			return client.Analyze().ListAndPrint(allNamespaces, utils.TransferModelJobType(jobType), format)
		},
	}
	command.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "show all the namespaces")
	command.Flags().StringVarP(&format, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The model job type, the possible option is [%v]. (optional)", utils.GetSupportModelJobTypesInfo()))
	return command
}
