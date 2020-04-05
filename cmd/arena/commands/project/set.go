package project

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util/command"
	"github.com/spf13/cobra"
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

	err = kubeClient.SetDefaultNamespace(project)

	if err != nil {
		return err
	} else {
		fmt.Printf("Project %s has been set as default project\n", project)
		return nil
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
