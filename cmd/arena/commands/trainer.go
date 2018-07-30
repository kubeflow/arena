package commands

import (
	"sort"

	"k8s.io/client-go/kubernetes"
)

// construct the trainer list
func NewTrainers(client *kubernetes.Clientset) []Trainer {

	trainers := []Trainer{}
	trainerInits := []func(client *kubernetes.Clientset) Trainer{
		NewMPIJobTrainer,
		NewStandaloneJobTrainer,
		NewTensorFlowJobTrainer}

	for _, init := range trainerInits {
		trainers = append(trainers, init(client))
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
