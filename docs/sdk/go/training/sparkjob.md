# Build Spark Training Job

This section introduces how to customly build a Spark training job.

## Path

    pkg/apis/training.SparkJobBuilder

## Function

    func NewSparkJobBuilder() *SparkJobBuilder

## Parameters

SparkJobBuilder has following functions to custom your Spark training job.

| function   |  description  | matches cli option |
|:---|:--:|:---|
|  Name(name string) *SparkJobBuilder   | specify the spark job name   |  --name  |
| Image(image string) *SparkJobBuilder    | specify the image   |  --image  |
|  ExecutorReplicas(replicas int) *SparkJobBuilder    | specify the executor replicas   |  --replicas  |
|  MainClass(mainClass string) *SparkJobBuilder   | specify the main class    | --main-class   |
| Jar(jar string) *SparkJobBuilder    |  specify the jar  |  --jar  |
|  DriverCPURequest(request int) *SparkJobBuilder   | specify the driver cpu request   |  --driver-cpu-request  |
|  DriverMemoryRequest(memory string) *SparkJobBuilder   | specify the driver memory request   |  --driver-memory-request  |
|  ExecutorCPURequest(request int) *SparkJobBuilder    | specify the executor cpu request   |  --executor-cpu-request |
|   ExecutorMemoryRequest(memory string) *SparkJobBuilder  |  specify the executor memory request   | --executor-memory-request   |
| Build() (*Job, error)    | build the spark job   |  -  |

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
        jobName := "test-sparkjob"
        jobType := types.SparkTrainingJob
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
            arena submit sparkjob  \
            --name=demo \
            --image=registry.aliyuncs.com/acs/spark:v2.4.0 \
            --main-class=org.apache.spark.examples.SparkPi \
            --jar=local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar
        */
        submitJob, err := training.NewSparkJobBuilder().
            Name(jobName).
            Image("registry.aliyuncs.com/acs/spark:v2.4.0").
            MainClass("org.apache.spark.examples.SparkPi").
            Jar("local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar").Build()
        if err != nil {
            fmt.Printf("failed to build sparkjob,reason: %v\n", err)
            return
        }
        // submit sparkjob
        if err := client.Training().Submit(submitJob); err != nil {
            fmt.Printf("failed to submit job,reason: %v\n", err)
            return
        }
    }
