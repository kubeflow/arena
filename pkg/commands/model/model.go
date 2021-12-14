package model

import "github.com/spf13/cobra"

var (
	serveLong = `submit a model analyze job.

Available Commands:
  profile          Submit a model profile job
  optimize         Submit a model optimize job.
  benchmark        Submit a model benchmark job`
)

func NewModelCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "model",
		Short: "Submit a model analyze job. (experimental feature)",
		Long:  serveLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	command.AddCommand(NewSubmitModelProfileJobCommand())
	command.AddCommand(NewSubmitModelOptimizeJobCommand())
	command.AddCommand(NewSubmitModelBenchmarkJobCommand())
	command.AddCommand(NewGetModelJobCommand())
	command.AddCommand(NewListModelJobsCommand())
	command.AddCommand(NewDeleteModelJobCommand())

	return command
}
