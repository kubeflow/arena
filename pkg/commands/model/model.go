package model

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/commands/model/analyze"
)

const (
	MLFLOW_TRACKING_URI      = "MLFLOW_TRACKING_URI"
	MLFLOW_TRACKING_USERNAME = "MLFLOW_TRACKING_USERNAME"
	MLFLOW_TRACKING_PASSWORD = "MLFLOW_TRACKING_PASSWORD"
)

var (
	mlflowTrackingUri      string
	mlflowTrackingUsername string
	mlflowTrackingPassword string
)

func NewModelCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "model",
		Short: "Model manage",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if mlflowTrackingUri == "" {
				mlflowTrackingUri = os.Getenv(MLFLOW_TRACKING_URI)
			}
			if mlflowTrackingUsername == "" {
				mlflowTrackingUsername = os.Getenv(MLFLOW_TRACKING_USERNAME)
			}
			if mlflowTrackingPassword == "" {
				mlflowTrackingPassword = os.Getenv(MLFLOW_TRACKING_PASSWORD)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewModelCreateCommand())
	command.AddCommand(NewModelGetCommand())
	command.AddCommand(NewModelListCommand())
	command.AddCommand(NewModelUpdateCommand())
	command.AddCommand(NewModelDeleteCommand())

	command.AddCommand(analyze.NewAnalyzeCommand())

	return command
}
