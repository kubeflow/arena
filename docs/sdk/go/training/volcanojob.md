# Build Volcano Training Job

This section introduces how to customly build a Volcano training job.

## Path

    pkg/apis/training.VolcanoJobBuilder 

## Function

    func NewVolcanoJobBuilder() *VolcanoJobBuilder

## Parameters

VolcanoJobBuilder has following functions to custom your Volcano training job.

| function   |  description  | matches cli option |
|:---|:--:|:---|
|  Name(name string) *VolcanoJobBuilder| specify the job name| --name|
| Command(args []string) *VolcanoJobBuilder| specify the job command| -|
| MinAvailable(minAvailable int) *VolcanoJobBuilder| specify the min avaliable tasks| --min-available|
|Queue(queue string) *VolcanoJobBuilder | specify the queue|--queue|
| SchedulerName(name string) *VolcanoJobBuilder | specify the scheduler name|--scheduler-name|
|TaskImages(images []string) *VolcanoJobBuilder| specify the task images|--task-images|
|TaskName(name string) *VolcanoJobBuilder| specify the task name| --task-name|
|TaskReplicas(replicas int) *VolcanoJobBuilder |specify the task replicas| --task-replicas|
|TaskCPU(cpu string) *VolcanoJobBuilder|specify the cpu limits of task|--task-cpu|
|TaskMemory(mem string) *VolcanoJobBuilder|specify the memory limits of task|--task-memory|
| TaskPort(port int) *VolcanoJobBuilder | specify the task port|--task-port|
| Build() (*Job, error) |build the volcano training job|-|
## Example

    package main

    import (
        "fmt"
        "time"

        "github.com/kubeflow/arena/pkg/apis/arenaclient"
        "github.com/kubeflow/arena/pkg/apis/training"
        "github.com/kubeflow/arena/pkg/apis/types"
    )

    func main() {
        jobName := "test-volcanojob"
        jobType := types.VolcanoTrainingJob
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
        // create spark job
        /* command:
            arena submit volcanojob \
            --name demo12 \
            --taskImages busybox,busybox  \
            --taskReplicas 2
        */
        submitJob, err := training.NewVolcanoJobBuilder().
            Name(jobName).
            TaskImages("busybox,busybox").
            TaskReplicas(2).
            Build()
        if err != nil {
            fmt.Printf("failed to build volcanojob,reason: %v\n", err)
            return
        }
        // submit volcanojob
        if err := client.Training().Submit(submitJob); err != nil {
            fmt.Printf("failed to submit job,reason: %v\n", err)
            return
        }
    }
