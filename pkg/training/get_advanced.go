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

	apistypes "github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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
func BuildJobInfo(job TrainingJob) *apistypes.TrainingJobInfo {
	chiefPodName := ""
	namespace := ""
	if job.ChiefPod() != nil {
		chiefPodName = job.ChiefPod().Name
		namespace = job.ChiefPod().Namespace
	}
	tensorboard, err := tensorboardURL(job.Name(), namespace)
	if err != nil {
		log.Debugf("Tensorboard dones't show up because of %v, or tensorboard url %s", err, tensorboard)
	}
	instances := []apistypes.TrainingJobInstance{}
	for _, pod := range job.AllPods() {
		isChief := false
		if pod.Name == chiefPodName {
			isChief = true
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		instances = append(instances, apistypes.TrainingJobInstance{
			Name:    pod.Name,
			Status:  status,
			Age:     util.ShortHumanDuration(job.Age()),
			Node:    pod.Status.HostIP,
			IsChief: isChief,
		})
	}

	return &apistypes.TrainingJobInfo{
		Name:        job.Name(),
		Namespace:   job.Namespace(),
		Status:      apistypes.TrainingJobStatus(GetJobRealStatus(job)),
		Duration:    util.ShortHumanDuration(job.Duration()),
		Trainer:     apistypes.TrainingJobType(job.Trainer()),
		Priority:    getPriorityClass(job),
		Tensorboard: tensorboard,
		ChiefName:   chiefPodName,
		Instances:   instances,
	}
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
