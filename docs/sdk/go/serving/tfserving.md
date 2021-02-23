# Build Tensorflow Serving Job

This section introduces how to customly build a Tensorflow serving job.

## Path

    pkg/apis/serving.TFServingJobBuilder

## Function

    func NewTFServingJobBuilder() *TFServingJobBuilder 

## Parameters

TFServingJobBuilder has following functions to custom your Tensorflow serving job.

| function  |  description  | matches cli option |
|:---|:--:|:---|
| Name(name string) *TFServingJobBuilder| specify the job name | --name|
| Namespace(namespace string) *TFServingJobBuilder| specify the namespace|-n/--namespace|
| Command(args []string) *TFServingJobBuilder|specify the command|-|
| GPUCount(count int) *TFServingJobBuilder| specify the gpu count| --gpus|
| GPUMemory(memory int) *TFServingJobBuilder |specify the gpu memory(gpushare)|--gpumemory|
| Image(image string) *TFServingJobBuilder|specify the image|--image|
| ImagePullPolicy(policy string) *TFServingJobBuilder|specify the image pull policy| --image-pull-policy|
| CPU(cpu string) *TFServingJobBuilder |specify the cpu limitation of job| --cpu|
|Memory(memory string) *TFServingJobBuilder |specify the memory limitation|--memory|
|Envs(envs map[string]string) *TFServingJobBuilder|specify the envs of container|--env|
| Replicas(count int) *TFServingJobBuilder|specify the replicas|--replicas|
|Port(port int) *TFServingJobBuilder |specify the grpc service port| --port|
| RestfulPort(port int) *TFServingJobBuilder|specify the restful service port|--restfulPort|
| EnableIstio() *TFServingJobBuilder|enable istio| --enable-istio|
|ExposeService() *TFServingJobBuilder | expose service| --expose-service|
| Version(version string) *TFServingJobBuilder|specify the job version| --version|
| Tolerations(tolerations []string) *TFServingJobBuilder|specify the tolerations of node taints|--toleration|
| NodeSelectors(selectors map[string]string) *TFServingJobBuilder|specify the node selectors| --selector|
| Annotations(annotations map[string]string) *TFServingJobBuilder | specify the annotations|--annotation|
| Datas(volumes map[string]string) *TFServingJobBuilder|specify the pvc name which stores dataset|--data|
| DataDirs(volumes map[string]string) *TFServingJobBuilder|specify the host path which stores dataset|--data-dir|
| VersionPolicy(policy string) *TFServingJobBuilder|specify the version policy|--version-policy|
| ModelConfigFile(filePath string) *TFServingJobBuilder|specify the model configuration file|--modeConfigFile|
| ModelName(name string) *TFServingJobBuilder |specify the model name| --model-name|
|ModelPath(path string) *TFServingJobBuilder|specify the model path|--model-path|
| Build() (*Job, error) |build the job|-|


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
        jobName := "test"
        jobType := types.TFServingJob
        // create arena client
        client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
            Kubeconfig: "",
            LogLevel:   "info",
            Namespace:  "default",
        })
        // create tfjob
        /* command:
        arena serve tensorflow \
        --name=mymnist1 \
        --model-name=mnist1  \
        --gpus=1   \
        --image=tensorflow/serving:latest-gpu \
        --data=tfmodel:/tfmodel \
        --model-path=/tfmodel/mnist \
        --versionPolicy=specific:1
        */
        job, err := serving.NewTFServingJobBuilder().
            Name(jobName).
            GPUCount(1).
            ModelName("mnist1").
            Image("tensorflow/serving:latest-gpu").
            Data(map[string]string{"tfmodel": "/tfmodel"}).
            ModelPath("/tfmodel/mnist").
            VersionPolicy("specific:1").Build()
        if err != nil {
            fmt.Printf("failed to build tensorflow serving job,reason: %v\n", err)
            return
        }
        // submit tfjob
        if err := client.Serving().Submit(job); err != nil {
            fmt.Printf("failed to submit job,reason: %v\n", err)
            return
        }
    }