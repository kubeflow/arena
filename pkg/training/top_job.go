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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v2"

	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
)

var topJobTemplate = `
Name:      	%v
Status:    	%v
Namespace: 	%v
Priority:  	%v
Trainer:   	%v
Duration:  	%v
CreateTime:	%v
EndTime:	%v
%v
`

func TopTrainingJobs(args []string, namespace string, allNamespaces bool, jobType types.TrainingJobType, instanceName string, notStop bool, format types.FormatStyle) error {
	if len(args) == 0 && notStop {
		return fmt.Errorf("You must specify the job name when using `-r` flag")
	}
	if !notStop {
		return topTrainingJobs(args, namespace, allNamespaces, jobType, instanceName, notStop, format)
	}
	for {
		err := topTrainingJobs(args, namespace, allNamespaces, jobType, instanceName, notStop, format)
		if err != nil {
			log.Errorf("%v", err)
		}
		t := time.Now()

		line := "------------------------------------------- %v ----------------------------------------------------"
		fmt.Printf(line+"\n", t.Format("2006-01-02 15:04:05"))
		time.Sleep(2 * time.Second)
	}
}

func topTrainingJobs(args []string, namespace string, allNamespaces bool, jobType types.TrainingJobType, instanceName string, notStop bool, format types.FormatStyle) error {
	if format == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	showSpecificJobMetric := false
	jobs := []TrainingJob{}
	if len(args) > 0 {
		showSpecificJobMetric = true
		job, err := SearchTrainingJob(args[0], namespace, jobType)
		if err != nil {
			return err
		}
		jobs = append(jobs, job)
	} else {
		allJobs, err := ListTrainingJobs(namespace, allNamespaces, jobType)
		if err != nil {
			return err
		}
		for _, j := range allJobs {
			jobs = append(jobs, j)
		}
	}
	jobs = makeTrainingJobOrderdByGPUCount(jobs)
	jobInfos := []types.TrainingJobInfo{}
	services, nodes := PrepareServicesAndNodesForTensorboard(jobs, allNamespaces)
	for _, job := range jobs {
		jobInfo := BuildJobInfo(job, true, services, nodes)
		jobInfos = append(jobInfos, *jobInfo)
	}
	switch format {
	case types.JsonFormat:
		outBytes, err := json.MarshalIndent(jobInfos, "", "    ")
		if err != nil {
			return err
		}
		fmt.Printf(string(outBytes))
		return nil
	case types.YamlFormat:
		outBytes, err := yaml.Marshal(jobInfos)
		if err != nil {
			return err
		}
		fmt.Printf(string(outBytes))
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	data := ""
	if showSpecificJobMetric {
		data = displayWithMetric(jobInfos, instanceName, notStop, format)
	} else {
		data = displayWithNoMetric(jobInfos, notStop, format, allNamespaces)
	}
	PrintLine(w, data)
	_ = w.Flush()
	return nil
}

func displayWithMetric(jobs []types.TrainingJobInfo, instanceName string, notStop bool, format types.FormatStyle) string {
	outputs := []string{}
	for _, jobInfo := range jobs {
		lines := []string{"", "Instances:", "  NAME\tSTATUS\tGPU(Request)\tNODE\tGPU(DeviceIndex)\tGPU(DutyCycle)\tGPU_MEMORY(Used/Total)"}
		lines = append(lines, "  ----\t------\t------------\t----\t----------------\t--------------\t---------------")
		for _, instance := range jobInfo.Instances {
			if instanceName != "" && instanceName != instance.Name {
				continue
			}
			if len(instance.GPUMetrics) == 0 {
				lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v\t%v\t%v\t%v",
					instance.Name,
					instance.Status,
					instance.RequestGPUs,
					instance.NodeIP,
					"N/A", "N/A", "N/A",
				))
				continue
			}
			for index, gpuId := range SortMapKeys(instance.GPUMetrics) {
				name := instance.Name
				status := instance.Status
				requestGPUs := fmt.Sprintf("%v", instance.RequestGPUs)
				hostIP := instance.NodeIP
				if index != 0 {
					name = ""
					status = ""
					requestGPUs = ""
					hostIP = ""
				}
				gpuMetric := instance.GPUMetrics[gpuId]
				lines = append(lines, fmt.Sprintf("  %v\t%v\t%v\t%v\t%v\t%.1f%%\t%.1f/%.1f(MiB)",
					name,
					status,
					requestGPUs,
					hostIP,
					gpuId,
					gpuMetric.GpuDutyCycle,
					fromByteToMiB(gpuMetric.GpuMemoryUsed),
					fromByteToMiB(gpuMetric.GpuMemoryTotal),
				))
			}
		}
		var duration int64
		var err error
		var endTime string
		jobInfo.Duration = strings.Replace(jobInfo.Duration, "s", "", -1)
		duration, err = strconv.ParseInt(jobInfo.Duration, 10, 64)
		if err != nil {
			log.Debugf("failed to parse duration: %v", err)

		}
		lines = append(lines, "", "GPUs:")
		lines = append(lines, fmt.Sprintf("  Allocated/Requested GPUs of Job: %v/%v", jobInfo.AllocatedGPU, jobInfo.RequestGPU))
		if jobInfo.Status == types.TrainingJobSucceeded || jobInfo.Status == types.TrainingJobFailed {
			endTime = util.GetFormatTime(jobInfo.CreationTimestamp + duration)
		}
		outputs = append(outputs, fmt.Sprintf(strings.Trim(topJobTemplate, "\n"),
			jobInfo.Name,
			jobInfo.Status,
			jobInfo.Namespace,
			jobInfo.Priority,
			strings.ToUpper(fmt.Sprintf("%v", jobInfo.Trainer)),
			util.ShortHumanDuration(time.Duration(duration)*time.Second),
			util.GetFormatTime(jobInfo.CreationTimestamp),
			endTime,
			strings.Join(lines, "\n"),
		))
	}
	return strings.Join(outputs, "\n")
}

