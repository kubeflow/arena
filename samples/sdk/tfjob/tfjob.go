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

package main

import (
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/logger"
	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
)

func main() {
	jobName := "test-dist-tfjob"
	jobType := types.TFTrainingJob
	// create arena client
	client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
		Kubeconfig: "",
		LogLevel:   "info",
		Namespace:  "default",
	})
	if err != nil {
		fmt.Printf("failed to create arena client,reason: %v", err)
		return
	}
	// create tfjob
	/* command:
			arena submit tfjob \
	        --name=tf-distributed-test \
	        --gpus=1 \
	        --workers=1 \
	        --worker-image=cheyang/tf-mnist-distributed:gpu \
	        --ps-image=cheyang/tf-mnist-distributed:cpu \
	        --ps=1 \
	        --tensorboard \
	        "python /app/main.py"
	*/
	submitJob, err := training.NewTFJobBuilder(nil).
		Name(jobName).
		GPUCount(1).
		WorkerCount(1).
		WorkerImage("cheyang/tf-mnist-distributed:gpu").
		PsImage("cheyang/tf-mnist-distributed:cpu").
		PsCount(1).
		EnableTensorboard().
		Command([]string{"'python /app/main.py'"}).Build()
	if err != nil {
		fmt.Printf("failed to build tfjob,reason: %v\n", err)
		return
	}
	// submit tfjob
	if err := client.Training().Submit(submitJob); err != nil {
		fmt.Printf("failed to submit job,reason: %v\n", err)
		return
	}
	// list all jobs
	jobInfos, err := client.Training().List(true, types.AllTrainingJob, false)
	if err != nil {
		fmt.Printf("failed to list all jobs in namespace,reason: %v\n", err)
		return
	}
	for _, job := range jobInfos {
		fmt.Printf("found job %s\n", job.Name)
	}
	// get the job information and wait it to be running,timeout: 500s
	for i := 250; i >= 0; i-- {
		time.Sleep(2 * time.Second)
		job, err := client.Training().Get(jobName, jobType, false)
		if err != nil {
			fmt.Printf("failed to get job,reason: %v\n", err)
			return
		}
		if job.Status == "PENDING" {
			fmt.Printf("current status of job %v is: %v,waiting...\n", jobName, job.Status)
			continue
		}
		if job.Status == "RUNNING" || job.Status == "FAILED" {
			fmt.Printf("job info: %v\n", job)
			break
		}
		if i == 0 {
			fmt.Printf("timeout for waiting job to be running,exit\n")
			return
		}
	}
	// get the job log,the status of job must be RUNNING
	logArgs, err := logger.NewLoggerBuilder().Follow().Build()
	if err != nil {
		fmt.Printf("failed to build log args,reason: %v\n", err)
	}
	if err := client.Training().Logs(jobName, jobType, logArgs); err != nil {
		fmt.Printf("failed to get job log,reason: %v\n", err)
		return
	}
	err = client.Training().Delete(jobType, jobName)
	if err != nil {
		fmt.Printf("failed to delete job,reason: %v", err)
		return
	}
}
