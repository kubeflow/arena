package flags

import (
	"github.com/spf13/cobra"
)

func GetProjectFlag(cmd *cobra.Command) string {
	return getFlagValue(cmd, ProjectFlag)
}

func getFlagValue(cmd *cobra.Command, name string) string {
	return cmd.Flags().Lookup(name).Value.String()
}
