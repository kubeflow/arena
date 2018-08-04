package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/arena/util/helm"
	"github.com/spf13/cobra"
)

func NewSubmitHorovodJobCommand() *cobra.Command {
	var (
		submitArgs submitHorovodJobArgs
	)

	var command = &cobra.Command{
		Use:     "horovodjob",
		Short:   "Submit horovodjob as training job.",
		Aliases: []string{"hj"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = ensureNamespace(client, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitHorovodJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().StringVar(&submitArgs.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")

	submitArgs.addCommonFlags(command)
	submitArgs.addSyncFlags(command)

	command.Flags().IntVar(&submitArgs.SSHPort, "sshPort", 33,
		"ssh port.")
	return command
}

type submitHorovodJobArgs struct {
	SSHPort int    `yaml:"sshPort"` // --sshPort
	Cpu     string `yaml:"cpu"`     // --cpu
	Memory  string `yaml:"memory"`  // --memory

	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	submitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	submitSyncCodeArgs `yaml:",inline"`
}

func (submitArgs *submitHorovodJobArgs) prepare(args []string) (err error) {
	submitArgs.Command = strings.Join(args, " ")

	err = submitArgs.check()
	if err != nil {
		return err
	}

	commonArgs := &submitArgs.submitArgs
	err = commonArgs.transform()
	if err != nil {
		return nil
	}

	if len(envs) > 0 {
		submitArgs.Envs = transformSliceToMap(envs, "=")
	}

	submitArgs.addHorovodInfoToEnv()

	return nil
}

func (submitArgs submitHorovodJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	if submitArgs.Image == "" {
		return fmt.Errorf("--image must be set ")
	}

	return nil
}

func (submitArgs *submitHorovodJobArgs) addHorovodInfoToEnv() {
	submitArgs.addJobInfoToEnv()
}

func submitHorovodJob(args []string, submitArgs *submitHorovodJobArgs) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	exist, err := helm.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	// the master is also considered as a worker
	submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	return helm.InstallRelease(name, namespace, submitArgs, horovod_training_chart)
}
