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

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubeclient"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func DeleteTrainingJob(jobName, namespace string, jobType types.TrainingJobType) error {
	var trainingTypes []string
	if jobType == types.UnknownTrainingJob {
		return fmt.Errorf("Unsupport job type,arena only supports: [%v]", utils.GetSupportTrainingJobTypesInfo())
	}
	// 1. Handle legacy training job
	err := helm.DeleteRelease(jobName)
	if err == nil {
		log.Infof("Delete the job %s successfully.", jobName)
		return nil
	}
	log.Debugf("%s wasn't deleted by helm due to %v", jobName, err)
	// if the jobType is sure,delete the job
	if jobType != types.AllTrainingJob {
		canDelete, err := kubeclient.CheckJobIsOwnedByUser(namespace, jobName, jobType)
		if err != nil {
			if err == kubeclient.ErrConfigMapNotFound {
				log.Errorf("The training job '%v' does not exist,skip to delete it", jobName)
				return types.ErrTrainingJobNotFound
			}
			return err
		}
		if !canDelete {
			return types.ErrNoPrivilegesToOperateJob
		}
		return workflow.DeleteJob(jobName, namespace, string(jobType))
	}
	// 2. Handle training jobs created by arena
	trainingTypes, err = getTrainingTypes(jobName, namespace)
	if err != nil {
		return err
	}
	err = workflow.DeleteJob(jobName, namespace, trainingTypes[0])
	if err != nil {
		return err
	}
	log.Infof("The training job %s has been deleted successfully", jobName)
	// (TODO: cheyang)3. Handle training jobs created by others, to implement
	return nil
}
