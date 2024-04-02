package model

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
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

			if version == "" {
				// Delete a registered model and all of its model versions
				if !force {
					prompt := "Delete a registered model will cascade delete all its model versions. Are you sure you want to perform this operation? (yes/no)"
					confirm := utils.ReadUserConfirmation(prompt)
					if !confirm {
						fmt.Println("Cancelled.")
						return nil
					}
				}
				if err := modelClient.DeleteRegisteredModel(name); err != nil {
					return err
				}
				log.Infof("registered model \"%s\" deleted\n", name)
				return nil
			} else {
				// Delete a model version
				if !force {
					prompt := "Are you sure you want to perform this operation? (yes/no)"
					confirm := utils.ReadUserConfirmation(prompt)
					if !confirm {
						fmt.Println("Cancelled.")
						return nil
					}
				}
				if err := modelClient.DeleteModelVersion(name, version); err != nil {
					return err
				}
				log.Infof("model version \"%s/%s\" deleted\n", name, version)
			}
			return nil
		},
	}
	command.Flags().StringVar(&name, "name", "", "model name")
	_ = command.MarkFlagRequired("name")
	command.Flags().StringVar(&version, "version", "", "model version")
	command.Flags().BoolVarP(&force, "force", "f", false, "If true, delete resources without confirmation")
	return command
}
