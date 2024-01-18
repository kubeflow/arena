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
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
)

// NewGetCommand
func NewGetCommand() *cobra.Command {
	var jobType string
	var showEvents bool
	var showGPUs bool
	var output string
	var command = &cobra.Command{
		Use:   "get JOB [-T JOB_TYPE]",
		Short: "Display a training job details",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return fmt.Errorf("not set job name,please set it")
			}
			now := time.Now()
			defer func() {
				log.Debugf("execute time of get training job: %v", time.Since(now))
			}()
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
			return client.Training().GetAndPrint(name, utils.TransferTrainingJobType(jobType), output, showEvents, showGPUs)
		},
	}
	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The training type to get, the possible option is %v. (optional)", utils.GetSupportTrainingJobTypesInfo()))
	command.Flags().BoolVarP(&showEvents, "events", "e", false, "Specify if show pending pod's events.")
	command.Flags().BoolVarP(&showGPUs, "gpus", "g", false, "Specify if show gpu utilizations of job.")
	command.Flags().StringVarP(&output, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	return command
}
