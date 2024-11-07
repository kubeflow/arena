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

package serving

import "github.com/spf13/cobra"

var (
	serveUpdateLong = `update a serve job.

Available Commands:
  tensorflow,tf  Update a TensorFlow Serving Job
  triton         Update a Nvidia Triton Serving Job
  custom         Update a Custom Serving Job
  kserve         Update a KServe Serving Job
  distributed    Update a Distributed Serving Job`
)

func NewUpdateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "update",
		Short: "Update a serving job.",
		Long:  serveUpdateLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	command.AddCommand(NewUpdateTensorflowCommand())
	command.AddCommand(NewUpdateTritonCommand())
	command.AddCommand(NewUpdateCustomCommand())
	command.AddCommand(NewUpdateKServeCommand())
	command.AddCommand(NewUpdateDistributedCommand())

	return command
}
