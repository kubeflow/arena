# Build TF Training Job

This section introduces how to customly build a TF training job.

## Path

    pkg/apis/training.TFJobBuilder

## Function

    func NewTFJobBuilder() *TFJobBuilder

## Parameters

TFJobBuilder has following functions to custom your TF training job.

| function                                                      |  description  | matches cli option   |
|:--------------------------------------------------------------|:--:|:---------------------|
| Name(name string) *TFJobBuilder                               | specify the job name   | --name               |
| Command(args []string) *TFJobBuilder                          |  specify the job command  | -                    |
| WorkingDir(dir string) *TFJobBuilder                          | specify the working dir  | --working-dir        |
| Envs(envs map[string]string) *TFJobBuilder                    | specify the container env   | --env                |
| GPUCount(count int) *TFJobBuilder                             |  specify the gpu count of each worker | --gpus               |
| Image(image string) *TFJobBuilder                             |  specify the image  | --image              |
| Tolerations(tolerations []string) *TFJobBuilder               | specify the k8s node taint tolerations   | --toleration         |
| ConfigFiles(files map[string]string) *TFJobBuilder            | specify the configuration files   | --config-file        |
| NodeSelectors(selectors map[string]string) *TFJobBuilder      | specify the node selectors   | --selector           |
| Annotations(annotations map[string]string) *TFJobBuilder      | specify the instance annotations   | --annotation         |
| EnableChief() *TFJobBuilder                                   | enable chief role| --chief              |
| ChiefCPU(cpu string) *TFJobBuilder                            | specify the chief role cpu limits| --chief-cpu          |
| ChiefMemory(mem string) *TFJobBuilder                         |specify the chief role memory limits| --chief-memory       |
| ChiefPort(port int) *TFJobBuilder                             | specify the chief role port| --chief-port         |
| ChiefSelectors(selectors map[string]string) *TFJobBuilder     | specify the node selectors of chief role| --chief-selector     |
| Datas(volumes map[string]string) *TFJobBuilder                | specify the data pvc   | --data               |
| DataDirs(volumes map[string]string) *TFJobBuilder             |  specify host path and its'  mapping container path  | --data-dir           |
| Devices(devices map[string]string) *TFJobBuilder              |  specify the chip vendors and count that used for resources, such as amd.com/gpu=1 gpu.intel.com/i915=1  | --device             |
| EnableEvaluator() *TFJobBuilder                               |enable evaluator role| --evaluator          |
| EvaluatorCPU(cpu string) *TFJobBuilder                        | specify the cpu limits of evaluator role| --evaluator-cpu      |
| EvaluatorMemory(mem string) *TFJobBuilder                     | specify the memory limits of evaluator role| --evaluator-memory   |
| EvaluatorSelectors(selectors map[string]string) *TFJobBuilder | specify the node selectors of evaluator role| --evaluator-selector |
| LogDir(dir string) *TFJobBuilder                              | specify the log dir   | --logdir             |
| Priority(priority string) *TFJobBuilder                       | specify the priority   | --priority           |
| PsCount(count int) *TFJobBuilder                              | specify the count of ps role| --ps                 |
| PsCPU(cpu string) *TFJobBuilder                               |specify the cpu limits of ps role| --ps-cpu             |
| PsMemory(mem string) *TFJobBuilder                            |specify the memory limits of ps memory| --ps-memory          |
| PsPort(port int) *TFJobBuilder                                |specify the port of ps role| --ps-port            |
| PsSelectors(selectors map[string]string) *TFJobBuilder        | specify the node selectors of ps role| --ps-selector        |
| PsImage(image string) *TFJobBuilder                           |specify the image of ps role| --ps-image           |
| EnableRDMA() *TFJobBuilder                                    | enable rdma   | --rdma               |
| SyncImage(image string) *TFJobBuilder                         | specify the sync image   | --sync-image         |
| SyncMode(mode string) *TFJobBuilder                           | specify the sync mode(rsync,git)| --sync-mode          |
| SyncSource(source string) *TFJobBuilder                       | specify the code address(eg: git url or rsync url)   | --sync-source        |
| EnableTensorboard() *TFJobBuilder                             | enable tensorboard   | --tensorboard        |
| TensorboardImage(image string) *TFJobBuilder                  | specify the tensorboard image   | --tensorboard-image  |
| WorkerCPU(cpu string) *TFJobBuilder                           | specify the cpu limits of worker role| --worker-cpu         |
| WorkerImage(image string) *TFJobBuilder                       | specify the worker image | --worker-image       |
| WorkerMemory(mem string) *TFJobBuilder                        | specify the worker memory limits| --worker-memory      |
| WorkerPort(port int) *TFJobBuilder                            | specify the worker port| --worker-porter      |
| WorkerSelectors(selectors map[string]string) *TFJobBuilder    | specify the node selectors of worker role| --worker-selector    |
|                                                               ||                      |
| ImagePullSecrets(secrets []string) *TFJobBuilder              | specify the image pull secret   | --image-pull-secret  |
| WorkerCount(count int) *TFJobBuilder                          | specify the worker count | --workers            |
| CleanPodPolicy(policy string) *TFJobBuilder                   | specify the policy of cleaning pod| --clean-task-policy  |
| RoleSequence(roles []string) *TFJobBuilder                    | specify the role sequence| --role-sequence      |
| Build() (*Job, error)                                         | build the TF training job | -                    |

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
        jobName := "test-dist-tfjob"
        jobType := types.TFTrainingJob
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
                arena submit tfjob \
                --name=tf-distributed-test \
                --gpus=1 \
                --workers=1 \
                --worker-image=cheyang/tf-mnist-distributed:gpu \
                --ps-image=cheyang/tf-mnist-distributed:cpu \
                --ps=1 \
                --tensorboard \
                "python /app/main.py"
        */
        submitJob, err := training.NewTFJobBuilder().
            Name(jobName).
            GPUCount(1).
            WorkerCount(1).
            WorkerImage("cheyang/tf-mnist-distributed:gpu").
            PsImage("cheyang/tf-mnist-distributed:cpu").
            PsCount(1).
            EnableTensorboard().
            Command([]string{"'python /app/main.py'"}).Build()
        if err != nil {
            fmt.Printf("failed to build tfjob,reason: %v\n", err)
            return
        }
        // submit tfjob
        if err := client.Training().Submit(submitJob); err != nil {
            fmt.Printf("failed to submit job,reason: %v\n", err)
            return
        }
    }