package model

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewModelCreateCommand() *cobra.Command {
	var name, version, description, tagsStr, versionDescription, versionTagsStr, versionSource string

	var command = &cobra.Command{
		Use:   "create",
		Short: "Create a registered model or model version",
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

			// Create a registered model if not exists
			var exists bool
			registered_model, _ := mlflowClient.GetRegisteredModel(name)
			if registered_model != nil {
				exists = true
			}

			if !exists {
				tags := []*types.RegisteredModelTag{
					{
						Key:   "createdBy",
						Value: "arena",
					},
				}
				tagsMap, err := ValidateAndParseTags(tagsStr)
				if err != nil {
					return err
				}
				for key, value := range tagsMap {
					tags = append(tags, &types.RegisteredModelTag{
						Key:   key,
						Value: value,
					})
				}
				_, err = mlflowClient.CreateRegisteredModel(name, tags, description)
				if err != nil && !strings.Contains(err.Error(), "resource already exists") {
					return err
				}
				fmt.Printf("registered model \"%s\" created\n", name)
			}

			// Create a model version
			if version != "auto" {
				return fmt.Errorf("version currently only supports `auto`")
			}
			versionTags := []*types.ModelVersionTag{
				{
					Key:   "createdBy",
					Value: "arena",
				},
			}
			versionTagsMap, err := ValidateAndParseTags(versionTagsStr)
			if err != nil {
				return err
			}
			for key, value := range versionTagsMap {
				versionTags = append(versionTags, &types.ModelVersionTag{
					Key:   key,
					Value: value,
				})
			}
			modelVersion, err := mlflowClient.CreateModelVersion(name, versionSource, "", versionTags, "", versionDescription)
			if err != nil {
				return err
			}
			fmt.Printf("model version %s for \"%s\" created\n", modelVersion.Version, modelVersion.Name)
			return nil
		},
	}

	command.Flags().StringVar(&name, "name", "", "model name")
	command.Flags().StringVar(&version, "version", "auto", "model version, available options are \"auto\"")
	command.Flags().StringVar(&description, "description", "", "model description")
	command.Flags().StringVar(&tagsStr, "tags", "", "model tags e.g. key1,key2=value2")
	command.Flags().StringVar(&versionDescription, "version-description", "", "model version description")
	command.Flags().StringVar(&versionTagsStr, "version-tags", "", "model version tags e.g. key1,key2=value2")
	command.Flags().StringVar(&versionSource, "version-source", "", "model version source")

	return command
}
