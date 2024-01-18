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
	"github.com/kubeflow/arena/pkg/apis/serving"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTrafficRouterSplitCommand() *cobra.Command {
	builder := serving.NewTrafficRouterBuilder()
	var command = &cobra.Command{
		Use:     "traffic-split",
		Short:   "Adjust traffic routing dynamically for tfserving jobs",
		Aliases: []string{"trs", "traffic-router", "traffic-router-split", "traffic-shift", "traffic-shifting"},
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
			trafficRouteSplitArgs, err := builder.Build()
			if err != nil {
				return fmt.Errorf("failed to validate args: %v", err)
			}
			return client.Serving().TrafficRouterSplit(trafficRouteSplitArgs)
		},
	}
	builder.AddCommandFlags(command)
	return command
}
