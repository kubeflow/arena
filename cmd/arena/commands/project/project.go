package project

import (
	"github.com/spf13/cobra"
)

func NewProjectCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "project",
		Short: "project commands",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	command.AddCommand(newListProjectsCommand())
	return command
}
