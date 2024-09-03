# Build Custom Serving Job

This section introduces how to customly build a Custom serving job.

## Path

    pkg/apis/serving.CustomServingJobBuilder

## Function

    func NewCustomServingJobBuilder() *CustomServingJobBuilder 

## Parameters

CustomServingJobBuilder has following functions to custom your Custom serving job.

| function                                                            |  description  | matches cli option  |
|:--------------------------------------------------------------------|:--:|:--------------------|
| Name(name string) *CustomServingJobBuilder                          |specify the job name| --name              |
| Namespace(namespace string) *CustomServingJobBuilder                |specify the namespace| --namespace/-n      |
| Command(args []string) *CustomServingJobBuilder                     |specify the command| -                   |
| GPUCount(count int) *CustomServingJobBuilder                        |specify the gpu count| --gpus              |
| GPUMemory(memory int) *CustomServingJobBuilder                      |specify the gpu memory(gpushare)| --gpumemory         |
| Image(image string) *CustomServingJobBuilder                        |specify the image| --image             |
| ImagePullPolicy(policy string) *CustomServingJobBuilder             |specify the image pull policy| --image-pull-policy |
| CPU(cpu string) *CustomServingJobBuilder                            | specify the cpu limitation| --cpu               |
| Memory(memory string) *CustomServingJobBuilder                      |specify the memory limitation| --memory            |
| Envs(envs map[string]string) *CustomServingJobBuilder               | specify the envs of containers| --env               |
| Replicas(count int) *CustomServingJobBuilder                        |specify the replicas| --replicas          |
| EnableIstio() *CustomServingJobBuilder                              |enable istio| --enable-istio      |
| ExposeService() *CustomServingJobBuilder                            |expose service| --expose-service    |
| Version(version string) *CustomServingJobBuilder                    | specify the version| --version           |
| Tolerations(tolerations []string) *CustomServingJobBuilder          |specify the node taint tolerations| --toleration        |
| NodeSelectors(selectors map[string]string) *CustomServingJobBuilder |specify the node selectors| --selector          |
| Annotations(annotations map[string]string) *CustomServingJobBuilder |specify the annotation| --annotation        |
| Datas(volumes map[string]string) *CustomServingJobBuilder           |specify the pvc which stores dataset| --data              |
| DataDirs(volumes map[string]string) *CustomServingJobBuilder        |specify the host path which stores dataset| --data-dir          |
| Devices(devices map[string]string) *CustomServingJobBuilder         |specify the chip vendors and count that used for resources, such as amd.com/gpu=1 gpu.intel.com/i915=1| --device            |
| RestfulPort(port int) *CustomServingJobBuilder                      |specify the http service port| --restful-port      |
| Port(port int) *CustomServingJobBuilder                             |specify the grpc service port| --port              |
| Build() (*Job, error)                                               |build the job| -                   |

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