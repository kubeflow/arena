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
	jobName := "mpi-standalone-test"
	jobType := types.MPITrainingJob
	// create arena client
	client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
		Kubeconfig: "",
		LogLevel:   "info",
		Namespace:  "default",
	})
	// create tfjob
	/* command:
	arena submit mpijob --name=mpi-standalone-test \
	--gpus=1 \
	--workers=1 \
	--working-dir=/perseus-demo/tensorflow-demo/ \
	--image=registry.cn-beijing.aliyuncs.com/kube-ai/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
	--tensorboard \
	"mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"
	*/
	job, err := training.NewMPIJobBuilder().
		Name(jobName).
		GPUCount(1).
		WorkerCount(2).
		WorkingDir("/perseus-demo/tensorflow-demo/").
		EnableTensorboard().
		Image("registry.cn-beijing.aliyuncs.com/kube-ai/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5").
		Command([]string{"mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64 --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"}).Build()
	if err != nil {
		fmt.Printf("failed to build mpijob,reason: %v\n", err)
		return
	}
	// submit tfjob
	if err := client.Training().Submit(job); err != nil {
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
