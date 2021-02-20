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
	"strconv"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"

	"github.com/kubeflow/arena/pkg/util"
	yaml "gopkg.in/yaml.v2"

	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	errJobNotFoundMessage = "Not found training job %s in namespace %s,please use 'arena submit' to create it."
	errGetMsg             = "Failed to get the training job %s, but the trainer config is found, please clean it by using 'arena delete %s %v'."
)
var getJobTemplate = `
Name:      %v
Status:    %v
Namespace: %v
Priority:  %v
Trainer:   %v
Duration:  %v
%v
`

/*
* search the training job with name and training type
 */
func SearchTrainingJob(jobName, namespace string, jobType types.TrainingJobType) (TrainingJob, error) {
	// 1.if job type is unknown,return error
	if jobType == types.UnknownTrainingJob {
		return nil, fmt.Errorf("Unsupport job type,arena only supports: [%v]", utils.GetSupportTrainingJobTypesInfo())
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
	trainers := GetAllTrainers()
	for _, trainer := range trainers {
		if !trainer.IsEnabled() {
			log.Debugf("the trainer %v is disabled,skip to use this trainer to get the training job", trainer.Type())
			continue
		}
		if string(trainer.Type()) == trainingType {
			return trainer.GetTrainingJob(name, namespace)
		}
		log.Debugf("the job %s with type %s in namespace %s is not expected type %v",
			name,
			trainer.Type(),
			namespace,
			trainingType,
		)
	}
	return nil, types.ErrTrainingJobNotFound
}

func getTrainingJobsByName(name, namespace string) (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	trainers := GetAllTrainers()
	for _, trainer := range trainers {
		if !trainer.IsEnabled() {
			log.Debugf("the trainer %v is disabled,skip to use this trainer to get the training job", trainer.Type())
			continue
		}
		if !trainer.IsSupported(name, namespace) {
			log.Debugf("the job %s in namespace %s is not supported by %v", name, namespace, trainer.Type())
			continue
		}
		job, err := trainer.GetTrainingJob(name, namespace)
		if err != nil {
			if err == types.ErrTrainingJobNotFound {
				continue
			}
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if len(jobs) == 0 {
		log.Debugf("Failed to find the training job %s in namespace %s", name, namespace)
		return nil, types.ErrTrainingJobNotFound
	}
	return jobs, nil
}

func PrintTrainingJob(job TrainingJob, format string, showEvents bool, showGPUs bool) {
	switch format {
	case "name":
		fmt.Println(job.Name())
		// for future CRD support
	case "json":
		outBytes, err := json.MarshalIndent(BuildJobInfo(job, showGPUs), "", "    ")
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Printf(string(outBytes))
		}
	case "yaml":
		outBytes, err := yaml.Marshal(BuildJobInfo(job, showGPUs))
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Printf(string(outBytes))
		}
	case "wide", "":
		printSingleJobHelper(BuildJobInfo(job, showGPUs), job.Resources(), showEvents, showGPUs)
		job.Resources()
	default:
		log.Fatalf("Unknown output format: %s", format)
	}
}

func printSingleJobHelper(job *types.TrainingJobInfo, resouce []Resource, showEvents bool, showGPU bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	lines := []string{"", "Instances:", "  NAME\tSTATUS\tAGE\tIS_CHIEF\tGPU(Requested)\tNODE"}
	lines = append(lines, "  ----\t------\t---\t--------\t--------------\t----")
	totalRequestGPUs := 0
	totalAllocatedGPUs := 0
	for _, instance := range job.Instances {
		totalRequestGPUs += instance.RequestGPUs
		if instance.Status == "Running" {
			totalAllocatedGPUs += instance.RequestGPUs
		}
		hostIP := instance.Node
		if len(hostIP) == 0 {
			hostIP = "N/A"
		}
		var duration int64
		var err error
		job.Duration = strings.Replace(job.Duration, "s", "", -1)
		duration, err = strconv.ParseInt(job.Duration, 10, 64)
		if err != nil {
			log.Debugf("failed to parse duration: %v", err)

		}
		lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v\t%v\t%v",
			instance.Name,
			instance.Status,
			util.ShortHumanDuration(time.Duration(duration)*time.Second),
			instance.IsChief,
			instance.RequestGPUs,
			hostIP,
		))
	}
	if job.Status != types.TrainingJobSucceeded {
		lines = displayGPUUsage(lines, job.Status, totalAllocatedGPUs, totalRequestGPUs, job.Instances, showGPU)
	}
	if job.Tensorboard != "" {
		lines = append(lines, "", "Tensorboard:")
		lines = append(lines, "  Your tensorboard will be available on: ")
		lines = append(lines, fmt.Sprintf("  %v", job.Tensorboard))
	}
	chiefPodNamespace := ""
	if job.ChiefName != "" {
		chiefPodNamespace = job.Namespace
	}
	if showEvents {
		lines = printEvents(lines, chiefPodNamespace, resouce)
	}
	var duration int64
	var err error
	job.Duration = strings.Replace(job.Duration, "s", "", -1)
	duration, err = strconv.ParseInt(job.Duration, 10, 64)
	if err != nil {
		log.Debugf("failed to parse duration: %v", err)

	}
	//startTime := time.Unix(job.CreationTimestamp, 0).Format("2006-01-02/15:04:05")
	PrintLine(w, fmt.Sprintf(strings.Trim(getJobTemplate, "\n"),
		job.Name,
		job.Status,
		job.Namespace,
		job.Priority,
		strings.ToUpper(string(job.Trainer)),
		util.ShortHumanDuration(time.Duration(duration)*time.Second),
		strings.Join(lines, "\n"),
	))
	PrintLine(w, "")
	_ = w.Flush()

}

