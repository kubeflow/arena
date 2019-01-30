
Arena supports distributed TensorFlow Training with gang scheduling by using [kube-arbitrator](https://github.com/kubernetes-incubator/kube-arbitrator). 

When running distributed TensorFlow, we'd better to make sure `all` or `nothing`. Gang scheduling can help such case. 


> Notice: the current [kubernetes gang scheduler](https://github.com/kubernetes-incubator/kube-arbitrator/tree/release-0.1) is not production ready. For example, it doesn't support Pod Affinity and PodFitsHostPorts for sheduling. 

> Limitation: when using gang scheduler, the tensorboard feature doesn't work well.

1. To enable gang scheduler, edit `/charts/tfjob/values.yaml`

Change `enableGangScheduler: false` to `enableGangScheduler: true`

2. To run a distributed Tensorflow Training, you need to specify:

 - GPUs of each worker (only for GPU workload)
 - The number of workers (required)
 - The number of PS (required)
 - The docker image of worker (required)
 - The docker image of PS (required)
 - The Port of Worker (default is 22222)
 - The Port of PS (default is 22223)

The following command is an example. In this example, it defines 2 workers and 1 PS, and each worker has 1 GPU. The source code of worker and PS are located in git, and the tensorboard are enabled.

```
# arena submit tf --name=tf-dist-git              \
              --gpus=1              \
              --workers=2              \
              --workerImage=tensorflow/tensorflow:1.5.0-devel-gpu  \
              --syncMode=git \
              --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
              --ps=1              \
              --psImage=tensorflow/tensorflow:1.5.0-devel   \
              "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --logdir /training_logs"
configmap/tf-dist-git-tfjob created
configmap/tf-dist-git-tfjob labeled
service/tf-dist-git-tensorboard created
deployment.extensions/tf-dist-git-tensorboard created
tfjob.kubeflow.org/tf-dist-git created
INFO[0001] The Job tf-dist-git has been submitted successfully
INFO[0001] You can run `arena get tf-dist-git --type tfjob` to check the job status

```

If there are no enough resources, all the instances of the job are `PENDING`. If it's not gang scheduler, you can see some of the pods are `RUNNING` and others are `PENDING`.

```
# arena get tf-dist-data
NAME          STATUS   TRAINER  AGE  INSTANCE                     NODE
tf-dist-data  PENDING  TFJOB    0s   tf-dist-data-tfjob-ps-0      N/A
tf-dist-data  PENDING  TFJOB    0s   tf-dist-data-tfjob-worker-0  N/A
tf-dist-data  PENDING  TFJOB    0s   tf-dist-data-tfjob-worker-1  N/A
tf-dist-data  PENDING  TFJOB    0s   tf-dist-data-tfjob-worker-2  N/A
tf-dist-data  PENDING  TFJOB    0s   tf-dist-data-tfjob-worker-3  N/A
```

When there are enough resources, the the instances become `RUNNING`

```
NAME          STATUS   TRAINER  AGE  INSTANCE                     NODE
tf-dist-data  RUNNING  TFJOB    4s   tf-dist-data-tfjob-ps-0      192.168.1.115
tf-dist-data  RUNNING  TFJOB    4s   tf-dist-data-tfjob-worker-0  192.168.1.119
tf-dist-data  RUNNING  TFJOB    4s   tf-dist-data-tfjob-worker-1  192.168.1.118
tf-dist-data  RUNNING  TFJOB    4s   tf-dist-data-tfjob-worker-2  192.168.1.120
```