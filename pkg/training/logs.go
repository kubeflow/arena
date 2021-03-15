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
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	podlogs "github.com/kubeflow/arena/pkg/podlogs"
	v1 "k8s.io/api/core/v1"
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
	// 3.if instance name not set,set the chief pod name to instance name
	if args.InstanceName == "" {
		name, err := getInstanceName(job)
		if err != nil {
			return err
		}
		args.InstanceName = name
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

func getInstanceName(job TrainingJob) (string, error) {
	pods := job.AllPods()
	// if not found pods,return an error
	if pods == nil || len(pods) == 0 {
		return "", fmt.Errorf("not found instances of the job %v", job.Name())
	}
	// if the job has only one pod,return its' name
	if len(pods) == 1 {
		return pods[0].Name, nil
	}
	// if job has many pods and the chief pod name is existed,return it
	if job.ChiefPod() != nil && job.ChiefPod().Name != "" {
		return job.ChiefPod().Name, nil
	}
	// return an error
	return "", fmt.Errorf("%v", moreThanOneInstanceHelpInfo(pods))
}

func moreThanOneInstanceHelpInfo(pods []*v1.Pod) string {
	header := fmt.Sprintf("There is %d instances have been found:", len(pods))
	lines := []string{}
	footer := fmt.Sprintf("please use '-i' or '--instance' to filter.")
	for _, p := range pods {
		lines = append(lines, fmt.Sprintf("%v", p.Name))
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(lines, "\n"), footer)
}
