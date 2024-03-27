package model

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewModelUpdateCommand() *cobra.Command {
	var name, version, description, tagsStr, versionDescription, versionTagsStr string
	var command = &cobra.Command{
		Use:   "update",
		Short: "Update a registered model or model version",
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

			// Update registered model
			if description != "" || tagsStr != "" {
				if description != "" {
					_, err := mlflowClient.UpdateRegisteredModel(name, description)
					if err != nil {
						return err
					}
				}
				tagsMap, err := ValidateAndParseTags(tagsStr)
				if err != nil {
					return err
				}
				for key, value := range tagsMap {
					if strings.HasSuffix(value, "-") {
						if err := mlflowClient.DeleteRegisteredModelTag(name, key); err != nil {
							return err
						}
					} else if value == "" && strings.HasSuffix(key, "-") {
						if err := mlflowClient.DeleteRegisteredModelTag(name, key[0:len(key)-1]); err != nil {
							return err
						}
					} else {
						if err := mlflowClient.SetRegisteredModelTag(name, key, value); err != nil {
							return err
						}
					}
				}
				fmt.Printf("registered model \"%s\" updated\n", name)
			}

			// Update model version
			if version != "" {
				if versionDescription != "" {
					_, err := mlflowClient.UpdateModelVersion(name, version, versionDescription)
					if err != nil {
						return err
					}
				}
				versionTagsMap, err := ValidateAndParseTags(versionTagsStr)
				if err != nil {
					return err
				}
				for key, value := range versionTagsMap {
					if strings.HasSuffix(value, "-") {
						if err := mlflowClient.DeleteModelVersionTag(name, version, key); err != nil {
							return err
						}
					} else if value == "" && strings.HasSuffix(key, "-") {
						if err := mlflowClient.DeleteModelVersionTag(name, version, key[0:len(key)-1]); err != nil {
							return err
						}
					} else {
						if err := mlflowClient.SetModelVersionTag(name, version, key, value); err != nil {
							return err
						}
					}
				}
				fmt.Printf("model version \"%s/%s\" updated\n", name, version)
				return nil
			}
			return nil
		},
	}
	command.Flags().StringVar(&name, "name", "", "model name")
	command.Flags().StringVar(&version, "version", "", "model version")
	command.Flags().StringVar(&description, "description", "", "model description")
	command.Flags().StringVar(&tagsStr, "tags", "", "model tags")
	command.Flags().StringVar(&versionDescription, "version-description", "", "model version description")
	command.Flags().StringVar(&versionTagsStr, "version-tags", "", "model version tags")
	return command
}
