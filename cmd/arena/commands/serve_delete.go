package commands

import (
	"os"

	"github.com/kubeflow/arena/util/helm"
	"github.com/spf13/cobra"
)

// NewDeleteCommand
func NewServingDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "delete a serving job",
		Short: "delete a serving job and its associated pods",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			setupKubeconfig()
			for _, jobName := range args {
				deleteServingJob(jobName)
			}
		},
	}

	return command
}

func deleteServingJob(servingJob string) error {
	return helm.DeleteRelease(servingJob)
}
