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

import "github.com/spf13/cobra"

var (
	scaleoutLong = `scaleout a job.

Available Commands:
  etjob,et           scaleout a ETJob.
    `
)

func NewScaleOutCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "scaleout",
		Short: "scaleout a job.",
		Long:  scaleoutLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewScaleOutETJobCommand())

	return command
}
