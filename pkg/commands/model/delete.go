package model

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDeleteModelJobCommand() *cobra.Command {
	var jobType string
	var command = &cobra.Command{
		Use:   "delete",
		Short: "Delete a model job",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
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
			return client.Model().Delete(utils.TransferModelJobType(jobType), names...)
		},
	}

	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The model job type to delete, the possible option is [%v]. (optional)", utils.GetSupportModelJobTypesInfo()))
	return command
}
