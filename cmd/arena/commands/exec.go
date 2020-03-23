package commands

import (
	"fmt"
	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"os"
)

func NewBashCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "bash",
		Short: "get a bash session inside a running job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			name = args[0]

			execute(cmd, name, "/bin/bash", []string{}, true, true, "bash")
		},
	}

	return command
}

func NewExecCommand() *cobra.Command {
	var (
		interactive bool
		TTY         bool
	)

	var command = &cobra.Command{
		Use:   "exec NAME COMMAND [ARG ...]",
		Short: "execute a command inside a running job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			name = args[0]
			command := args[1]
			commandArgs := args[2:]

			execute(cmd, name, command, commandArgs, interactive, TTY, "exec")
		},
	}

	command.Flags().BoolVarP(&interactive, "stdin", "i", false, "Pass stdin to the container")
	command.Flags().BoolVarP(&TTY, "tty", "t", false, "Stdin is a TTY")

	return command
}

func execute(cmd *cobra.Command, name string, command string, commandArgs []string, interactive bool, TTY bool, runaiCommandName string) {

	kubeClient, err := client.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	clientset = kubeClient.GetClientset()
	namespace := flags.GetProjectFlag(cmd, kubeClient)

	job, err := searchTrainingJob(clientset, name, "", namespace)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	chiefPod := job.ChiefPod()

	if chiefPod == nil || chiefPod.Status.Phase != v1.PodRunning {
		log.Errorf("Cannot %s into the job. It is still pending execution.", runaiCommandName)
		os.Exit(1)
	}

	kubectl.Exec(chiefPod.Name, chiefPod.Namespace, command, commandArgs, interactive, TTY)
}
