package analyze

import "github.com/spf13/cobra"

func NewAnalyzeCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "analyze",
		Short: "Submit a model analyze job. (experimental feature)",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	command.AddCommand(NewSubmitModelProfileJobCommand())
	command.AddCommand(NewSubmitModelOptimizeJobCommand())
	command.AddCommand(NewSubmitModelBenchmarkJobCommand())
	command.AddCommand(NewSubmitModelEvaluateJobCommand())
	command.AddCommand(NewGetModelJobCommand())
	command.AddCommand(NewListModelJobsCommand())
	command.AddCommand(NewDeleteModelJobCommand())

	return command
}
