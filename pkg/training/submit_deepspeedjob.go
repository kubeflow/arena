// Copyright 2023 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
)

func SubmitDeepSpeedJob(namespace string, submitArgs *types.SubmitDeepSpeedJobArgs) (err error) {
	submitArgs.Namespace = namespace
	// generate ssh secret
	if submitArgs.SSHSecret == "" {
		submitArgs.SecretData, err = util.GenerateRsaKey()
		if err != nil {
			log.Infof("The Job %s generate ssh secret failed, err: %s", submitArgs.Name, err)
			return err
		}
	}

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
		if err == types.ErrNoPrivilegesToOperateJob {
			return fmt.Errorf("the job %s is already exist and it owned by other user,you have no privileges to operate it", submitArgs.Name)
		}
		return err
	}
	// the master is also considered as a worker
	deepspeedjobChart := util.GetChartsFolder() + "/etjob"
	err = workflow.SubmitJob(submitArgs.Name, string(types.DeepSpeedTrainingJob), namespace, submitArgs, deepspeedjobChart, submitArgs.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", submitArgs.Name)
	log.Infof("You can run `arena get %s --type %s -n %s` to check the job status", submitArgs.Name, submitArgs.TrainingType, submitArgs.Namespace)
	return nil
}
