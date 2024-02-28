package main

import (
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/logger"
	"github.com/kubeflow/arena/pkg/apis/serving"
	"github.com/kubeflow/arena/pkg/apis/types"
)

func main() {
	jobName := "fast-style-transfer"
	jobVersion := "alpha"
	jobType := types.CustomServingJob
	// create arena client
	client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
		Kubeconfig: "",
		LogLevel:   "info",
		Namespace:  "default",
	})
	if err != nil {
		fmt.Printf("failed to create arena client, err: %v\n", err)
		return
	}
	// create tfjob
	/* command:
	arena serve custom \
	--name=fast-style-transfer \
	--gpus=1 \
	--version=alpha \
	--replicas=1 \
	--restful-port=5000 \
	--image=happy365/fast-style-transfer:latest \
	"python app.py"
	*/
	job, err := serving.NewCustomServingJobBuilder().
		Name(jobName).
		Namespace("default").
		GPUCount(1).
		Version(jobVersion).
		Replicas(1).
		RestfulPort(5000).
		Image("happy365/fast-style-transfer:latest").Command([]string{"python app.py"}).
		Annotations(map[string]string{"testAnnotation": "v1"}).
		Build()
	if err != nil {
		fmt.Printf("failed to build custom serving job, reason: %v\n", err)
		return
	}

	// submit custom serving
	if err := client.Serving().Submit(job); err != nil {
		fmt.Printf("failed to submit custom serving job, reason: %v\n", err)
		return
	}

	// update custom serve
	updateJob, err := serving.NewUpdateCustomServingJobBuilder().
		Name(jobName).
		Namespace("default").
		Version(jobVersion).
		Replicas(1).
		Annotations(map[string]string{"testAnnotation": "v2"}).
		Build()
	if err != nil {
		fmt.Printf("failed to build update custom serving job, reason: %v\n", err)
		return
	}

	if err := client.Serving().Update(updateJob); err != nil {
		fmt.Printf("failed to update custom serving job, resion: %v\n", err)
		return
	}

	// list all jobs
	jobInfos, err := client.Serving().List(true, types.AllServingJob)
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
		job, err := client.Serving().Get(jobName, jobVersion, jobType)
		if err != nil {
			fmt.Printf("failed to get job,reason: %v\n", err)
			return
		}
		if i == 0 {
			fmt.Printf("timeout for waiting job to be running,exit\n")
			return
		}
		if job.Available != job.Desired {
			fmt.Printf("name: %v, available: %v,job desired: %v,waiting...\n", jobName, job.Available, job.Desired)
			continue
		}
		fmt.Printf("job info: %v\n", job)
		break
	}

	// get the job log,the status of job must be RUNNING
	logArgs, err := logger.NewLoggerBuilder().Build()
	if err != nil {
		fmt.Printf("failed to build log args,reason: %v\n", err)
	}
	if err := client.Serving().Logs(jobName, jobVersion, jobType, logArgs); err != nil {
		fmt.Printf("failed to get job log,reason: %v\n", err)
		return
	}
	err = client.Serving().Delete(jobType, jobVersion, jobName)
	if err != nil {
		fmt.Printf("failed to delete job,reason: %v", err)
		return
	}
}
