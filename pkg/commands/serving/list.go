package serving

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewListCommand() *cobra.Command {
	var allNamespaces bool
	var format string
	var servingType string
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the serving jobs",
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
				return fmt.Errorf("failed to create arena client: %v", err)
			}
			return client.Serving().ListAndPrint(allNamespaces, utils.TransferServingJobType(servingType), format)
		},
	}
	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")
	command.Flags().MarkDeprecated("allNamespaces", "please use --all-namespaces instead")
	command.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "show all the namespaces")
	command.Flags().StringVarP(&format, "output", "o", "wide", "Output format. One of: json|yaml|wide")
	command.Flags().StringVarP(&servingType, "type", "T", "", fmt.Sprintf("The serving type, the possible option is [%v]. (optional)", utils.GetSupportServingJobTypesInfo()))
	return command
}
