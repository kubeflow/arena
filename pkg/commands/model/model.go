package model

import (
	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/commands/model/analyze"
)

func NewModelCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "model",
		Short: "Model manage",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewModelCreateCommand())
	command.AddCommand(NewModelGetCommand())
	command.AddCommand(NewModelListCommand())
	command.AddCommand(NewModelUpdateCommand())
	command.AddCommand(NewModelDeleteCommand())

	command.AddCommand(analyze.NewAnalyzeCommand())

	return command
}
