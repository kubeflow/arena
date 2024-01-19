package serving

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/serving"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUpdateCustomCommand update a custom serving
func NewUpdateCustomCommand() *cobra.Command {
	builder := serving.NewUpdateCustomServingJobBuilder()
	var command = &cobra.Command{
		Use:   "custom",
		Short: "Update a custom serving job and its associated instances",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
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
				return err
			}

			job, err := builder.Namespace(config.GetArenaConfiger().GetNamespace()).Command(args).Build()
			if err != nil {
				return fmt.Errorf("failed to validate command args: %v", err)
			}
			return client.Serving().Update(job)
		},
	}

	builder.AddCommandFlags(command)
	return command
}
