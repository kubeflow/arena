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
	"sort"

	"github.com/kubeflow/arena/pkg/client"
)

// construct the trainer list
func NewTrainers(kubeClient *client.Client) []Trainer {

	trainers := []Trainer{}
	trainerInits := []func(kubeClient client.Client) Trainer{
		// NewHorovodJobTrainer,
		// NewStandaloneJobTrainer,
		// NewTensorFlowJobTrainer,
		NewMPIJobTrainer,
		// NewSparkJobTrainer,
		// NewVolcanoJobTrainer,
		NewRunaiTrainer}

	for _, init := range trainerInits {
		trainers = append(trainers, init(*kubeClient))
	}

	return trainers
}

type orderedTrainingJob []TrainingJob

func (this orderedTrainingJob) Len() int {
	return len(this)
}

func (this orderedTrainingJob) Less(i, j int) bool {
	return this[i].RequestedGPU() > this[j].RequestedGPU()
}

func (this orderedTrainingJob) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type orderedTrainingJobByAge []TrainingJob

func (this orderedTrainingJobByAge) Len() int {
	return len(this)
}

func (this orderedTrainingJobByAge) Less(i, j int) bool {
	if this[i].StartTime() == nil {
		return true
	} else if this[j].StartTime() == nil {
		return false
	}

	return this[i].StartTime().After(this[j].StartTime().Time)
}

func (this orderedTrainingJobByAge) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func makeTrainingJobOrderdByAge(jobList []TrainingJob) []TrainingJob {
	newJoblist := make(orderedTrainingJobByAge, 0, len(jobList))
	for _, v := range jobList {
		newJoblist = append(newJoblist, v)
	}
	sort.Sort(newJoblist)
	return []TrainingJob(newJoblist)
}

func makeTrainingJobOrderdByGPUCount(jobList []TrainingJob) []TrainingJob {
	newJoblist := make(orderedTrainingJob, 0, len(jobList))
	for _, v := range jobList {
		newJoblist = append(newJoblist, v)
	}
	sort.Sort(newJoblist)
	return []TrainingJob(newJoblist)
}
