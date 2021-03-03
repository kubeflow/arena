package serving

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/serving"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSubmitCustomServingJobCommand() *cobra.Command {
	builder := serving.NewCustomServingJobBuilder()
	var command = &cobra.Command{
		Use:     "custom",
		Short:   "Submit custom serving to deploy and serve machine learning models.",
		Aliases: []string{"custom"},
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

			job, err := builder.Namespace(config.GetArenaConfiger().GetNamespace()).Command(args).Build()
			if err != nil {
				return fmt.Errorf("failed to validate command args: %v", err)
			}
			return client.Serving().Submit(job)
		},
	}
	builder.AddCommandFlags(command)
	return command
}
