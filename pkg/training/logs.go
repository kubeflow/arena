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
	podlogs "github.com/kubeflow/arena/pkg/podlogs"
)

// AcceptJobLog is used for arena-go-sdk
func AcceptJobLog(jobName string, trainingType types.TrainingJobType, args *types.LogArgs) error {
	namespace := args.Namespace
	// 1.transfer the training job type
	if trainingType == types.UnknownTrainingJob {
		return fmt.Errorf("Unsupport job type,arena only supports: [%v]", utils.GetSupportTrainingJobTypesInfo())
	}
	// 2.search the training job
	job, err := SearchTrainingJob(jobName, namespace, trainingType)
	if err != nil {
		return err
	}
	chiefPod := job.ChiefPod()
	// 3.if instance name not set,set the chief pod name to instance name
	if args.InstanceName == "" && chiefPod != nil {
		args.InstanceName = chiefPod.Name
	}
	podStatuses := map[string]string{}
	for _, pod := range job.AllPods() {
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		podStatuses[pod.Name] = status
	}
	// 4.if the instance name is invalid,return error
	_, ok := podStatuses[args.InstanceName]
	if !ok {
		return fmt.Errorf("invalid instance name %v in job %v,please use 'arena get %v' to make sure instance name.",
			args.InstanceName,
			jobName,
			jobName,
		)
	}
	// 5.if the instance status is not running,return error
	//if status != "Running" {
	//	return fmt.Errorf("failed to get logs of instance %v,because it is not running,please use 'arena get %v' to make sure instance status",
	//		args.InstanceName,
	//		jobName,
	//	)
	//}
	logger := podlogs.NewPodLogger(args)
	_, err = logger.AcceptLogs()
	return err
}

func getTrainingJobTypes() []string {
	jobTypes := []string{}
	for _, trainingType := range utils.GetTrainingJobTypes() {
		jobTypes = append(jobTypes, string(trainingType))
	}
	return jobTypes
}
