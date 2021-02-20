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

package top

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTopNodeCommand() *cobra.Command {
	var (
		showDetails bool
		output      string
		nodeType    string
		notStop     bool
	)
	var command = &cobra.Command{
		Use:   "node",
		Short: "Display Resource (GPU) usage of nodes.",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			isDaemonMode := false
			if notStop {
				isDaemonMode = true
			}
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
			return client.Node().ListAndPrintNodes(args, utils.TransferNodeType(nodeType), utils.TransferPrintFormat(output), showDetails, notStop)
		},
	}
	command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	command.Flags().BoolVarP(&notStop, "refresh", "r", false, "Display continuously")
	command.Flags().StringVarP(&nodeType, "gpu-mode", "m", "", fmt.Sprintf("Display node information with following gpu mode:[%v]", strings.Join(utils.GetSupportedNodeTypes(), "|")))
	command.Flags().StringVarP(&output, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	return command
}
