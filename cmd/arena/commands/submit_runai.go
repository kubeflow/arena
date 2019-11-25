package commands

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	runaiChart = util.GetChartsFolder() + "/runai"
)

const (
	defaultRunaiTrainingType = "runai"
)

func NewRunaiJobCommand() *cobra.Command {
	submitArgs := NewSubmitRunaiJobArgs()
	var command = &cobra.Command{
		Use:     "runai",
		Short:   "Submit a Runai job.",
		Aliases: []string{"ra"},
		Run: func(cmd *cobra.Command, args []string) {

			util.SetLogLevel(logLevel)
			setupKubeconfig()

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

			err = submitRunaiJob(args, submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	submitArgs.addFlags(command)

	return command
}

func NewSubmitRunaiJobArgs() *submitRunaiJobArgs {
	return &submitRunaiJobArgs{}
}

type submitRunaiJobArgs struct {
	Project string `yaml:"project"`
	GPUS    int    `yaml:"gpus"`
	Image   string `yaml:"image"`
	Name    string `yaml:"name"`
	HostIPC bool   `yaml:"hostIPC"`
}

// add flags to submit spark args
func (sa *submitRunaiJobArgs) addFlags(command *cobra.Command) {
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")

	command.Flags().IntVar(&(sa.GPUS), "gpus", 1, "Number of GPUs the job requires.")
	command.Flags().StringVar(&(sa.Project), "project", "default", "Specifies the project to use for this job, leave empty to use default project")
	command.Flags().StringVar(&(sa.Image), "image", "", "Specifies job image")
	command.Flags().BoolVar(&(sa.HostIPC), "host-ipc", false, "Use the host's ipc namespace. Optional: Default to false.")
	command.MarkFlagRequired("image")
}

func submitRunaiJob(args []string, submitArgs *submitRunaiJobArgs) error {

	err := workflow.SubmitJob(name, defaultRunaiTrainingType, namespace, submitArgs, runaiChart)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", name, defaultRunaiTrainingType)
	return nil
}
