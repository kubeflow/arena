package commands

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/user"
	"strings"
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

			if submitArgs.IsJupiter {
				submitArgs.UseJupiterDefaultValues()
			}

			err = submitRunaiJob(args, submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if submitArgs.ServiceType == "portforward" {
				localPorts := []string{}
				for _, port := range submitArgs.Ports {
					split := strings.Split(port, ":")
					localPorts = append(localPorts, split[0])
				}

				err = kubectl.WaitForReadyStatefulSet(name, namespace)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = kubectl.PortForward(localPorts, name, namespace)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
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
	Project     string   `yaml:"project"`
	GPU         int      `yaml:"gpu"`
	Image       string   `yaml:"image"`
	HostIPC     bool     `yaml:"hostIPC"`
	Interactive bool     `yaml:"interactive"`
	Volumes     []string `yaml:"volumes"`
	NodeType    string   `yaml:"node_type"`
	User        string   `yaml:"user"`
	Ports       []string `yaml:"ports"`
	ServiceType string   `yaml:"serviceType"`
	Command     []string `yaml:"command"`
	Args        []string `yaml:"args"`
	IsJupiter   bool
}

func (sa *submitRunaiJobArgs) UseJupiterDefaultValues() {
	var (
		jupiterPort    = "8888"
		jupiterImage   = "jupyter/scipy-notebook"
		jupiterCommand = "start-notebook.sh"
		jupiterArgs    = "--NotebookApp.base_url=/%s"
	)

	if len(sa.Ports) == 0 {
		sa.Ports = []string{jupiterPort}
		log.Infof("Expose default jupiter notebook port %s", jupiterPort)
	}
	if sa.Image == "" {
		sa.Image = "jupyter/scipy-notebook"
		log.Infof("Use default jupiter notebook image \"%s\"", jupiterImage)
	}
	if len(sa.Command) == 0 && sa.ServiceType == "ingress" {
		sa.Command = []string{jupiterCommand}
		log.Infof("Use default jupiter notebook command for using ingress service \"%s\"", jupiterCommand)
	}
	if len(sa.Args) == 0 && sa.ServiceType == "ingress" {
		baseUrlArg := fmt.Sprintf(jupiterArgs, name)
		sa.Args = []string{baseUrlArg}
		log.Infof("Use default jupiter notebook arg for using ingress service \"%s\"", baseUrlArg)
	}
}

// add flags to submit spark args
func (sa *submitRunaiJobArgs) addFlags(command *cobra.Command) {
	currentUser, _ := user.Current()
	defaultUser := currentUser.Username

	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")

	command.Flags().IntVarP(&(sa.GPU), "gpu", "g", 1, "Number of GPUs the job requires.")
	command.Flags().StringVarP(&(sa.Project), "project", "p", "default", "Specifies the project to use for this job, leave empty to use default project")
	command.Flags().StringVarP(&(sa.Image), "image", "i", "", "Specifies job image")
	command.Flags().BoolVar(&(sa.HostIPC), "host-ipc", false, "Use the host's ipc namespace. Optional: Default to false.")
	command.Flags().BoolVar(&(sa.Interactive), "interactive", false, "Create an interactive job")
	command.Flags().StringArrayVarP(&(sa.Volumes), "volumes", "v", []string{}, "Volumes to mount into the container")
	command.Flags().StringVar(&(sa.NodeType), "node-type", "", "Define node type for the job")
	command.Flags().StringVarP(&(sa.User), "user", "u", defaultUser, "Use different user to run the job")
	command.Flags().StringArrayVar(&(sa.Ports), "port", []string{}, "Add port mapping to job")
	command.Flags().StringVarP(&(sa.ServiceType), "service-type", "s", "", "Service type for the interactive job. Options are: portforward, loadbalancer, nodeport, ingress")
	command.Flags().StringArrayVar(&(sa.Command), "command", []string{}, "Command to run in the job contaner.")
	command.Flags().StringArrayVar(&(sa.Args), "args", []string{}, "Arguments to pass to the command")
	command.Flags().BoolVar(&(sa.IsJupiter), "jupiter", false, "Is this job a jupiter notebook server. Will use default configuration for jupiter notebook")
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
