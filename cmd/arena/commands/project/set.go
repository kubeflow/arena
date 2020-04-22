package project

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/util/command"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runSetCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.HelpFunc()(cmd, args)
		return nil
	} else if len(args) > 1 {
		return fmt.Errorf("Accepts 1 argument, received %d", len(args))
	}

	project := args[0]
	kubeClient, err := client.GetClient()

	if err != nil {
		return err
	}

	namespaceList, err := kubeClient.GetClientset().CoreV1().Namespaces().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", RUNAI_QUEUE_LABEL, project),
	})

	if err != nil {
		return err
	}

	if namespaceList != nil && len(namespaceList.Items) != 0 {
		err = kubeClient.SetDefaultNamespace(namespaceList.Items[0].Name)
		if err != nil {
			return err
		} else {
			fmt.Printf("Project %s has been set as default project\n", project)
			return nil
		}
	} else {
		return fmt.Errorf("project %s was not found. Please run '%s project list' to view all avaliable projects", project, config.CLIName)
	}
}

func newSetProjectCommand() *cobra.Command {
	commandWrapper := command.NewCommandWrapper(runSetCommand)
	var command = &cobra.Command{
		Use:   "set [PROJECT]",
		Short: "Set default project",
		Run:   commandWrapper.Run,
	}

	return command
}
