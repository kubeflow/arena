package serving

import "github.com/spf13/cobra"

var (
	serveUpdateLong = `update a serve job.

Available Commands:
  tensorflow,tf  Update a TensorFlow Serving Job
  triton         Update a Nvidia Triton Serving Job
  custom         Update a Custom Serving Job`
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

	return command
}
