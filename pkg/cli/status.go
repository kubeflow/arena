package cli

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <name>",
	Short: "Show job status (alias for get)",
	Args:  cobra.ExactArgs(1),
	RunE:  getCmd.RunE,
}

func init() {
	jobCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&getDetails, "details", false, "show job configuration details")
	statusCmd.ValidArgsFunction = completeJobName
}
