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

package commands

import (
	"os"

	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/spf13/cobra"
)

// NewDeleteCommand
func NewServingDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "delete a serving job",
		Short: "delete a serving job and its associated pods",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			setupKubeconfig()
			for _, jobName := range args {
				deleteServingJob(jobName)
			}
		},
	}

	return command
}

func deleteServingJob(servingJob string) error {
	return helm.DeleteRelease(servingJob)
}
