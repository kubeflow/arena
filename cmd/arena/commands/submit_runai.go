package commands

import (
	"fmt"
	"math"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/clusterConfig"
	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	runaiChart       = path.Join(util.GetChartsFolder(), "runai")
	ttlAfterFinished *time.Duration
	configArg        string
)

const (
	defaultRunaiTrainingType = "runai"
	runaiNamespace           = "runai"
)

func NewRunaiJobCommand() *cobra.Command {

	submitArgs := NewSubmitRunaiJobArgs()
	var command = &cobra.Command{
		Use:     "submit [NAME]",
		Short:   "Submit a Runai job.",
		Aliases: []string{"ra"},
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			kubeClient, err := client.GetClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			clientset := kubeClient.GetClientset()

			submitArgs.setCommonRun(cmd, args, kubeClient)

			index, err := getJobIndex(clientset)

			if err != nil {
				log.Debug("Could not get job index. Will not set a label.")
			} else {
				submitArgs.Labels = make(map[string]string)
				submitArgs.Labels["runai/job-index"] = index
			}

			if ttlAfterFinished != nil {
				ttlSeconds := int(math.Round(ttlAfterFinished.Seconds()))
				log.Debugf("Using time to live seconds %d", ttlSeconds)
				submitArgs.TTL = &ttlSeconds
			}

			if submitArgs.IsJupyter {
				submitArgs.UseJupyterDefaultValues()
			}

			if submitArgs.RunAsCurrentUser {
				currentUser, err := user.Current()
				if err != nil {
					submitArgs.RunAsUser = currentUser.Uid
					submitArgs.RunAsGroup = currentUser.Gid
				}
			}

			err = submitRunaiJob(args, submitArgs, clientset)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			printJobInfoIfNeeded(submitArgs)
			if submitArgs.IsJupyter || (submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.ServiceType == "portforward") {
				err = kubectl.WaitForReadyStatefulSet(submitArgs.Name, submitArgs.Namespace)

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if submitArgs.IsJupyter {
					runaiTrainer := NewRunaiTrainer(*kubeClient)
					job, err := runaiTrainer.GetTrainingJob(submitArgs.Name, submitArgs.Namespace)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					pod := job.ChiefPod()
					logs, err := kubectl.Logs(pod.Name, pod.Namespace)

					token, err := getTokenFromJupyterLogs(string(logs))

					if err != nil {
						fmt.Println(err)
						fmt.Printf("Please run '%s logs %s' to view the logs.\n", config.CLIName, submitArgs.Name)
					}

					fmt.Printf("Jupyter notebook token: %s\n", token)
				}

				if submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.ServiceType == "portforward" {
					localPorts := []string{}
					for _, port := range submitArgs.Ports {
						split := strings.Split(port, ":")
						localPorts = append(localPorts, split[0])
					}

					localUrls := []string{}
					for _, localPort := range localPorts {
						localUrls = append(localUrls, fmt.Sprintf("localhost:%s", localPort))
					}

					accessPoints := strings.Join(localUrls, ",")
					fmt.Printf("Open access point(s) to service from %s\n", accessPoints)
					err = kubectl.PortForward(localPorts, submitArgs.Name, submitArgs.Namespace)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
		},
	}

	submitArgs.addCommonFlags(command)
	submitArgs.addFlags(command)

	return command
}

func printJobInfoIfNeeded(submitArgs *submitRunaiJobArgs) {
	if submitArgs.Interactive != nil && *submitArgs.Interactive && submitArgs.IsPreemptible != nil && *submitArgs.IsPreemptible {
		log.Infof("Using the preemptible flag may lead to your resources being preempted without notice")
	}
}

func getJobIndex(clientset *kubernetes.Clientset) (string, error) {
	for true {
		index, shouldTryAgain, err := tryGetJobIndexOnce(clientset)

		if index != "" || !shouldTryAgain {
			return index, err
		}
	}

	return "", nil
}

func tryGetJobIndexOnce(clientset *kubernetes.Clientset) (string, bool, error) {
	var (
		indexKey      = "index"
		configMapName = "runai-cli-index"
	)

	configMap, err := clientset.CoreV1().ConfigMaps(runaiNamespace).Get(configMapName, metav1.GetOptions{})

	// If configmap does not exists try to create it
	if err != nil {
		data := make(map[string]string)
		data[indexKey] = "1"
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: configMapName,
			},
			Data: data,
		}

		_, err := clientset.CoreV1().ConfigMaps(runaiNamespace).Create(configMap)

		// Might be someone already created this configmap. Try the process again.
		if err != nil {
			return "", true, nil
		}

		return "1", false, nil
	}

	lastIndex, err := strconv.Atoi(configMap.Data[indexKey])

	if err != nil {
		return "", false, err
	}

	newIndex := fmt.Sprintf("%d", lastIndex+1)
	configMap.Data[indexKey] = newIndex

	_, err = clientset.CoreV1().ConfigMaps(runaiNamespace).Update(configMap)

	// Might be someone already updated this configmap. Try the process again.
	if err != nil {
		return "", true, err
	}

	return newIndex, false, nil
}

func getTokenFromJupyterLogs(logs string) (string, error) {
	re, err := regexp.Compile(`\?token=(.*)\n`)
	if err != nil {
		return "", err
	}

	res := re.FindStringSubmatch(logs)
	if len(res) < 2 {
		return "", fmt.Errorf("Could not find token string in logs")
	}
	return res[1], nil
}

func NewSubmitRunaiJobArgs() *submitRunaiJobArgs {
	return &submitRunaiJobArgs{}
}