func printEvents(lines []string, namespace string, resouces []Resource) []string {
	lines = append(lines, "", "Events:")
	clientset := config.GetArenaConfiger().GetClientSet()
	eventsMap, err := GetResourcesEvents(clientset, namespace, resouces)
	if err != nil {
		lines = append(lines, fmt.Sprintf("  Get job events failed, due to: %v", err))
		return lines
	}
	if len(eventsMap) == 0 {
		lines = append(lines, "  No events for resources")
		return lines
	}
	lines = append(lines, "  SOURCE\tTYPE\tAGE\tMESSAGE")
	lines = append(lines, "  ------\t----\t---\t-------")
	for resourceName, events := range eventsMap {
		for _, event := range events {
			instanceName := fmt.Sprintf("%s/%s", strings.ToLower(event.InvolvedObject.Kind), resourceName)
			lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v",
				instanceName,
				event.Type,
				util.ShortHumanDuration(time.Now().Sub(event.CreationTimestamp.Time)),
				fmt.Sprintf("[%s] %s", event.Reason, event.Message),
			))
		}
		lines = append(lines, "")
	}
	return lines
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

func displayGPUUsage(lines []string, status types.TrainingJobStatus, totalAllocatedGPUs, totalRequestGPUs int, instances []types.TrainingJobInstance, showGPU bool) []string {
	if !showGPU || totalRequestGPUs == 0 {
		return lines
	}
	lines = append(lines, "", "GPUs:")
	lines = append(lines, "  INSTANCE\tNODE(IP)\tGPU(Requested)\tGPU(IndexId)\tGPU(DutyCycle)\tGPU Memory(Used/Total)")
	lines = append(lines, "  --------\t--------\t--------------\t------------\t--------------\t----------------------")
	for _, instance := range instances {
		if len(instance.GPUMetrics) == 0 {
			lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v\t%v\t%v",
				instance.Name,
				instance.NodeIP,
				instance.RequestGPUs,
				"N/A", "N/A", "N/A",
			))
			continue
		}
		for index, gpuID := range SortMapKeys(instance.GPUMetrics) {
			name := instance.Name
			requestGPUs := fmt.Sprintf("%v", instance.RequestGPUs)
			hostIP := instance.NodeIP
			if index != 0 {
				name = ""
				requestGPUs = ""
				hostIP = ""
			}
			gpuMetric := instance.GPUMetrics[gpuID]
			lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v\t%.1f%%\t%.1f/%.1f(MiB)",
				name,
				hostIP,
				requestGPUs,
				gpuID,
				gpuMetric.GpuDutyCycle,
				fromByteToMiB(gpuMetric.GpuMemoryUsed),
				fromByteToMiB(gpuMetric.GpuMemoryTotal),
			))
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Allocated/Requested GPUs of Job: %v/%v", totalAllocatedGPUs, totalRequestGPUs))
	return lines
}
