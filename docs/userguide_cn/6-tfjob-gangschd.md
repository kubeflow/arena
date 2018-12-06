
Arena 支持利用 [kube-arbitrator](https://github.com/kubernetes-incubator/kube-arbitrator)，通过群调度 (gang scheduling) 算法执行分布式 TensorFlow 训练。 

运行分布式 TensorFlow 时，我们最好确保使用 `all` 或 `nothing`。在这种情况下，群调度可以提供帮助。 


> 注意：当前的 [kubernetes gang scheduler](https://github.com/kubernetes-incubator/kube-arbitrator/tree/release-0.1) 尚未准备好投入生产应用。例如，它在调度中不支持 Pod 亲和度和 PodFitsHostPorts。 

> 限制：使用群调度器时，Tensorboard 存在一定的问题。

1.为启用群调度器，首先要编辑 `/charts/tfjob/values.yaml` 文件

将 `enableGangScheduler: false` 更改为 `enableGangScheduler: true`

2.要运行分布式 Tensorflow 训练，您需要指定以下信息：

 - 各 Worker 的 GPU（仅 GPU 工作负载需要）
 - Worker 的数量（必需）
 - PS 的数量（必需）
 - Worker 的 docker 映像（必需）
 - PS 的 docker 映像（必需）
 - Worker 的端口（默认为 22222）
 - PS 的端口（默认为 22223）

如下命令提供了一个示例。本例中定义了 2 个 Worker 和 1 个 PS，每个 Worker 有 1 个 GPU。Worker 和 PS 的源代码位于 git 中，Tensorboard 已启用。

```
arena submit tf --name=tf-dist-git \
              --gpus=1 \
              --workers=2 \
              --workerImage=tensorflow/tensorflow:1.5.0-devel-gpu \
              --syncMode=git \
              --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
              --ps=1 \
              --psImage=tensorflow/tensorflow:1.5.0-devel \
              "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --logdir /training_logs"
NAME:   tf-dist-git
LAST DEPLOYED: Mon Jul 23 21:32:20 2018
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1beta1/Deployment
NAME DESIRED CURRENT UP-TO-DATE AVAILABLE AGE
tf-dist-git-tfjob 1 0 0 0 0s

==> v1alpha2/TFJob
NAME AGE
tf-dist-git-tfjob 0s

```

如果没有足够的资源，所有作业实例均处于 'PENDING' 状态。如果不是群调度器，您可以看到部分 pod 处于 `RUNNING` 状态，其他 pod 处于 `PENDING` 状态。

```
# arena get tf-dist-data
NAME STATUS TRAINER AGE INSTANCE NODE
tf-dist-data PENDING TFJOB 0s tf-dist-data-tfjob-ps-0 N/A
tf-dist-data PENDING TFJOB 0s tf-dist-data-tfjob-worker-0 N/A
tf-dist-data PENDING TFJOB 0s tf-dist-data-tfjob-worker-1 N/A
tf-dist-data PENDING TFJOB 0s tf-dist-data-tfjob-worker-2 N/A
tf-dist-data PENDING TFJOB 0s tf-dist-data-tfjob-worker-3 N/A
```

在有充足的资源时，实例状态会变为 `RUNNING`

```
NAME STATUS TRAINER AGE INSTANCE NODE
tf-dist-data RUNNING TFJOB 4s tf-dist-data-tfjob-ps-0 192.168.1.115
tf-dist-data RUNNING TFJOB 4s tf-dist-data-tfjob-worker-0 192.168.1.119
tf-dist-data RUNNING TFJOB 4s tf-dist-data-tfjob-worker-1 192.168.1.118
tf-dist-data RUNNING TFJOB 4s tf-dist-data-tfjob-worker-2 192.168.1.120
```
