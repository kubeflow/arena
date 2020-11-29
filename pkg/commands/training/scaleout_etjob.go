package training

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewScaleOutETJobCommand() *cobra.Command {
	builder := training.NewScaleOutETJobBuilder()
	var command = &cobra.Command{
		Use:     "etjob",
		Short:   "scalein a etjob",
		Aliases: []string{"et"},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
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
				return fmt.Errorf("failed to create arena client: %v\n", err)
			}
			job, err := builder.Build()
			if err != nil {
				return fmt.Errorf("failed to validate command args: %v", err)
			}
			return client.Training().ScaleOut(job)
		},
	}
	builder.AddCommandFlags(command)
	return command
}
