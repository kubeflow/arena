package commands

import (
	"github.com/spf13/cobra"
)

func NewTemplateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "template",
		Short: "Information about different templates in the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	command.AddCommand(NewTemplateListCommand())
	command.AddCommand(NewTemplateGetCommand())

	return command
}
