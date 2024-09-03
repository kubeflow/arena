# Build Pytorch Training Job

This section introduces how to customly build a Pytorch training job.

## Path

    pkg/apis/training.PytorchJobBuilder

## Function

    func NewPytorchJobBuilder() *PytorchJobBuilder

## Parameters

PytorchJobBuilder has following functions to custom your Pytorch training job.

| function                                                      |  description  | matches cli option  |
|:--------------------------------------------------------------|:--:|:--------------------|
| Name(name string) *PytorchJobBuilder                          | specify the job name   | --name              |
| Command(args []string) *PytorchJobBuilder                     |  specify the job command  | -                   |
| WorkingDir(dir string) *PytorchJobBuilder                     | specify the working dir  | --working-dir       |
| Envs(envs map[string]string) *PytorchJobBuilder               | specify the container env   | --env               |
| GPUCount(count int) *PytorchJobBuilder                        |  specify the gpu count of each worker | --gpus              |
| Image(image string) *PytorchJobBuilder                        |  specify the image  | --image             |
| Tolerations(tolerations []string) *PytorchJobBuilder          | specify the k8s node taint tolerations   | --toleration        |
| ConfigFiles(files map[string]string) *PytorchJobBuilder       | specify the configuration files   | --config-file       |
| NodeSelectors(selectors map[string]string) *PytorchJobBuilder | specify the node selectors   | --selector          |
| Annotations(annotations map[string]string) *PytorchJobBuilder | specify the instance annotations   | --annotation        |
| Datas(volumes map[string]string) *PytorchJobBuilder           | specify the data pvc   | --data              |
| DataDirs(volumes map[string]string) *PytorchJobBuilder        |  specify host path and its'  mapping container path  | --data-dir          |
| Devices(devices map[string]string) *PytorchJobBuilder         |  specify the chip vendors and count that used for resources, such as amd.com/gpu=1 gpu.intel.com/i915=1  | --devices           |
| LogDir(dir string) *PytorchJobBuilder                         | specify the log dir   | --logdir            |
| Priority(priority string) *PytorchJobBuilder                  | specify the priority   | --priority          |
| EnableRDMA() *PytorchJobBuilder                               | enable rdma   | --rdma              |
| SyncImage(image string) *PytorchJobBuilder                    | specify the sync image   | --sync-image        |
| SyncMode(mode string) *PytorchJobBuilder                      | specify the sync mode(rsync,git)| --sync-mode         |
| SyncSource(source string) *PytorchJobBuilder                  | specify the code address(eg: git url or rsync url)   | --sync-source       |
| EnableTensorboard() *PytorchJobBuilder                        | enable tensorboard   | --tensorboard       |
| TensorboardImage(image string) *PytorchJobBuilder             | specify the tensorboard image   | --tensorboard-image |
| ImagePullSecrets(secrets []string) *PytorchJobBuilder         | specify the image pull secret   | --image-pull-secret |
| WorkerCount(count int) *PytorchJobBuilder                     | specify the worker count | --workers           |
| CPU(cpu string) *PytorchJobBuilder                            | specify the cpu limits| --cpu               |
| Memory(memory string) *PytorchJobBuilder                      | specify the memory limits| --memory            |
| CleanPodPolicy(policy string) *PytorchJobBuilder              | specify the cleaning pod policy| --clean-task-policy |
| Build() (*Job, error)                                         | build the Pytorch training job | -                   |

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
    }