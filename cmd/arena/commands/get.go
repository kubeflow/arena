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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

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

			job, err := searchTrainingJob(name, trainingType, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			printTrainingJob(job, printArgs)
		},
	}

	command.Flags().StringVar(&trainingType, "type", "", "The training type to get, the possible option is tfjob, mpijob, sparkjob,volcanojob,horovodjob or standalonejob. (optional)")
	command.Flags().BoolVarP(&printArgs.ShowEvents, "events", "e", false, "Specify if show pending pod's events.")
	command.Flags().StringVarP(&printArgs.Output, "output", "o", "", "Output format. One of: json|yaml|wide")
	return command
}

type PrintArgs struct {
	ShowEvents bool
	Output     string
}

/*
* search the training job with name and training type
 */
func searchTrainingJob(jobName, trainingType, namespace string) (job TrainingJob, err error) {
	if len(trainingType) > 0 {
		if isKnownTrainingType(trainingType) {
			job, err = getTrainingJobByType(clientset, jobName, namespace, trainingType)
			if err != nil {
				if isTrainingConfigExist(jobName, trainingType, namespace) {
					log.Warningf("Failed to get the training job %s, but the trainer config is found, please clean it by using 'arena delete %s --type %s'.",
						jobName,
						jobName,
						trainingType)
				}
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("%s is unknown training type, please choose a known type from %v",
				trainingType,
				knownTrainingTypes)
		}
	} else {
		jobs, err := getTrainingJobsByName(clientset, jobName, namespace)
		if err != nil {
			if len(getTrainingTypes(jobName, namespace)) > 0 {
				log.Warningf("Failed to get the training job %s, but the trainer config is found, please clean it by using 'arena delete %s'.",
					jobName,
					jobName)
			}
			return nil, err
		}

		if len(jobs) > 1 {
			return nil, fmt.Errorf("There are more than 1 training jobs with the same name %s, please check it with `arena list | grep %s`",
				jobName,
				jobName)
		} else {
			job = jobs[0]
		}
	}

	return job, nil
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

func getTrainingJobByType(client *kubernetes.Clientset, name, namespace, trainingType string) (job TrainingJob, err error) {
	// trainers := NewTrainers(client, )

	trainers := NewTrainers(client)
	for _, trainer := range trainers {
		if trainer.Type() == trainingType {
			return trainer.GetTrainingJob(name, namespace)
		} else {
			log.Debugf("the job %s with type %s in namespace %s is not expected type %v",
				name,
				trainer.Type(),
				namespace,
				trainingType)
		}
	}

	return nil, fmt.Errorf("Failed to find the training job %s in namespace %s", name, namespace)
}

func getTrainingJobsByName(client *kubernetes.Clientset, name, namespace string) (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	trainers := NewTrainers(client)
	for _, trainer := range trainers {
		if trainer.IsSupported(name, namespace) {
			job, err := trainer.GetTrainingJob(name, namespace)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, job)
		} else {
			log.Debugf("the job %s in namespace %s is not supported by %v", name, namespace, trainer.Type())
		}
	}

	if len(jobs) == 0 {
		log.Debugf("Failed to find the training job %s in namespace %s", name, namespace)
		return nil, fmt.Errorf("The job %s in namespace %s doesn't exist, please create it first. use 'arena submit'\n", name, namespace)
	}

	return jobs, nil
}

func printTrainingJob(job TrainingJob, printArgs PrintArgs) {
	switch printArgs.Output {
	case "name":
		fmt.Println(job.Name())
		// for future CRD support
	case "json":
		outBytes, err := json.MarshalIndent(BuildJobInfo(job), "", "    ")
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Println(string(outBytes))
		}
	case "yaml":
		outBytes, err := yaml.Marshal(BuildJobInfo(job))
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Println(string(outBytes))
		}
	case "wide", "":
		printSingleJobHelper(job, printArgs)
	default:
		log.Fatalf("Unknown output format: %s", printArgs.Output)
	}
}

func printSingleJobHelper(job TrainingJob, printArgs PrintArgs) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printJobSummary(w, job)

	// apply a dummy FgDefault format to align tabwriter with the rest of the columns

	fmt.Fprintf(w, "NAME\tSTATUS\tTRAINER\tAGE\tINSTANCE\tNODE\n")
	pods := job.AllPods()

	for _, pod := range pods {
		// hostIP := "N/A"

		// if pod.Status.Phase == v1.PodRunning {
		hostIP := pod.Status.HostIP
		// }

		if len(hostIP) == 0 {
			hostIP = "N/A"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", job.Name(),
			strings.ToUpper(string(pod.Status.Phase)),
			strings.ToUpper(job.Trainer()),
			util.ShortHumanDuration(job.Age()),
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
		printEvents(w, job.ChiefPod().Namespace, job)
	}

	_ = w.Flush()

}

func printJobSummary(w io.Writer, job TrainingJob) {
	fmt.Fprintf(w, "STATUS: %s\n", GetJobRealStatus(job))
	fmt.Fprintf(w, "NAMESPACE: %s\n", job.Namespace())
	fmt.Fprintf(w, "PRIORITY: %s\n", getPriorityClass(job))
	fmt.Fprintf(w, "TRAINING DURATION: %s\n", util.ShortHumanDuration(job.Duration()))
	fmt.Fprintln(w, "")

}

func printEvents(w io.Writer, namespace string, job TrainingJob) {
	fmt.Fprintf(w, "\nEvents: \n")
	eventsMap, err := GetResourcesEvents(clientset, namespace, job.Resources())
	if err != nil {
		fmt.Fprintf(w, "Get job events failed, due to: %v", err)
		return
	}
	if len(eventsMap) == 0 {
		fmt.Fprintln(w, "No events for resources")
		return
	}
	fmt.Fprintf(w, "SOURCE\tTYPE\tAGE\tMESSAGE\n")
	fmt.Fprintf(w, "--------\t----\t---\t-------\n")

	for resourceName, events := range eventsMap {
		for _, event := range events {
			instanceName := fmt.Sprintf("%s/%s", strings.ToLower(event.InvolvedObject.Kind), resourceName)
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
				log.Debugf("pod %s is pending", pod.Name)
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
func GetResourcesEvents(client *kubernetes.Clientset, namespace string, resources []Resource) (map[string][]v1.Event, error) {
	eventMap := make(map[string][]v1.Event)
	events, err := client.CoreV1().Events(namespace).List(metav1.ListOptions{})
	if err != nil {
		return eventMap, err
	}
	for _, resource := range resources {
		eventMap[resource.Name] = []v1.Event{}
		for _, event := range events.Items {
			if event.InvolvedObject.Kind == string(resource.ResourceType) && string(event.InvolvedObject.UID) == resource.Uid {
				eventMap[resource.Name] = append(eventMap[resource.Name], event)
			}
		}
	}
	return eventMap, nil
}
