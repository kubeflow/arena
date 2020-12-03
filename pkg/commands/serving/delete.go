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

package serving

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var servingType string
	var servingVersion string
	var command = &cobra.Command{
		Use:   "delete a serving job",
		Short: "delete a serving job and its associated instances",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("not set job name,please set it")
			}
			names := args
			client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
				Kubeconfig:     viper.GetString("config"),
				LogLevel:       viper.GetString("loglevel"),
				Namespace:      viper.GetString("namespace"),
				ArenaNamespace: viper.GetString("arena-namespace"),
				IsDaemonMode:   false,
			})
			if err != nil {
				return err
			}
			return client.Serving().Delete(utils.TransferServingJobType(servingType), servingVersion, names...)
		},
	}
	command.Flags().StringVarP(&servingVersion, "version", "v", "", "The serving version to delete.")
	command.Flags().StringVarP(&servingType, "type", "T", "", fmt.Sprintf("The serving type to delete, the possible option is [%v]. (optional)", utils.GetSupportServingJobTypesInfo()))
	return command
}
