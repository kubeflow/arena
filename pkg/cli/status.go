package cli

import (
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Show job status (alias for get)",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}
	cmd.Flags().BoolVar(&getDetails, "details", false, "show job configuration details")
	registerOutputFlag(cmd)
	cmd.ValidArgsFunction = completeJobName
	return cmd
}
