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

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/prometheus"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

/*
* get App Configs by name, which is created by arena
 */
func getTrainingTypes(name, namespace string) (cms []string) {
	cms = []string{}
	for _, trainingType := range utils.GetTrainingJobTypes() {
		found := isTrainingConfigExist(name, string(trainingType), namespace)
		if found {
			cms = append(cms, string(trainingType))
		}
	}

	return cms
}

/**
*  check if the training config exist
 */
func isTrainingConfigExist(name, trainingType, namespace string) bool {
	configName := fmt.Sprintf("%s-%s", name, trainingType)
	return kubectl.CheckAppConfigMap(configName, namespace)
}

/**
* BuildTrainingJobInfo returns types.TrainingJobInfo
 */
func BuildJobInfo(job TrainingJob, showGPUs bool, services []*v1.Service, nodes []*v1.Node) *types.TrainingJobInfo {
	chiefPodName := ""
	//namespace := ""
	if job.ChiefPod() != nil {
		chiefPodName = job.ChiefPod().Name
		//namespace = job.ChiefPod().Namespace
	}
	tensorboard, err := tensorboardURL(job.Name(), job.Namespace(), services, nodes)
	if err != nil {
		log.Debugf("Tensorboard dones't show up because of %v, or tensorboard url %s", err, tensorboard)
	}
	jobGPUMetric := prometheus.JobGpuMetric{}
	instances := []types.TrainingJobInstance{}
	if showGPUs {
		jobGPUMetric, err = GetJobGpuMetric(config.GetArenaConfiger().GetClientSet(), job)
	}
	for _, pod := range job.AllPods() {
		isChief := false
		if pod.Name == chiefPodName {
			isChief = true
		}
		gpuMetrics := map[string]types.GpuMetric{}
		metrics, ok := jobGPUMetric[pod.Name]
		if ok {
			for gpuid, m := range metrics {
				log.Debugf("gpuid: %v,metric: %v", gpuid, *m)
				metric := *m
				gpuMetrics[gpuid] = types.GpuMetric{
					GpuDutyCycle:   metric.GpuDutyCycle,
					GpuMemoryTotal: metric.GpuMemoryTotal,
					GpuMemoryUsed:  metric.GpuMemoryUsed,
				}
			}
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		nodeName := "N/A"
		nodeIP := "N/A"
		if pod.Status.Phase != v1.PodPending {
			nodeIP = pod.Status.HostIP
			nodeName = pod.Spec.NodeName
		}
		count := utils.GPUCountInPod(pod)
		count += utils.AliyunGPUCountInPod(pod)
		instances = append(instances, types.TrainingJobInstance{
			Name:        pod.Name,
			IP:          pod.Status.PodIP,
			Status:      status,
			Age:         fmt.Sprintf("%vs", int(utils.GetDurationOfPod(pod).Seconds())),
			Node:        nodeName,
			NodeIP:      nodeIP,
			IsChief:     isChief,
			RequestGPUs: count,
			GPUMetrics:  gpuMetrics,
		})
	}

	trainingJobInfo := &types.TrainingJobInfo{
		Name:      job.Name(),
		Namespace: job.Namespace(),
		Status:    types.TrainingJobStatus(GetJobRealStatus(job)),
		//Duration:     util.ShortHumanDuration(job.Duration()),
		Duration:     fmt.Sprintf("%vs", int(job.Duration().Seconds())),
		Trainer:      types.TrainingJobType(job.Trainer()),
		Priority:     getPriorityClass(job),
		Tensorboard:  tensorboard,
		ChiefName:    chiefPodName,
		Instances:    instances,
		RequestGPU:   job.RequestedGPU(),
		AllocatedGPU: job.AllocatedGPU(),
	}

	if job.StartTime() != nil {
		trainingJobInfo.CreationTimestamp = job.StartTime().Unix()
	}

	return trainingJobInfo
}

/**
* getPriorityClass returns priority class name
 */
func getPriorityClass(job TrainingJob) string {
	pc := job.GetPriorityClass()
	if len(pc) == 0 {
		pc = "N/A"
	}

	return pc
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

func GetJobGpuMetric(client *kubernetes.Clientset, job TrainingJob) (jobMetric prometheus.JobGpuMetric, err error) {
	runningPods := []string{}
	jobStatus := job.GetStatus()
	jobGPUMetrics := prometheus.JobGpuMetric{}
	if jobStatus != "RUNNING" {
		return jobGPUMetrics, nil
	}
	pods := job.AllPods()
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodRunning {
			runningPods = append(runningPods, pod.Name)
		}
	}
	if len(runningPods) == 0 {
		return jobGPUMetrics, nil
	}
	podsMetrics, err := prometheus.GetPodsGpuInfo(client, runningPods)
	return podsMetrics, err
}
