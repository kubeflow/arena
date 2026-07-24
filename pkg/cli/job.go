package cli

import (
	"github.com/spf13/cobra"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var (
	outputFormat string
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage training jobs",
	Long:  `Commands for submitting, listing, inspecting, and managing training jobs.`,
}

func init() {
	jobCmd.PersistentFlags().StringVarP(
		&outputFormat,
		"output",
		"o",
		string(outputpkg.DefaultFormat),
		outputpkg.FormatHelpText,
	)
	_ = jobCmd.RegisterFlagCompletionFunc("output", completeOutputFormat)
	rootCmd.AddCommand(jobCmd)
}
