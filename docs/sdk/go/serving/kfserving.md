# Build Custom Serving Job

This section introduces how to customly build a Kubeflow serving job.

## Path

    pkg/apis/serving.KFServingJobBuilder

## Function

    func NewKFServingJobBuilder() *KFServingJobBuilder 

## Parameters

KFServingJobBuilder has following functions to custom your Kubeflow serving job.

| function  |  description  | matches cli option |
|:---|:--:|:---|
|Name(name string) *KFServingJobBuilder|specify the job name|--name|
| Namespace(namespace string) *KFServingJobBuilder|specify the namespace|--namespace/-n|
|Command(args []string) *KFServingJobBuilder|specify the command|-|
| GPUCount(count int) *KFServingJobBuilder|specify the gpu count|--gpus|
| GPUMemory(memory int) *KFServingJobBuilder |specify the gpu memory(gpushare)| --gpumemory|
| Image(image string) *KFServingJobBuilder|specify the image|--image|
| ImagePullPolicy(policy string) *KFServingJobBuilder|specify the image pull policy|--image-pull-policy|
| CPU(cpu string) *KFServingJobBuilder | specify the cpu limitation|--cpu|
|Memory(memory string) *KFServingJobBuilder |specify the memory limitation|--memory|
|Envs(envs map[string]string) *KFServingJobBuilder | specify the envs of containers| --env |
| Replicas(count int) *KFServingJobBuilder|specify the replicas| --replicas|
| EnableIstio() *KFServingJobBuilder|enable istio|--enable-istio|
| ExposeService() *KFServingJobBuilder|expose service|--expose-service|
| Version(version string) *KFServingJobBuilder| specify the version|--version|
| Tolerations(tolerations []string) *KFServingJobBuilder|specify the node taint tolerations| --toleration|
| NodeSelectors(selectors map[string]string) *KFServingJobBuilder|specify the node selectors|--selector|
|Annotations(annotations map[string]string) *KFServingJobBuilder |specify the annotation|--annotation|
|Datas(volumes map[string]string) *KFServingJobBuilder|specify the pvc which stores dataset|--data|
| DataDirs(volumes map[string]string) *KFServingJobBuilder|specify the host path which stores dataset|--data-dir|
| ModelType(modeType string) *KFServingJobBuilder |specify the model type|--model-type|
| CanaryPercent(percent int) *KFServingJobBuilder|specify the canary percent|--canary-percent|
|StorageUri(uri string) *KFServingJobBuilder|specify the storage url|--storage-uri|
| Port(port int) *KFServingJobBuilder |specify the port|--port|
|Build() (*Job, error) |build the job|-|

## Example

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
            arena serve kfserving \
            --name=max-object-detector \
            --port=5000 \
            --image=codait/max-object-detector \
            --model-type=custom 
        */
        submitJob, err := serving.NewKFServingJobBuilder().
            Name(jobName).
            Port(5000).
            Image("codait/max-object-detector").
            ModelType("custom").
            Build()
        if err != nil {
            fmt.Printf("failed to build kubeflow serving job,reason: %v\n", err)
            return
        }
        // submit kubeflow serving job
        if err := client.Serving().Submit(submitJob); err != nil {
            fmt.Printf("failed to submit job,reason: %v\n", err)
            return
        }
    }