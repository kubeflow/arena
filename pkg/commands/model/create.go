package model

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
)

func NewModelCreateCommand() *cobra.Command {
	var name, version, description, tagsStr, versionDescription, versionTagsStr, source string

	var command = &cobra.Command{
		Use:   "create",
		Short: "Create a registered model or model version",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
				Kubeconfig:     viper.GetString("config"),
				LogLevel:       viper.GetString("loglevel"),
				Namespace:      viper.GetString("namespace"),
				ArenaNamespace: viper.GetString("arena-namespace"),
				IsDaemonMode:   false,
			})
			if err != nil {
				return fmt.Errorf("failed to create arena client: %v", err)
			}

			modelClient, err := client.Model()
			if err != nil {
				return fmt.Errorf("failed to create arena model client: %v", err)
			}

			// Create a registered model if not exists
			var exists bool
			registered_model, _ := modelClient.GetRegisteredModel(name)
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
				tagsMap, err := utils.ValidateAndParseTags(tagsStr)
				if err != nil {
					return err
				}
				for key, value := range tagsMap {
					tags = append(tags, &types.RegisteredModelTag{
						Key:   key,
						Value: value,
					})
				}
				_, err = modelClient.CreateRegisteredModel(name, tags, description)
				if err != nil && !strings.Contains(err.Error(), "resource already exists") {
					return err
				}
				log.Infof("registered model \"%s\" created\n", name)
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
			versionTagsMap, err := utils.ValidateAndParseTags(versionTagsStr)
			if err != nil {
				return err
			}
			for key, value := range versionTagsMap {
				versionTags = append(versionTags, &types.ModelVersionTag{
					Key:   key,
					Value: value,
				})
			}
			modelVersion, err := modelClient.CreateModelVersion(name, source, "", versionTags, "", versionDescription)
			if err != nil {
				return err
			}
			log.Infof("model version %s for \"%s\" created\n", modelVersion.Version, modelVersion.Name)
			return nil
		},
	}

	command.Flags().StringVar(&name, "name", "", "model name")
	_ = command.MarkFlagRequired("name")
	command.Flags().StringVar(&version, "version", "auto", "model version, available options are \"auto\"")
	command.Flags().StringVar(&description, "description", "", "model description")
	command.Flags().StringVar(&tagsStr, "tags", "", "model tags e.g. key1,key2=value2")
	command.Flags().StringVar(&versionDescription, "version-description", "", "model version description")
	command.Flags().StringVar(&versionTagsStr, "version-tags", "", "model version tags e.g. key1,key2=value2")
	command.Flags().StringVar(&source, "source", "", "model version source")
	_ = command.MarkFlagRequired("source")

	return command
}
