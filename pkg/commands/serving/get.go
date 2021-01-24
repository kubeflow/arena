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

var output string

// NewGetCommand
func NewGetCommand() *cobra.Command {
	var servingType string
	var version string
	var output string
	var bashCompletionFlags = map[string]string{
		"version": "__arena_serve_all_version",
		"type":    "__arena_serve_all_type",
	}
	var command = &cobra.Command{
		Use:   "get JOB [-T JOB_TYPE] [-v JOB_VERSION]",
		Short: "Display a serving job details",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
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
			return client.Serving().GetAndPrint(name, version, utils.TransferServingJobType(servingType), output)
		},
	}
	command.Flags().StringVarP(&version, "version", "v", "", "Set the serving job version")
	command.Flags().StringVarP(&servingType, "type", "T", "", fmt.Sprintf("The serving type, the possible option is [%v]. (optional)", utils.GetSupportServingJobTypesInfo()))
	command.Flags().StringVarP(&output, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	for name, completion := range bashCompletionFlags {
		if command.Flag(name) != nil {
			if command.Flag(name).Annotations == nil {
				command.Flag(name).Annotations = map[string][]string{}
			}
			command.Flag(name).Annotations[cobra.BashCompCustom] = append(
				command.Flag(name).Annotations[cobra.BashCompCustom],
				completion,
			)
		}
	}
	return command
}
