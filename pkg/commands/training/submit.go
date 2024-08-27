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

import "github.com/spf13/cobra"

var (
	submitLong = `Submit a job.

Available Commands:
  tfjob,tf             Submit a TFJob.
  pytorchjob,pytorch   Submit a PyTorchJob.
  mpijob,mpi           Submit a MPIJob.
  etjob,et             Submit a ETJob.
  horovod,hj           Submit a Horovod Job.
  volcanojob,vj        Submit a VolcanoJob.
  rayjob,rj            Submit a RayJob.
    `
)

func NewSubmitCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "submit",
		Short: "Submit a training job.",
		Long:  submitLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	command.AddCommand(NewSubmitTFJobCommand())
	command.AddCommand(NewSubmitMPIJobCommand())
	command.AddCommand(NewSubmitPytorchJobCommand())
	command.AddCommand(NewSubmitHorovodJobCommand())
	// Warning: Spark is not work,skip it
	command.AddCommand(NewSubmitSparkJobCommand())
	command.AddCommand(NewVolcanoJobCommand())
	command.AddCommand(NewSubmitETJobCommand())
	command.AddCommand(NewSubmitDeepSpeedJobCommand())
	command.AddCommand(NewSubmitRayJobCommand())
	return command
}