func displayWithNoMetric(jobs []types.TrainingJobInfo, notStop bool, format types.FormatStyle, allNamespaces bool) string {
	var (
		totalAllocatedGPUs int64
		totalRequestedGPUs int64
	)
	namespace := ""
	if allNamespaces {
		namespace = "NAMESPACE\t"
	}
	lines := []string{fmt.Sprintf("%vNAME\tSTATUS\tTRAINER\tAGE\tGPU(Requested)\tGPU(Allocated)\tNODE", namespace)}
	for _, jobInfo := range jobs {
		if jobInfo.Status == "RUNNING" {
			totalRequestedGPUs += jobInfo.RequestGPU
			totalAllocatedGPUs += jobInfo.AllocatedGPU
		}
		hostIP := "N/A"
		for _, instance := range jobInfo.Instances {
			if instance.IsChief {
				hostIP = instance.NodeIP
			}
		}
		namespace = ""
		if allNamespaces {
			namespace = fmt.Sprintf("%v\t", jobInfo.Namespace)
		}
		var duration int64
		var err error
		jobInfo.Duration = strings.Replace(jobInfo.Duration, "s", "", -1)
		duration, err = strconv.ParseInt(jobInfo.Duration, 10, 64)
		if err != nil {
			log.Debugf("failed to parse duration: %v", err)

		}
		lines = append(lines, fmt.Sprintf("%v%v\t%v\t%v\t%v\t%v\t%v\t%v",
			namespace,
			jobInfo.Name,
			jobInfo.Status,
			strings.ToUpper(fmt.Sprintf("%v", jobInfo.Trainer)),
			util.ShortHumanDuration(time.Duration(duration)*time.Second),
			jobInfo.RequestGPU,
			jobInfo.AllocatedGPU,
			hostIP,
		))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Total Allocated/Requested GPUs of Training Jobs: %v/%v", totalAllocatedGPUs, totalRequestedGPUs))
	return strings.Join(lines, "\n")
}

func fromByteToMiB(value float64) float64 {
	return value / 1048576
}

func SortMapKeys(podMetric map[string]types.GpuMetric) []string {
	var keys []string
	for k := range podMetric {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
