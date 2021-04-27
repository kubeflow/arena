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

package cron

import (
	"github.com/spf13/cobra"
)

var (
	dataLong = `manage cron tasks.

Available Commands:
  tfjob                Submit a cron tfjob.
  list,ls              List the crons.
  get                  Get cron by name.
  delete,del           Delete cron by name.
  suspend              Suspend a cron.
  resume               Resume the suspend cron.
    `
)

// manage cron task
func NewCronCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "cron",
		Short: "manage cron.",
		Long:  dataLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewCronTFJobCommand())
	command.AddCommand(NewCronGetCommand())
	command.AddCommand(NewCronListCommand())
	command.AddCommand(NewCronDeleteCommand())
	command.AddCommand(NewCronSuspendCommand())
	command.AddCommand(NewCronResumeCommand())

	return command
}
