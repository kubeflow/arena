package cli

import (
	"github.com/spf13/cobra"
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
