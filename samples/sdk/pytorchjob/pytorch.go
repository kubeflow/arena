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
	jobName := "pytorch-test"
	jobType := types.PytorchTrainingJob
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
	   arena \
	   submit \
	   pytorchjob \
	   --name=pytorch-standalone-test \
	   --gpus=1 \
	   --sync-mode=git \
	   --tensorboard \
	   --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
	   --loglevel debug \
	   --image=registry.cn-shanghai.aliyuncs.com/ai-samples/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
	   "python /root/code/mnist-pytorch/mnist.py --backend gloo"
	*/
	submitJob, err := training.NewPytorchJobBuilder().
		Name(jobName).
		GPUCount(1).
		SyncMode("git").
		SyncSource("https://code.aliyun.com/370272561/mnist-pytorch.git").
		Image("registry.cn-shanghai.aliyuncs.com/ai-samples/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime").
		Command([]string{"python /root/code/mnist-pytorch/mnist.py --backend gloo"}).Build()
	if err != nil {
		fmt.Printf("failed to build pytorchjob,reason: %v\n", err)
		return
	}
	// submit tfjob
	if err := client.Training().Submit(submitJob); err != nil {
		fmt.Printf("failed to submit job,reason: %v\n", err)
		return
	}
	// list all jobs
	jobInfos, err := client.Training().List(true, types.AllTrainingJob)
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
		job, err := client.Training().Get(jobName, jobType)
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
