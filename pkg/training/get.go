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

package training

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"

	"github.com/kubeflow/arena/pkg/util"
	yaml "gopkg.in/yaml.v2"

	"io"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	errGetMsg = "Failed to get the training job %s, but the trainer config is found, please clean it by using 'arena delete %s %v'."
)

/*
* search the training job with name and training type
 */
func SearchTrainingJob(jobName, namespace string, jobType types.TrainingJobType) (TrainingJob, error) {
	// 1.if job type is unknown,return error
	if jobType == types.UnknownTrainingJob {
		return nil, fmt.Errorf("Unsupport job type,arena only supports: [%v]", strings.Join(getTrainingJobTypes(), ","))
	}
	// 2.if job type is given,search the job
	if jobType != types.AllTrainingJob {
		job, err := getTrainingJobByType(jobName, namespace, string(jobType))
		if err != nil {
			if isTrainingConfigExist(jobName, string(jobType), namespace) {
				log.Warningf(errGetMsg, jobName, jobName, "--type "+string(jobType))
			}
			return nil, err
		}
		return job, nil
	}
	// 3.if job type is not given,search job by name
	jobs, err := getTrainingJobsByName(jobName, namespace)
	if err != nil {
		if len(getTrainingTypes(jobName, namespace)) > 0 {
			log.Warningf(errGetMsg, jobName, jobName, "")
		}
		return nil, err
	}
	if len(jobs) == 1 {
		return jobs[0], nil
	}
	return nil, fmt.Errorf("There are more than 1 training jobs with the same name %s, please check it with `arena list | grep %s`",
		jobName,
		jobName,
	)
}

func getTrainingJobByType(name, namespace, trainingType string) (job TrainingJob, err error) {
	trainers := NewSupportedTrainers()
	for _, trainer := range trainers {
		if trainer.Type() == trainingType {
			return trainer.GetTrainingJob(name, namespace)
		}
		log.Debugf("the job %s with type %s in namespace %s is not expected type %v",
			name,
			trainer.Type(),
			namespace,
			trainingType,
		)
	}
	return nil, fmt.Errorf("Failed to find the training job %s in namespace %s", name, namespace)
}

func getTrainingJobsByName(name, namespace string) (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	trainers := NewSupportedTrainers()
	for _, trainer := range trainers {
		if !trainer.IsSupported(name, namespace) {
			log.Debugf("the job %s in namespace %s is not supported by %v", name, namespace, trainer.Type())
			continue
		}
		job, err := trainer.GetTrainingJob(name, namespace)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if len(jobs) == 0 {
		log.Debugf("Failed to find the training job %s in namespace %s", name, namespace)
		return nil, fmt.Errorf("The job '%s' in namespace %s doesn't exist, please use 'arena submit' to create it first.\n", name, namespace)
	}
	return jobs, nil
}

func PrintTrainingJob(job TrainingJob, format string, showEvents bool) {
	switch format {
	case "name":
		fmt.Println(job.Name())
		// for future CRD support
	case "json":
		outBytes, err := json.MarshalIndent(BuildJobInfo(job), "", "    ")
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Printf(string(outBytes))
		}
	case "yaml":
		outBytes, err := yaml.Marshal(BuildJobInfo(job))
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Printf(string(outBytes))
		}
	case "wide", "":
		printSingleJobHelper(BuildJobInfo(job), job.Resources(), showEvents)
		job.Resources()
	default:
		log.Fatalf("Unknown output format: %s", format)
	}
}

func printSingleJobHelper(job *types.TrainingJobInfo, resouce []Resource, showEvents bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printJobSummary(w, job)

	// apply a dummy FgDefault format to align tabwriter with the rest of the columns

	fmt.Fprintf(w, "NAME\tSTATUS\tTRAINER\tAGE\tINSTANCE\tNODE\n")

	for _, instance := range job.Instances {
		// hostIP := "N/A"

		// if pod.Status.Phase == v1.PodRunning {
		hostIP := instance.Node
		// }

		if len(hostIP) == 0 {
			hostIP = "N/A"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", job.Name,
			instance.Status,
			strings.ToUpper(string(job.Trainer)),
			instance.Age,
			instance.Name,
			hostIP)
	}

	if job.Tensorboard != "" {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Your tensorboard will be available on:")
		fmt.Fprintf(w, "%s \t\n", job.Tensorboard)
	}
	chiefPodNamespace := ""
	if job.ChiefName != "" {
		chiefPodNamespace = job.Namespace
	}
	if showEvents {
		printEvents(w, chiefPodNamespace, resouce)
	}

	_ = w.Flush()

}

func printJobSummary(w io.Writer, job *types.TrainingJobInfo) {
	log.Debugf("--->STATUS: %s", job.Status)
	log.Debugf("--->NAMESPACE: %s", job.Namespace)
	log.Debugf("--->PRIORITY: %s", job.Priority)
	log.Debugf("--->TRAINING DURATION: %s", job.Duration)

	fmt.Fprintf(w, "STATUS: %s\n", job.Status)
	fmt.Fprintf(w, "NAMESPACE: %s\n", job.Namespace)
	fmt.Fprintf(w, "PRIORITY: %s\n", job.Priority)
	fmt.Fprintf(w, "TRAINING DURATION: %s\n", job.Duration)
	fmt.Fprintln(w, "")

}

func printEvents(w io.Writer, namespace string, resouces []Resource) {
	fmt.Fprintf(w, "\nEvents: \n")
	clientset := config.GetArenaConfiger().GetClientSet()
	eventsMap, err := GetResourcesEvents(clientset, namespace, resouces)
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
