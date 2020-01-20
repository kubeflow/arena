package commands

import (
	"github.com/spf13/cobra"
)

func NewEnvironmentCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "env",
		Short: "Information about different environments in the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	command.AddCommand(NewEnvironmentListCommand())
	command.AddCommand(NewEnvironmentGetCommand())

	return command
}
