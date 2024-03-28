package model

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewModelGetCommand() *cobra.Command {
	var name, version string

	var command = &cobra.Command{
		Use:   "get",
		Short: "Get a registered model or model version",
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
				// Get a registered model
				registeredModel, err := mlflowClient.GetRegisteredModel(name)
				if err != nil {
					return err
				}
				filter := fmt.Sprintf("name='%s'", registeredModel.Name)
				orderBy := []string{
					"name",
					"version_number",
				}
				modelVersions, err := mlflowClient.SearchModelVersions(filter, 100, orderBy)
				if err != nil {
					return err
				}
				printRegisteredModel(registeredModel)
				printModelVersions(modelVersions)
				return nil
			} else {
				// Get a model version
				modelVersion, err := mlflowClient.GetModelVersion(name, version)
				if err != nil {
					return err
				}
				printModelVersion(modelVersion)
			}

			return nil
		},
	}

	command.Flags().StringVar(&name, "name", "", "model name")
	command.Flags().StringVar(&version, "version", "", "model version")

	return command
}
