package model

import (
	"github.com/kubeflow/arena/pkg/commands/model/analyze"

	"github.com/spf13/cobra"
)

func NewModelCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "model",
		Short: "Model manage and model analyze",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	command.AddCommand(analyze.NewAnalyzeCommand())
	return command
}
