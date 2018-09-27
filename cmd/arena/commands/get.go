// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	"io"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var output string

var dashboardURL string

// NewGetCommand
func NewGetCommand() *cobra.Command {
	printArgs := PrintArgs{}
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
			printTrainingJob(job, printArgs)
		},
	}

	command.Flags().BoolVarP(&printArgs.ShowEvents, "events", "e", false, "Specify if show pending pod's events.")
	command.Flags().StringVarP(&printArgs.Output, "output", "o", "", "Output format. One of: json|yaml|wide")
	return command
}

type PrintArgs struct {
	ShowEvents bool
	Output string
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

func printTrainingJob(job TrainingJob, printArgs PrintArgs) {
	switch printArgs.Output {
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
		printSingleJobHelper(job, printArgs)
	default:
		log.Fatalf("Unknown output format: %s", printArgs.Output)
	}
}

func printSingleJobHelper(job TrainingJob, printArgs PrintArgs) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// apply a dummy FgDefault format to align tabwriter with the rest of the columns

	fmt.Fprintf(w, "NAME\tSTATUS\tTRAINER\tAGE\tINSTANCE\tNODE\n")
	pods := job.AllPods()

	for _, pod := range pods {
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

	if printArgs.ShowEvents {
		printEvents(w, job.ChiefPod().Namespace, pods)
	}

	_ = w.Flush()

}

func printEvents(w io.Writer, namespace string, pods []v1.Pod) {
	eventsMap, _ := GetPodEvents(clientset, namespace, pods)
	pendingPods := []v1.Pod{}
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		pendingPods = append(pendingPods, pod)
	}
	fmt.Fprintf(w, "\nEvents: \n")
	if len(pendingPods) == 0 {
		fmt.Fprintln(w, "No events for pending pod")
		return
	}
	fmt.Fprintf(w, "INSTANCE\tTYPE\tAGE\tMESSAGE\n")
	fmt.Fprintf(w, "--------\t----\t---\t-------\n")

	for _, pod := range pendingPods {
		events := eventsMap[pod.Name]
		for _, event := range events {
			instanceName := pod.Name
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
				instanceName,
				event.Type,
				util.ShortHumanDuration(time.Now().Sub(event.CreationTimestamp.Time)),
				fmt.Sprintf("[%s] %s", event.Reason, event.Message))
		}
		// empty line for per pod
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "", "", "", "", "", "")
	}
}

// Get real job status
// WHen has pods being pending, tfJob still show in Running state, it should be Pending
func GetJobRealStatus(job TrainingJob) string {
	hasPendingPod := false
	jobStatus := job.GetStatus()
	if jobStatus == "RUNNING" {
		pods := job.AllPods()
		for _, pod := range pods {
			if pod.Status.Phase == v1.PodPending {
				hasPendingPod = true
				break
			}
		}
		if hasPendingPod {
			jobStatus = "PENDING"
		}
	}
	return jobStatus
}

// Get Event of the Job
func GetPodEvents(client *kubernetes.Clientset, namespace string, pods []v1.Pod) (map[string][]v1.Event, error) {
	eventMap := make(map[string][]v1.Event)
	events, err := client.CoreV1().Events(namespace).List(metav1.ListOptions{})
	if err != nil {
		return eventMap, err
	}
	for _, pod := range pods {
		eventMap[pod.Name] = []v1.Event{}
		for _, event := range events.Items {
			if event.InvolvedObject.Kind == "Pod" && event.InvolvedObject.Name == pod.Name {
				eventMap[pod.Name] = append(eventMap[pod.Name], event)
			}
		}
	}
	return eventMap, nil
}
