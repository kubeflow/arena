package cli

import (
	"github.com/spf13/cobra"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var (
	outputFormat string
)

func newJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage training jobs",
		Long:  `Commands for submitting, listing, inspecting, and managing training jobs.`,
	}
	cmd.PersistentFlags().StringVarP(
		&outputFormat,
		"output",
		"o",
		string(outputpkg.DefaultFormat),
		outputpkg.FormatHelpText,
	)
	_ = cmd.RegisterFlagCompletionFunc("output", completeOutputFormat)

	cmd.AddCommand(
		newRunCmd(),
		newGetCmd(),
		newStatusCmd(),
		newListCmd(),
		newLogsCmd(),
		newDeleteCmd(),
		newSuspendCmd(),
		newResumeCmd(),
	)
	return cmd
}
