package commands

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

			execute(cmd, name, "/bin/bash", []string{}, true, true)
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
		Use:   "exec",
		Short: "execute a command inside a running job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			name = args[0]
			command := args[1]
			commandArgs := args[2:]

			execute(cmd, name, command, commandArgs, interactive, TTY)
		},
	}

	command.Flags().BoolVarP(&interactive, "stdin", "i", false, "Pass stdin to the container")
	command.Flags().BoolVarP(&TTY, "tty", "t", false, "Stdin is a TTY")

	return command
}

func execute(cmd *cobra.Command, name string, command string, commandArgs []string, interactive bool, TTY bool) {

	util.SetLogLevel(logLevel)
	_, err := initKubeClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = updateNamespace(cmd)
	if err != nil {
		log.Debugf("Failed due to %v", err)
		fmt.Println(err)
		os.Exit(1)
	}

	job, err := searchTrainingJob(name, "", namespace)
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	kubectl.Exec(job.ChiefPod().Name, job.ChiefPod().Namespace, command, commandArgs, interactive, TTY)
}
