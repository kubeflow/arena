// Copyright 2024 The Kubeflow Authors
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

package serving

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/podlogs"
)

func AcceptJobLog(name, version string, jobType types.ServingJobType, args *types.LogArgs) error {
	namespace := args.Namespace
	job, err := SearchServingJob(namespace, name, version, jobType)
	if err != nil {
		return err
	}
	jobInfo := job.Convert2JobInfo()
	// 1.if not found instances,return an error
	if len(jobInfo.Instances) == 0 {
		return fmt.Errorf("not found instances of serving job,please use 'arena serve get %v' to get job information", name)
	}
	// 2.if instance name is null and job has more than one instance,return an error
	// push user to select one
	if len(jobInfo.Instances) > 1 && args.InstanceName == "" {
		return fmt.Errorf("%v", moreThanOneInstanceHelpInfo(jobInfo.Instances))
	}
	// 3.if user not specify the instance name and the job has only one instance name,pick the instance
	if args.InstanceName == "" {
		args.InstanceName = jobInfo.Instances[0].Name
	}
	// 4.if the instance name is invalid,return an error
	exists := map[string]bool{}
	for _, i := range jobInfo.Instances {
		exists[i.Name] = true
	}
	if _, ok := exists[args.InstanceName]; !ok {
		return fmt.Errorf("invalid instance name %v of serving job %v,please use 'arena serve get %v' to get instance names.", args.InstanceName, name, name)
	}
	logger := podlogs.NewPodLogger(args)
	_, err = logger.AcceptLogs()
	return err
}
