package model

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewModelListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:     "list",
		Short:   "List all registered models",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())

			if mlflowTrackingUri == "" {
				return fmt.Errorf("MLflow tracking server URI must be specified by MLFLOW_TRACKING_URI environment variable")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mlflowClient := arenaclient.NewMlflowClient(mlflowTrackingUri, mlflowTrackingUsername, mlflowTrackingPassword)
			registeredModels, err := mlflowClient.SearchRegisteredModels("", 100, []string{})
			if err != nil {
				return err
			}
			printRegisteredModels(registeredModels)
			return nil
		},
	}
	return command
}
