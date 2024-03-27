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

func NewModelUpdateCommand() *cobra.Command {
	var name, version, description, tagsStr, versionDescription, versionTagsStr string
	var command = &cobra.Command{
		Use:   "update",
		Short: "Update a registered model or model version",
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

			// Update registered model
			if description != "" || tagsStr != "" {
				if description != "" {
					_, err := modelClient.UpdateRegisteredModel(name, description)
					if err != nil {
						return err
					}
				}
				tagsMap, err := utils.ValidateAndParseTags(tagsStr)
				if err != nil {
					return err
				}
				for key, value := range tagsMap {
					if strings.HasSuffix(value, "-") {
						if err := modelClient.DeleteRegisteredModelTag(name, key); err != nil {
							return err
						}
					} else if value == "" && strings.HasSuffix(key, "-") {
						if err := modelClient.DeleteRegisteredModelTag(name, key[0:len(key)-1]); err != nil {
							return err
						}
					} else {
						if err := modelClient.SetRegisteredModelTag(name, key, value); err != nil {
							return err
						}
					}
				}
				log.Infof("registered model \"%s\" updated\n", name)
			}

			// Update model version
			if version != "" {
				if versionDescription != "" {
					_, err := modelClient.UpdateModelVersion(name, version, versionDescription)
					if err != nil {
						return err
					}
				}
				versionTagsMap, err := utils.ValidateAndParseTags(versionTagsStr)
				if err != nil {
					return err
				}
				for key, value := range versionTagsMap {
					if strings.HasSuffix(value, "-") {
						if err := modelClient.DeleteModelVersionTag(name, version, key); err != nil {
							return err
						}
					} else if value == "" && strings.HasSuffix(key, "-") {
						if err := modelClient.DeleteModelVersionTag(name, version, key[0:len(key)-1]); err != nil {
							return err
						}
					} else {
						if err := modelClient.SetModelVersionTag(name, version, key, value); err != nil {
							return err
						}
					}
				}
				log.Infof("model version \"%s/%s\" updated\n", name, version)
				return nil
			}
			return nil
		},
	}
	command.Flags().StringVar(&name, "name", "", "model name")
	_ = command.MarkFlagRequired("name")
	command.Flags().StringVar(&version, "version", "", "model version")
	command.Flags().StringVar(&description, "description", "", "model description")
	command.Flags().StringVar(&tagsStr, "tags", "", "model tags")
	command.Flags().StringVar(&versionDescription, "version-description", "", "model version description")
	command.Flags().StringVar(&versionTagsStr, "version-tags", "", "model version tags")
	return command
}
