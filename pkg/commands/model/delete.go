package model

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewModelDeleteCommand() *cobra.Command {
	var name, version string
	var force bool
	var command = &cobra.Command{
		Use:   "delete",
		Short: "Delete a registered model or model version",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())

			if mlflowTrackingUri == "" {
				return fmt.Errorf("MLflow tracking server URI must be specified by MLFLOW_TRACKING_URI environment variable")
			}
			if name == "" {
				return fmt.Errorf("model name must be specified by --name flag")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mlflowClient := arenaclient.NewMlflowClient(mlflowTrackingUri, mlflowTrackingUsername, mlflowTrackingPassword)

			if version == "" {
				// Delete a registered model and all of its model versions
				if !force {
					prompt := "Delete a registered model will cascade delete all its model versions. Are you sure you want to perform this operation? (yes/no)"
					confirm := readUserConfirmation(prompt)
					if !confirm {
						fmt.Println("Cancelled.")
						return nil
					}
				}
				if err := mlflowClient.DeleteRegisteredModel(name); err != nil {
					return err
				}
				fmt.Printf("registered model \"%s\" deleted\n", name)
				return nil
			} else {
				// Delete a model version
				if !force {
					prompt := "Are you sure you want to perform this operation? (yes/no)"
					confirm := readUserConfirmation(prompt)
					if !confirm {
						fmt.Println("Cancelled.")
						return nil
					}
				}
				if err := mlflowClient.DeleteModelVersion(name, version); err != nil {
					return err
				}
				fmt.Printf("model version \"%s/%s\" deleted\n", name, version)
			}
			return nil
		},
	}
	command.Flags().StringVar(&name, "name", "", "model name")
	command.Flags().StringVar(&version, "version", "", "model version")
	command.Flags().BoolVarP(&force, "force", "f", false, "If true, delete resources without confirmation")
	return command
}
