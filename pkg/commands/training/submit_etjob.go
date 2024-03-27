package training

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util/kubectl"
)

func NewSubmitETJobCommand() *cobra.Command {
	builder := training.NewETJobBuilder()
	var command = &cobra.Command{
		Use:     "etjob",
		Short:   "Submit ETJob as training job.",
		Aliases: []string{"et"},
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlags(cmd.Flags())
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
				return fmt.Errorf("failed to create arena client: %v", err)
			}
			job, err := builder.Command(args).Build()
			if err != nil {
				return fmt.Errorf("failed to validate command args: %v", err)
			}
			if err := client.Training().Submit(job); err != nil {
				return err
			}
			fullSubmitCommand := getFullSubmitCommand(cmd, args)
			_, modelVersion, err := createRegisteredModelAndModelVersion(client, job, fullSubmitCommand)
			if modelVersion == nil {
				return err
			}
			if err := kubectl.AddTrainingJobLabel(job, "modelVersion", modelVersion.Version); err != nil {
				return fmt.Errorf("failed to patch label `modelVersion=%s` to job %s/%s: %v", modelVersion.Version, job.Type(), job.Name(), err)
			}
			return nil
		},
	}
	builder.AddCommandFlags(command)
	return command
}
