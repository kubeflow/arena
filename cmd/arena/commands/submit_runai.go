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

			if submitArgs.IsJupyter {
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
	IsJupyter   bool
}

func (sa *submitRunaiJobArgs) UseJupiterDefaultValues() {
	var (
		jupyterPort    = "8888"
		jupyterImage   = "jupyter/scipy-notebook"
		jupyterCommand = "start-notebook.sh"
		jupyterArgs    = "--NotebookApp.base_url=/%s"
	)

	if len(sa.Ports) == 0 {
		sa.Ports = []string{jupyterPort}
		log.Infof("Exposing default jupyter notebook port %s", jupyterPort)
	}
	if sa.Image == "" {
		sa.Image = "jupyter/scipy-notebook"
		log.Infof("Using default jupyter notebook image \"%s\"", jupyterImage)
	}
	if len(sa.Command) == 0 && sa.ServiceType == "ingress" {
		sa.Command = []string{jupyterCommand}
		log.Infof("Using default jupyter notebook command for using ingress service \"%s\"", jupyterCommand)
	}
	if len(sa.Args) == 0 && sa.ServiceType == "ingress" {
		baseUrlArg := fmt.Sprintf(jupyterArgs, name)
		sa.Args = []string{baseUrlArg}
		log.Infof("Using default jupyter notebook command argument for using ingress service \"%s\"", baseUrlArg)
	}
}

// add flags to submit spark args
func (sa *submitRunaiJobArgs) addFlags(command *cobra.Command) {
	currentUser, _ := user.Current()
	defaultUser := currentUser.Username

	command.Flags().StringVar(&name, "name", "", "Job name")
	command.MarkFlagRequired("name")

	command.Flags().IntVarP(&(sa.GPU), "gpu", "g", 1, "Number of GPUs to allocation to the Job.")
	command.Flags().StringVarP(&(sa.Project), "project", "p", "default", "Specifies the Run:AI project to use for this Job.")
	command.Flags().StringVarP(&(sa.Image), "image", "i", "", "Image to use when creating the container for this Job.")
	command.Flags().BoolVar(&(sa.HostIPC), "host-ipc", false, "Use the host's ipc namespace. (default 'false').")
	command.Flags().BoolVar(&(sa.Interactive), "interactive", false, "Mark this Job as unattended or interactive. (default 'false')")
	command.Flags().StringArrayVarP(&(sa.Volumes), "volumes", "v", []string{}, "Volumes to mount into the container.")
	command.Flags().StringVar(&(sa.NodeType), "node-type", "", "Enforce node type affinity by setting a node-type label.")
	command.Flags().StringVarP(&(sa.User), "user", "u", defaultUser, "Use different user to run the Job.")
	command.Flags().StringArrayVar(&(sa.Ports), "port", []string{}, "Expose ports from the Job container.")
	command.Flags().StringVarP(&(sa.ServiceType), "service-type", "s", "", "Service exposure method for interactive Job. Options are: portforward, loadbalancer, nodeport, ingress.")
	command.Flags().StringArrayVar(&(sa.Command), "command", []string{}, "Run this command on container start. Use together with --args.")
	command.Flags().StringArrayVar(&(sa.Args), "args", []string{}, "Arguments to pass to the command run on container start. Use together with --command.")
	command.Flags().BoolVar(&(sa.IsJupyter), "jupyter", false, "Shortcut for running a jupyter notebook container. Uses a pre-created image and a default notebookd configuration.")
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
