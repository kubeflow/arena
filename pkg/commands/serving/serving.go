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
  seldon         Submit a Seldon Serving Job`
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
	command.AddCommand(NewListCommand())
	command.AddCommand(NewDeleteCommand())
	command.AddCommand(NewGetCommand())
	command.AddCommand(NewAttachCommand())
	command.AddCommand(NewLogsCommand())
	command.AddCommand(NewTrafficRouterSplitCommand())
	command.AddCommand(NewUpdateCommand())

	return command
}
