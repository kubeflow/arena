package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version      = "dev"
	gitCommit    = "unknown"
	buildDate    = "unknown"
	gitTag       = ""
	gitTreeState = "unknown"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Arena v2")
			fmt.Fprintf(out, "  Version:     %s\n", version)
			fmt.Fprintf(out, "  Git Commit:  %s\n", gitCommit)
			fmt.Fprintf(out, "  Git Tag:     %s\n", gitTag)
			fmt.Fprintf(out, "  Build Date:  %s\n", buildDate)
			fmt.Fprintf(out, "  Tree State:  %s\n", gitTreeState)
			return nil
		},
	}
	cmd.ValidArgsFunction = cobra.NoFileCompletions
	return cmd
}
