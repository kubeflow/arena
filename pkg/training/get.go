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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/util"
)

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
			if strings.Contains(err.Error(), "forbidden: User") {
				return nil, fmt.Errorf("the user has no privileges to get the training job in namespace %v,reason: %v", namespace, err)
			}
			return nil, err
		}
		return job, nil
	}
	// 3.if job type is not given,search job by name
	jobs, err := getTrainingJobsByName(jobName, namespace)
	if err != nil {
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
		if string(trainer.Type()) != trainingType {
			log.Debugf("the job %s with type %s in namespace %s is not expected type %v", name, trainer.Type(), namespace, trainingType)
			continue
		}
		return trainer.GetTrainingJob(name, namespace)
	}
	return nil, types.ErrTrainingJobNotFound
}

func getTrainingJobsByName(name, namespace string) (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	trainers := GetAllTrainers()
	var wg sync.WaitGroup
	locker := new(sync.RWMutex)
	noPrivileges := false
	var errOwnerIsNotYou error
	for _, trainer := range trainers {
		wg.Add(1)
		t := trainer
		go func() {
			defer wg.Done()
			if !t.IsEnabled() {
				log.Debugf("the trainer %v is disabled,skip to use this trainer to get the training job", t.Type())
				return
			}
			job, err := t.GetTrainingJob(name, namespace)
			if err != nil {
				if strings.Contains(err.Error(), "forbidden: User") {
					log.Debugf("the user has no privileges to get the %v in namespace %v,reason: %v", t.Type(), namespace, err)
					noPrivileges = true
					return
				}
				if err != types.ErrNoPrivilegesToOperateJob {
					log.Debugf("the job %s in namespace %s is not supported by %v", name, namespace, t.Type())
					return
				}
				errOwnerIsNotYou = err
			}
			locker.Lock()
			jobs = append(jobs, job)
			locker.Unlock()
		}()
	}
	wg.Wait()
	if noPrivileges {
		return nil, fmt.Errorf("the user has no privileges to get the training job in namespace %v", namespace)
	}
	if len(jobs) == 0 {
		log.Debugf("Failed to find the training job %s in namespace %s", name, namespace)
		return nil, types.ErrTrainingJobNotFound
	}
	if errOwnerIsNotYou != nil {
		return nil, errOwnerIsNotYou
	}
	return jobs, nil
}

func PrintTrainingJob(job TrainingJob, modelVersion *types.ModelVersion, format string, showEvents bool, showGPUs bool) {
	services, nodes := PrepareServicesAndNodesForTensorboard([]TrainingJob{job}, false)
	switch format {
	case "name":
		fmt.Println(job.Name())
		// for future CRD support
	case "json":
		jobInfo := BuildJobInfo(job, showGPUs, services, nodes)
		patchModelInfo(jobInfo, modelVersion)
		outBytes, err := json.MarshalIndent(jobInfo, "", "    ")
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Print(string(outBytes))
		}
	case "yaml":
		jobInfo := BuildJobInfo(job, showGPUs, services, nodes)
		patchModelInfo(jobInfo, modelVersion)
		outBytes, err := yaml.Marshal(jobInfo)
		if err != nil {
			fmt.Printf("Failed due to %v", err)
		} else {
			fmt.Print(string(outBytes))
		}
	case "wide", "":
		jobInfo := BuildJobInfo(job, showGPUs, services, nodes)
		patchModelInfo(jobInfo, modelVersion)
		printSingleJobHelper(jobInfo, job.Resources(), showEvents, showGPUs)
		job.Resources()
	default:
		log.Fatalf("Unknown output format: %s", format)
	}
}

func printSingleJobHelper(job *types.TrainingJobInfo, resource []Resource, showEvents bool, showGPU bool) {
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
		lines = printEvents(lines, chiefPodNamespace, resource)
	}
	var duration int64
	var err error
	var endTime string
	job.Duration = strings.Replace(job.Duration, "s", "", -1)
	duration, err = strconv.ParseInt(job.Duration, 10, 64)
	if err != nil {
		log.Debugf("failed to parse duration: %v", err)

	}
	if job.Status == types.TrainingJobSucceeded || job.Status == types.TrainingJobFailed {
		endTime = util.GetFormatTime(job.CreationTimestamp + duration)
	}
	//startTime := time.Unix(job.CreationTimestamp, 0).Format("2006-01-02/15:04:05")
	fmt.Fprintf(w, "Name:\t%v\n", job.Name)
	fmt.Fprintf(w, "Status:\t%v\n", job.Status)
	fmt.Fprintf(w, "Namespace:\t%v\n", job.Namespace)
	fmt.Fprintf(w, "Priority:\t%v\n", job.Priority)
	fmt.Fprintf(w, "Trainer:\t%v\n", strings.ToUpper(string(job.Trainer)))
	fmt.Fprintf(w, "Duration:\t%v\n", util.ShortHumanDuration(time.Duration(duration)*time.Second))
	fmt.Fprintf(w, "CreateTime:\t%v\n", util.GetFormatTime(job.CreationTimestamp))
	fmt.Fprintf(w, "EndTime:\t%v\n", endTime)
	if job.ModelName != "" {
		fmt.Fprintf(w, "ModelName:\t%v\n", job.ModelName)
	}
	if job.ModelVersion != "" {
		fmt.Fprintf(w, "ModelVersion:\t%v\n", job.ModelVersion)
	}
	if job.ModelSource != "" {
		fmt.Fprintf(w, "ModelSource:\t%v\n", job.ModelSource)
	}
	fmt.Fprintf(w, "%v\n", strings.Join(lines, "\n"))
	w.Flush()
}

func printEvents(lines []string, namespace string, resources []Resource) []string {
	lines = append(lines, "", "Events:")
	clientset := config.GetArenaConfiger().GetClientSet()
	eventsMap, err := GetResourcesEvents(clientset, namespace, resources)
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
				util.ShortHumanDuration(time.Since(event.CreationTimestamp.Time)),
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
	events, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
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

func patchModelInfo(jobInfo *types.TrainingJobInfo, modelVersion *types.ModelVersion) *types.TrainingJobInfo {
	if modelVersion == nil {
		return jobInfo
	}
	jobInfo.ModelName = modelVersion.Name
	jobInfo.ModelVersion = modelVersion.Version
	jobInfo.ModelSource = modelVersion.Source
	return jobInfo
}
