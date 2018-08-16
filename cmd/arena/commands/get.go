package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/util"
	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/util/helm"
	"github.com/spf13/cobra"
	// podv1 "k8s.io/api/core/v1"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var output string

var dashboardURL string

// NewGetCommand
func NewGetCommand() *cobra.Command {
	var (
		output string
	)

	var command = &cobra.Command{
		Use:   "get training job",
		Short: "display details of a training job",
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name = args[0]

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			exist, err := helm.CheckRelease(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !exist {
				fmt.Printf("The job %s doesn't exist, please create it first. use 'arena submit'\n", name)
				os.Exit(1)
			}
			job, err := getTrainingJob(client, name, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			printTrainingJob(job, output)
		},
	}

	command.Flags().StringVarP(&output, "output", "o", "", "Output format. One of: json|yaml|wide")
	return command
}

func getTrainingJob(client *kubernetes.Clientset, name, namespace string) (job TrainingJob, err error) {
	// trainers := NewTrainers(client, )

	trainers := NewTrainers(client)
	for _, trainer := range trainers {
		if trainer.IsSupported(name, namespace) {
			return trainer.GetTrainingJob(name, namespace)
		} else {
			log.Debugf("the job %s in namespace %s is not supported by %v", name, namespace, trainer.Type())
		}
	}

	return nil, fmt.Errorf("Failed to find the training job %s in namespace %s", name, namespace)
}

func printTrainingJob(job TrainingJob, outFmt string) {
	switch outFmt {
	case "name":
		fmt.Println(job.Name())
	// for future CRD support
	// case "json":
	// 	outBytes, _ := json.MarshalIndent(job, "", "    ")
	// 	fmt.Println(string(outBytes))
	// case "yaml":
	// 	outBytes, _ := yaml.Marshal(job.)
	// 	fmt.Print(string(outBytes))
	case "wide", "":
		printSingleJobHelper(job, outFmt)
	default:
		log.Fatalf("Unknown output format: %s", outFmt)
	}
}

func printSingleJobHelper(job TrainingJob, outFmt string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// apply a dummy FgDefault format to align tabwriter with the rest of the columns

	fmt.Fprintf(w, "NAME\tSTATUS\tTRAINER\tAGE\tINSTANCE\tNODE\n")
	for _, pod := range job.AllPods() {
		hostIP := "N/A"

		if pod.Status.Phase == v1.PodRunning {
			hostIP = pod.Status.HostIP
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", job.Name(),
			strings.ToUpper(string(pod.Status.Phase)),
			strings.ToUpper(job.Trainer()),
			job.Age(),
			pod.Name,
			hostIP)
	}

	url, err := tensorboardURL(job.Name(), job.ChiefPod().Namespace)
	if url == "" || err != nil {
		log.Debugf("Tensorboard dones't show up because of %v, or url %s", err, url)
	} else {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Your tensorboard will be available on:")
		fmt.Fprintf(w, "%s \t\n", url)
	}

	_ = w.Flush()

}
