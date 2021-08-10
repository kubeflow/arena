package evaluate

import (
	"github.com/spf13/cobra"
)

var (
	dataLong = `manage evaluate job.

Available Commands:
  model                Submit a evaluate job.
  list,ls              List the evaluate job.
  get                  Get evaluate job by name.
  delete,del           Delete evaluate job by name.
`
)


// NewEvaluateCommand manage evaluate job
func NewEvaluateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "evaluate",
		Short: "manage evaluate job.",
		Long:  dataLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewEvaluateModelCommand())
	command.AddCommand(NewEvaluateDeleteCommand())
	command.AddCommand(NewEvaluateListCommand())
	command.AddCommand(NewEvaluateGetCommand())

	return command
}