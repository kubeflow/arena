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
	"github.com/kubeflow/arena/pkg/util/kubeclient"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

/*
* get App Configs by name, which is created by arena
 */
func getTrainingTypes(name, namespace string) (cms []string, err error) {
	cms = []string{}
	var errNoPrivilege error
	for _, trainingType := range utils.GetTrainingJobTypes() {
		canDelete, err := kubeclient.CheckJobIsOwnedByUser(namespace, name, trainingType)
		if err != nil {
			continue
		}
		if !canDelete {
			errNoPrivilege = types.ErrNoPrivilegesToOperateJob
		}
		cms = append(cms, string(trainingType))
	}
	if len(cms) == 0 {
		log.Infof("The training job '%v' does not exist,skip to delete it", name)
		return nil, types.ErrTrainingJobNotFound
	}
	if len(cms) > 1 {
		return nil, fmt.Errorf("there are more than 1 training jobs with the same name %s, please double check with `arena list | grep %s`. And use `arena delete %s --type` to delete the exact one",
			name,
			name,
			name)
	}
	if errNoPrivilege != nil {
		return nil, errNoPrivilege
	}
	return cms, nil
}

/**
* BuildTrainingJobInfo returns types.TrainingJobInfo
 */
func BuildJobInfo(job TrainingJob, showGPUs bool, services []*corev1.Service, nodes []*corev1.Node) *types.TrainingJobInfo {
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
		if err != nil {
			log.Debugf("get job gpu metric failed, err: %s", err)
		}
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
		if pod.Status.Phase != corev1.PodPending {
			nodeIP = pod.Status.HostIP
			nodeName = pod.Spec.NodeName
		}
		count := utils.GPUCountInPod(pod)
		count += utils.AliyunGPUCountInPod(pod)
		instances = append(instances, types.TrainingJobInstance{
			Name:              pod.Name,
			IP:                pod.Status.PodIP,
			Status:            status,
			Age:               fmt.Sprintf("%vs", int(utils.GetDurationOfPod(pod).Seconds())),
			Node:              nodeName,
			NodeIP:            nodeIP,
			IsChief:           isChief,
			RequestGPUs:       count,
			GPUMetrics:        gpuMetrics,
			CreationTimestamp: pod.CreationTimestamp.Unix(),
		})
	}

	trainingJobInfo := &types.TrainingJobInfo{
		Name:      job.Name(),
		UUID:      job.Uid(),
		Namespace: job.Namespace(),
		Status:    types.TrainingJobStatus(GetJobRealStatus(job)),
		//Duration:     util.ShortHumanDuration(job.Duration()),
		Duration:     fmt.Sprintf("%vs", int(job.Duration().Seconds())),
		Trainer:      job.Trainer(),
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
			if pod.Status.Phase == corev1.PodPending {
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
		if pod.Status.Phase == corev1.PodRunning {
			runningPods = append(runningPods, pod.Name)
		}
	}
	if len(runningPods) == 0 {
		return jobGPUMetrics, nil
	}
	podsMetrics, err := prometheus.GetPodsGpuInfo(client, runningPods)
	return podsMetrics, err
}
