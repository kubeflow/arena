package commands

import (
	"fmt"

	"github.com/kubeflow/arena"
	"github.com/spf13/cobra"
)

func NewVersionCmd(cliName string) *cobra.Command {
	var short bool
	versionCmd := cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print version information"),
		Run: func(cmd *cobra.Command, args []string) {
			version := arena.GetVersion()
			fmt.Printf("%s: %s\n", cliName, version)
			if short {
				return
			}
			fmt.Printf("  BuildDate: %s\n", version.BuildDate)
			fmt.Printf("  GitCommit: %s\n", version.GitCommit)
			fmt.Printf("  GitTreeState: %s\n", version.GitTreeState)
			if version.GitTag != "" {
				fmt.Printf("  GitTag: %s\n", version.GitTag)
			}
			fmt.Printf("  GoVersion: %s\n", version.GoVersion)
			fmt.Printf("  Compiler: %s\n", version.Compiler)
			fmt.Printf("  Platform: %s\n", version.Platform)
		},
	}
	versionCmd.Flags().BoolVar(&short, "short", false, "print just the version number")
	return &versionCmd
}
