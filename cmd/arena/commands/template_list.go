package commands

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/clusterConfig"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

func NewTemplateListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "List different templates in the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			clientset, err := util.GetClientSet()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			clusterConfigs := clusterConfig.NewClusterConfigs(clientset)
			configs, err := clusterConfigs.ListClusterConfigs()

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			PrintTemplates(configs)
		},
	}

	return command
}

func PrintTemplates(configs []clusterConfig.ClusterConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "DESCRIPTION"}

	PrintLine(w, labelField...)

	for _, config := range configs {
		configName := config.Name
		if config.IsDefault {
			configName = fmt.Sprintf("%s (default)", config.Name)
		}
		PrintLine(w, configName, config.Description)
	}

	w.Flush()
}
