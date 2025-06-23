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

package training

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util/kubectl"
)

// NewSubmitHorovodJobCommand
func NewSubmitHorovodJobCommand() *cobra.Command {
	builder := training.NewHorovodJobBuilder()
	var command = &cobra.Command{
		Use:     "horovodjob",
		Short:   "Submit horovodjob as training job.",
		Aliases: []string{"horovod", "hj"},
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return fmt.Errorf("not found command args")
			}
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
			job, err := builder.Command(args).Build()
			if err != nil {
				return fmt.Errorf("failed to validate command args: %v", err)
			}
			if err := client.Training().Submit(job); err != nil {
				return err
			}
			fullSubmitCommand := getFullSubmitCommand(cmd, args)
			_, modelVersion, err := createRegisteredModelAndModelVersion(client, job, fullSubmitCommand)
			if modelVersion == nil {
				return err
			}
			if err := kubectl.AddTrainingJobLabel(job, "modelVersion", modelVersion.Version); err != nil {
				return fmt.Errorf("failed to patch label `modelVersion=%s` to job %s/%s: %v", modelVersion.Version, job.Type(), job.Name(), err)
			}
			return nil
		},
	}
	builder.AddCommandFlags(command)
	return command
}
