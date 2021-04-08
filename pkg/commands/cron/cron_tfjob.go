package cron

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/cron"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCronTFJobCommand() *cobra.Command {
	builder := cron.NewCronTFJobBuilder()
	var command = &cobra.Command{
		Use:     "tfjob",
		Short:   "Submit a cron TFJob.",
		Aliases: []string{"tf"},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				return fmt.Errorf("not found command args")
			}
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
			job, err := builder.Command(args).Build()
			if err != nil {
				return fmt.Errorf("failed to validate command args: %v", err)
			}
			return client.Cron().SubmitCronTrainingJob(job)
		},
	}
	builder.AddCommandFlags(command)
	return command
}
