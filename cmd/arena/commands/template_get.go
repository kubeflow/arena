package commands

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/clusterConfig"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/spf13/cobra"
	"os"
)

func NewTemplateGetCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "get",
		Short: "Get information on one of the templates in the system",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(0)
			}

			clientset, err := util.GetClientSet()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			clusterConfigs := clusterConfig.NewClusterConfigs(clientset)
			configName := args[0]
			config, err := clusterConfigs.GetClusterConfig(configName)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if config == nil {
				fmt.Printf("Template '%s' not found\n", configName)
				os.Exit(1)
			}

			fmt.Printf("Name: %s\n", configName)
			fmt.Printf("Description: %s\n\n", config.Description)
			fmt.Println("Values:")
			fmt.Println("---------------------------")
			fmt.Println(config.Values)
		},
	}

	return command
}
