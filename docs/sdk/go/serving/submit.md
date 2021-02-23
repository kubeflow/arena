# Submit The Serving Job

This API is used to submit a serving job.

## Path

pkg/apis/arenaclient.ServingJobClient

## Function

	func (t *ServingJobClient) Submit(job *apiserving.Job) error 

## Parameters

* job(type pkg/apis/serving.Job) => the job which will be submitted,it must be created by some serving job builders, please refer the apis of serving job builders to build your serving jobs.
  
## Example

### Submit a custom serving job

    package main

    import (
        "fmt"
        "time"

        "github.com/kubeflow/arena/pkg/apis/arenaclient"
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
            fmt.Printf("failed to create arena client,reason: %v", err)
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
        submitJob, err := serving.NewCustomServingJobBuilder().
            Name(jobName).
            GPUCount(1).
            Version(jobVersion).
            Replicas(1).
            RestfulPort(5000).
            Image("happy365/fast-style-transfer:latest").
            Command("python app.py").
            Build()
        if err != nil {
            fmt.Printf("failed to build custom serving job,reason: %v\n", err)
            return
        }
        // submit custom serving job
        if err := client.Serving().Submit(submitJob); err != nil {
            fmt.Printf("failed to submit job,reason: %v\n", err)
            return
        }
    }