type submitRunaiJobArgs struct {
	// These arguments should be omitted when empty, to support default values file created in the cluster
	// So any empty ones won't override the default values
	submitArgs       `yaml:",inline"`
	HostIPC          *bool             `yaml:"hostIPC,omitempty"`
	GPUInt           *int              `yaml:"gpuInt,omitempty"`
	GPUFraction      string            `yaml:"gpuFraction,omitempty"`
	GPUFractionFixed string            `yaml:"gpuFractionFixed,omitempty"`
	Interactive      *bool             `yaml:"interactive,omitempty"`
	Volumes          []string          `yaml:"volume,omitempty"`
	Command          []string          `yaml:"command"`
	Ports            []string          `yaml:"ports,omitempty"`
	ServiceType      string            `yaml:"serviceType,omitempty"`
	Elastic          *bool             `yaml:"elastic,omitempty"`
	LargeShm         *bool             `yaml:"shm,omitempty"`
	NumberProcesses  int               `yaml:"numProcesses"` // --workers
	LocalImage       *bool             `yaml:"localImage,omitempty"`
	HostNetwork      *bool             `yaml:"hostNetwork,omitempty"`
	TTL              *int              `yaml:"ttlSecondsAfterFinished,omitempty"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	IsJupyter        bool
	IsPreemptible    *bool  `yaml:"isPreemptible,omitempty"`
	WorkingDir       string `yaml:"workingDir,omitempty"`
	RunAsUser        string `yaml:"runAsUser,omitempty"`
	RunAsGroup       string `yaml:"runAsGroup,omitempty"`
	RunAsCurrentUser bool
}

func (sa *submitRunaiJobArgs) UseJupyterDefaultValues() {
	var (
		jupyterPort        = "8888"
		jupyterImage       = "jupyter/scipy-notebook"
		jupyterCommand     = "start-notebook.sh"
		jupyterArgs        = "--NotebookApp.base_url=/%s"
		jupyterServiceType = "portforward"
	)

	interactive := true
	sa.Interactive = &interactive
	if len(sa.Ports) == 0 {
		sa.Ports = []string{jupyterPort}
		log.Infof("Exposing default jupyter notebook port %s", jupyterPort)
	}
	if sa.Image == "" {
		sa.Image = "jupyter/scipy-notebook"
		log.Infof("Using default jupyter notebook image \"%s\"", jupyterImage)
	}
	if sa.ServiceType == "" {
		sa.ServiceType = jupyterServiceType
		log.Infof("Using default jupyter notebook service type %s", jupyterServiceType)
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

	flags.AddBoolNullableFlag(command.Flags(), &(sa.HostIPC), "host-ipc", "Use the host's ipc namespace.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.Interactive), "interactive", "Mark this Job as unattended or interactive.")
	command.Flags().StringArrayVar(&(sa.Ports), "port", []string{}, "Expose ports from the Job container.")
	command.Flags().StringVarP(&(sa.ServiceType), "service-type", "s", "", "Service exposure method for interactive Job. Options are: portforward, loadbalancer, nodeport, ingress.")
	command.Flags().StringVar(&(sa.WorkingDir), "working-dir", "", "Container's working directory.")
	command.Flags().BoolVar(&(sa.IsJupyter), "jupyter", false, "Shortcut for running a jupyter notebook container. Uses a pre-created image and a default notebook configuration.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.Elastic), "elastic", "Mark the job as elastic.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.IsPreemptible), "preemptible", "Mark the job as preemptible.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.LargeShm), "large-shm", "Mount a large /dev/shm device. Specific software might need this feature.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.LocalImage), "local-image", "Use a local image for this job. NOTE: this image must exists on the local server.")
	flags.AddBoolNullableFlag(command.Flags(), &(sa.HostNetwork), "host-network", "Use the host's network stack inside the container.")
	command.Flags().StringArrayVar(&(sa.Command), "command", []string{}, "Run this command on container start. Use together with --args.")
	command.Flags().BoolVar(&(sa.RunAsCurrentUser), "run-as-user", false, "Run in the context of the current user running the Run:AI command rather than the root user.")

	flags.AddDurationNullableFlagP(command.Flags(), &(ttlAfterFinished), "ttl-after-finish", "", "Define the duration, post job finish, after which the job is automatically deleted (5s, 2m, 3h, .etc).")

	command.Flags().StringVarP(&(configArg), "template", "t", "", "Use a specific template to run this job. (otherwise use the default one if exists)")

	command.Flags().StringArrayVarP(&(sa.Volumes), "volume", "v", []string{}, "Volumes to mount into the container.")
	command.Flags().StringArrayVar(&(sa.Volumes), "volumes", []string{}, "Volumes to mount into the container.")
	command.Flags().MarkDeprecated("volumes", "please use 'volume' flag instead.")
}

func submitRunaiJob(args []string, submitArgs *submitRunaiJobArgs, clientset *kubernetes.Clientset) error {
	configs := clusterConfig.NewClusterConfigs(clientset)

	var configToUse *clusterConfig.ClusterConfig
	var err error
	if configArg == "" {
		configToUse, err = configs.GetClusterDefaultConfig()
	} else {
		configToUse, err = configs.GetClusterConfig(configArg)
		if configToUse == nil {
			return fmt.Errorf("Could not find runai template %s. Please run '%s template list'", configArg, config.CLIName)
		}
	}

	if err != nil {
		return err
	}

	configValues := ""
	if configToUse != nil {
		configValues = configToUse.Values
	}

	err = handleRequestedGPUs(submitArgs)
	if err != nil {
		return err
	}

	err = workflow.SubmitJob(submitArgs.Name, defaultRunaiTrainingType, submitArgs.Namespace, submitArgs, configValues, runaiChart, clientset, dryRun)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", submitArgs.Name)
	log.Infof("You can run `%s get %s` to check the job status", config.CLIName, submitArgs.Name)
	return nil
}
