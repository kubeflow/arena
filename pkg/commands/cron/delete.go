package cron

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCronDeleteCommand
func NewCronDeleteCommand() *cobra.Command {
	var trainingType string
	var command = &cobra.Command{
		Use:     "delete JOB1 JOB2 ...JOBn [-T JOB_TYPE]",
		Short:   "Delete a cron job and its associated instances",
		Aliases: []string{"del"},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return fmt.Errorf("not set job name,please set it")
			}
			names := args
			client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
				Kubeconfig:     viper.GetString("config"),
				LogLevel:       viper.GetString("loglevel"),
				Namespace:      viper.GetString("namespace"),
				ArenaNamespace: viper.GetString("arena-namespace"),
				IsDaemonMode:   false,
			})
			if err != nil {
				return err
			}
			return client.Cron().Delete(names...)
		},
	}
	command.Flags().StringVarP(&trainingType, "type", "T", "", fmt.Sprintf("The training type to delete, the possible option is %v. (optional)", utils.GetSupportTrainingJobTypesInfo()))

	return command
}
