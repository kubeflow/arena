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
	return command
}
