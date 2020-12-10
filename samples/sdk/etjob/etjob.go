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
	jobName := "etjob-test"
	jobType := types.ETTrainingJob
	// create arena client
	client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
		Kubeconfig: "",
		LogLevel:   "debug",
		Namespace:  "default",
	})
	// create tfjob
	/* command:
		arena submit etjob \
	    --name=elastic-training \
	    --gpus=1 \
	    --workers=3 \
	    --max-workers=9 \
	    --min-workers=1 \
	    --image=registry.cn-hangzhou.aliyuncs.com/ai-samples/horovod:0.20.0-tf2.3.0-torch1.6.0-mxnet1.6.0.post0-py3.7-cuda10.1 \
	    --working-dir=/examples \
	    "horovodrun
	    -np \$((\${workers}*\${gpus}))
	    --min-np \$((\${minWorkers}*\${gpus}))
	    --max-np \$((\${maxWorkers}*\${gpus}))
	    --host-discovery-script /usr/local/bin/discover_hosts.sh
	    python /examples/elastic/tensorflow2_mnist_elastic.py
	    "
	*/
	job, err := training.NewETJobBuilder().
		Name(jobName).
		GPUCount(1).
		WorkerCount(3).
		WorkingDir("/examples").
		MaxWorkers(9).
		MinWorkers(1).
		Image("registry.cn-hangzhou.aliyuncs.com/ai-samples/horovod:0.20.0-tf2.3.0-torch1.6.0-mxnet1.6.0.post0-py3.7-cuda10.1").
		Command([]string{
			`horovodrun`,
			`-np $((${workers}*${gpus}))`,
			`--min-np $((${minWorkers}*${gpus}))`,
			`--max-np $((${maxWorkers}*${gpus}))`,
			`--host-discovery-script /usr/local/bin/discover_hosts.sh`,
			`python /examples/elastic/tensorflow2_mnist_elastic.py`,
		}).Build()
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
