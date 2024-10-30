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
	serveLong = `serve a job.

Available Commands:
  tensorflow,tf  Submit a TensorFlow Serving Job
  triton         Submit a Nvidia Triton Serving Job
  custom         Submit a Custom Serving Job  
  kfserving,kfs  Submit a kubeflow Serving Job
  kserve         Submit a KServe Serving Job
  seldon         Submit a Seldon Serving Job
  distributed    Submit a Distributed Serving Job`
)

func NewServeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "serve",
		Short: "Serve a job.",
		Long:  serveLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	command.AddCommand(NewSubmitTFServingJobCommand())
	//command.AddCommand(NewSubmitTRTServingJobCommand())
	command.AddCommand(NewSubmitCustomServingJobCommand())
	command.AddCommand(NewSubmitKFServingJobCommand())
	command.AddCommand(NewSubmitKServeJobCommand())
	command.AddCommand(NewSubmitSeldonServingJobCommand())
	command.AddCommand(NewSubmitTritonServingJobCommand())
	command.AddCommand(NewSubmitDistributedServingJobCommand())
	command.AddCommand(NewListCommand())
	command.AddCommand(NewDeleteCommand())
	command.AddCommand(NewGetCommand())
	command.AddCommand(NewAttachCommand())
	command.AddCommand(NewLogsCommand())
	command.AddCommand(NewTrafficRouterSplitCommand())
	command.AddCommand(NewUpdateCommand())

	return command
}
