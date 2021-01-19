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
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/AliyunContainerService/et-operator/pkg/api/v1alpha1"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

const (
	ETJOB_MAXWORKERS = 1000
	ETJOB_MINWORKERS = 1
)

func SubmitETJob(namespace string, submitArgs *types.SubmitETJobArgs) (err error) {
	submitArgs.Namespace = namespace
	trainers := GetAllTrainers()
	trainer, ok := trainers[submitArgs.TrainingType]
	if !ok {
		return fmt.Errorf("not found trainer whose type is %v", submitArgs.TrainingType)
	}
	job, err := trainer.GetTrainingJob(submitArgs.Name, namespace)
	// if job has been existed,skip to create it and return an error
	if err == nil && job != nil {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", submitArgs.Name, submitArgs.Name)
	}
	// if error is unknown,return an error
	if err != types.ErrTrainingJobNotFound {
		return err
	}
	// the master is also considered as a worker
	etjobChart := util.GetChartsFolder() + "/etjob"
	err = workflow.SubmitJob(submitArgs.Name, string(types.ETTrainingJob), namespace, submitArgs, etjobChart, submitArgs.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", submitArgs.Name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", submitArgs.Name, submitArgs.TrainingType)
	return nil
}

func SubmitScaleInETJob(namespace string, submitArgs *types.ScaleInETJobArgs) error {
	etjobName := submitArgs.Name
	trainers := GetAllTrainers()
	trainer, ok := trainers[submitArgs.JobType]
	if !ok {
		return fmt.Errorf("not found trainer whose type is %v", submitArgs.JobType)
	}
	job, err := trainer.GetTrainingJob(etjobName, namespace)
	if err != nil {
		if err == types.ErrTrainingJobNotFound {
			return err
		}
		return fmt.Errorf("Check %s exist due to error %v", etjobName, err)
	}
	if job.GetStatus() != "RUNNING" && job.GetStatus() != "SCALING" {
		return fmt.Errorf("the job: %s status: %s , is not RUNNING or SCALING, please try again later.", etjobName, job.GetStatus())
	}
	currentWorkers := getETJobCurrentReplicas(job)
	minWorkers := getETJobMinReplicas(job)
	log.Debugf("currentWorkers: %v, minWorkers: %v", currentWorkers, minWorkers)
	if currentWorkers-submitArgs.Count < minWorkers {
		return fmt.Errorf("the number of current workers minus the number of scaling in is less than the min-workers. please try again later.")
	}
	scaleName := fmt.Sprintf("%s-%d", etjobName, time.Now().Unix())
	log.Debugf("submitArgs: %v", submitArgs)
	scaleinETChart := util.GetChartsFolder() + "/scalein"
	err = workflow.SubmitJob(scaleName, "scalein", namespace, submitArgs, scaleinETChart)
	if err != nil {
		return err
	}
	log.Infof("The scalein job %s has been submitted successfully", scaleName)
	return nil
}

func SubmitScaleOutETJob(namespace string, submitArgs *types.ScaleOutETJobArgs) error {
	etjobName := submitArgs.Name
	trainers := GetAllTrainers()
	trainer, ok := trainers[submitArgs.JobType]
	if !ok {
		return fmt.Errorf("not found trainer whose type is %v", submitArgs.JobType)
	}
	job, err := trainer.GetTrainingJob(etjobName, namespace)
	if err != nil {
		if err == types.ErrTrainingJobNotFound {
			return err
		}
		return fmt.Errorf("Check %s exist due to error %v", etjobName, err)
	}
	if job.GetStatus() != "RUNNING" && job.GetStatus() != "SCALING" {
		return fmt.Errorf("the job: %s status: %s , is not RUNNING or SCALING, please try again later.", etjobName, job.GetStatus())
	}
	currentWorkers := getETJobCurrentReplicas(job)
	maxWorkers := getETJobMaxReplicas(job)
	log.Debugf("currentWorkers: %v, maxWorkers: %v", currentWorkers, maxWorkers)
	if currentWorkers+submitArgs.Count > maxWorkers {
		return fmt.Errorf("The number of scaling out plus the number of current workers exceeds the max-workers. please try again later.")
	}
	scaleName := fmt.Sprintf("%s-%d", etjobName, time.Now().Unix())
	log.Debugf("submitArgs: %v", submitArgs)
	scaleoutETChart := util.GetChartsFolder() + "/scaleout"
	err = workflow.SubmitJob(scaleName, "scaleout", namespace, submitArgs, scaleoutETChart)
	if err != nil {
		return err
	}
	log.Infof("The scaleout job %s has been submitted successfully", scaleName)
	return nil
}

func getETJobMaxReplicas(job TrainingJob) (maxReplicas int) {
	etJob := job.GetTrainJob().(*v1alpha1.TrainingJob)
	_, worker := parseAnnotations(etJob)
	maxReplicas = ETJOB_MAXWORKERS
	if worker != nil {
		if _, ok := worker["maxReplicas"]; ok {
			maxReplicas = int(worker["maxReplicas"].(float64))
		}
	}
	return maxReplicas
}

func getETJobMinReplicas(job TrainingJob) (minReplicas int) {
	etJob := job.GetTrainJob().(*v1alpha1.TrainingJob)
	_, worker := parseAnnotations(etJob)
	minReplicas = ETJOB_MINWORKERS
	if worker != nil {
		if _, ok := worker["minReplicas"]; ok {
			minReplicas = int(worker["minReplicas"].(float64))
		}
	}
	return minReplicas
}

func getETJobCurrentReplicas(job TrainingJob) (currentReplicas int) {
	return len(job.AllPods()) - 1
}
