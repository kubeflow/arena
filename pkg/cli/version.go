package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Arena v2\n")
		fmt.Printf("  Version:    %s\n", version)
		fmt.Printf("  Git Commit: %s\n", gitCommit)
		fmt.Printf("  Build Date: %s\n", buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
