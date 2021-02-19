package training

import (
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewListCommand() *cobra.Command {
	var allNamespaces bool
	var format string
	var jobType string
	var command = &cobra.Command{
		Use:     "list",
		Short:   "List all the training jobs",
		Aliases: []string{"ls"},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()
			defer func() {
				log.Debugf("execute time of listing training jobs: %v\n", time.Now().Sub(now))
			}()
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
			return client.Training().ListAndPrint(allNamespaces, format, utils.TransferTrainingJobType(jobType))
		},
	}
	command.Flags().StringVarP(&jobType, "type", "T", "", fmt.Sprintf("The training type to list, the possible option is %v. (optional)", utils.GetSupportTrainingJobTypesInfo()))
	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")
	command.Flags().MarkDeprecated("allNamespaces", "please use --all-namespaces instead")
	command.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "show all the namespaces")
	command.Flags().StringVarP(&format, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	return command
}
