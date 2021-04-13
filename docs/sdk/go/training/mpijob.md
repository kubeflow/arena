# Build MPI Training Job

This section introduces how to customly build a MPI training job.

## Path

    pkg/apis/training.MPIJobBuilder

## Function

    func NewMPIJobBuilder() *MPIJobBuilder

## Parameters

MPIJobBuilder has following functions to custom your MPI training job.

| function  |  description  | matches cli option |
|:---|:--:|:---|
| Name(name string) *MPIJobBuilder   | specify the job name   | --name   |
| Command(args []string) *MPIJobBuilder   |  specify the job command  | -   |
| WorkingDir(dir string) *MPIJobBuilder   | specify the working dir  |  --working-dir  |
| Envs(envs map[string]string) *MPIJobBuilder   | specify the container env   |  --env  |
| GPUCount(count int) *MPIJobBuilder   |  specify the gpu count of each worker | --gpus   |
| Image(image string) *MPIJobBuilder   |  specify the image  |  --image  |
| Tolerations(tolerations []string) *MPIJobBuilder    | specify the k8s node taint tolerations   |  --toleration  |
| ConfigFiles(files map[string]string) *MPIJobBuilder    | specify the configuration files   |  --config-file  |
| NodeSelectors(selectors map[string]string) *MPIJobBuilder   | specify the node selectors   | --selector   |
| Annotations(annotations map[string]string) *MPIJobBuilder   | specify the instance annotations   | --annotation   |
| Datas(volumes map[string]string) *MPIJobBuilder   | specify the data pvc   | --data   |
| DataDirs(volumes map[string]string) *MPIJobBuilder   |  specify host path and its'  mapping container path  |  --data-dir  |
| LogDir(dir string) *MPIJobBuilder   | specify the log dir   | --logdir   |
| Priority(priority string) *MPIJobBuilder   | specify the priority   | --priority   |
| EnableRDMA() *MPIJobBuilder    | enable rdma   |  --rdma  |
| SyncImage(image string) *MPIJobBuilder   | specify the sync image   | --sync-image   |
| SyncMode(mode string) *MPIJobBuilder | specify the sync mode(rsync,git)| --sync-mode |
| SyncSource(source string) *MPIJobBuilder   | specify the code address(eg: git url or rsync url)   | --sync-source   |
| EnableTensorboard() *MPIJobBuilder   | enable tensorboard   |  --tensorboard  |
| TensorboardImage(image string) *MPIJobBuilder   | specify the tensorboard image   |  --tensorboard-image  |
| ImagePullSecrets(secrets []string) *MPIJobBuilder   | specify the image pull secret   |  --image-pull-secret  |
| WorkerCount(count int) *MPIJobBuilder| specify the worker count | --workers|
| CPU(cpu string) *MPIJobBuilder| specify the cpu limits| --cpu |
| Memory(memory string) *MPIJobBuilder| specify the memory limits| --memory|
| EnableGPUTopology() *MPIJobBuilder| enable gpu topology scheduling| --gputopology|
|  Build() (*Job, error) | build the MPI training job | - |

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
        jobName := "mpi-standalone-test"
        jobType := types.MPITrainingJob
        // create arena client
        client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
            Kubeconfig: "",
            LogLevel:   "info",
            Namespace:  "default",
        })
        // create tfjob
        /* command:
        arena submit mpijob --name=mpi-standalone-test \
        --gpus=1 \
        --workers=1 \
        --working-dir=/perseus-demo/tensorflow-demo/ \
        --image=registry.cn-beijing.aliyuncs.com/kube-ai/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
        --tensorboard \
        "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"
        */
        job, err := training.NewMPIJobBuilder().
            Name(jobName).
            GPUCount(1).
            WorkerCount(2).
            WorkingDir("/perseus-demo/tensorflow-demo/").
            EnableTensorboard().
            Image("registry.cn-beijing.aliyuncs.com/kube-ai/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5").
            Command([]string{"mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64 --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"}).Build()
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