# Build Elastic Training Job

This section introduces how to customly build a elastic training job.

## Path

    pkg/apis/training.ETJobBuilder

## Function

    func NewETJobBuilder() *ETJobBuilder

## Parameters

ETJobBuilder has following functions to custom your elastic training job.

| function                                                 |  description  | matches cli option  |
|:---------------------------------------------------------|:--:|:--------------------|
| Name(name string) *ETJobBuilder                          | specify the job name   | --name              |
| Command(args []string) *ETJobBuilder                     |  specify the job command  | -                   |
| WorkingDir(dir string) *ETJobBuilder                     | specify the working dir  | --working-dir       |
| Envs(envs map[string]string) *ETJobBuilder               | specify the container env   | --env               |
| GPUCount(count int) *ETJobBuilder                        |  specify the gpu count of each worker | --gpus              |
| Image(image string) *ETJobBuilder                        |  specify the image  | --image             |
| Tolerations(tolerations []string) *ETJobBuilder          | specify the k8s node taint tolerations   | --toleration        |
| ConfigFiles(files map[string]string) *ETJobBuilder       | specify the configuration files   | --config-file       |
| NodeSelectors(selectors map[string]string) *ETJobBuilder | specify the node selectors   | --selector          |
| Annotations(annotations map[string]string) *ETJobBuilder | specify the instance annotations   | --annotation        |
| Datas(volumes map[string]string) *ETJobBuilder           | specify the data pvc   | --data              |
| DataDirs(volumes map[string]string) *ETJobBuilder        |  specify host path and its'  mapping container path  | --data-dir          |
| Devices(devices map[string]string) *ETJobBuilder         |  specify the chip vendors and count that used for resources, such as amd.com/gpu=1 gpu.intel.com/i915=1  | --device            |
| LogDir(dir string) *ETJobBuilder                         | specify the log dir   | --logdir            |
| Priority(priority string) *ETJobBuilder                  | specify the priority   | --priority          |
| EnableRDMA() *ETJobBuilder                               | enable rdma   | --rdma              |
| SyncImage(image string) *ETJobBuilder                    | specify the sync image   | --sync-image        |
| SyncMode(mode string) *ETJobBuilder                      | specify the sync mode(rsync,git)| --sync-mode         |
| SyncSource(source string) *ETJobBuilder                  | specify the code address(eg: git url or rsync url)   | --sync-source       |
| EnableTensorboard() *ETJobBuilder                        | enable tensorboard   | --tensorboard       |
| TensorboardImage(image string) *ETJobBuilder             | specify the tensorboard image   | --tensorboard-image |
| ImagePullSecrets(secrets []string) *ETJobBuilder         | specify the image pull secret   | --image-pull-secret |
| WorkerCount(count int) *ETJobBuilder                     | specify the worker count | --workers           |
| CPU(cpu string) *ETJobBuilder                            | specify the cpu limits| --cpu               |
| Memory(memory string) *ETJobBuilder                      | specify the memory limits| --memory            |
| MaxWorkers(count int) *ETJobBuilder                      | specify the max workers| --max-workers       |
| MinWorkers(count int) *ETJobBuilder                      | specify the min workers| --min-workers       |
| Build() (*Job, error)                                    | build the elastic training job | -                   |

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
    }