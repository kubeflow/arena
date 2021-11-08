# Tensorflow job with gang scheduling enabled

Arena supports distributed TensorFlow Training with gang scheduling by using [scheduler-plugins](https://github.com/kubernetes-sigs/scheduler-plugins/tree/master/pkg/coscheduling). We can enable gang scheduling by adding parameter `--gang`.  Learn more https://help.aliyun.com/document_detail/178169.html

When running distributed TensorFlow, we'd better to make sure ``all`` or ``nothing``. Gang scheduling can help such case.


!!! warning

    Limitation: when using gang scheduling, the tensorboard feature doesn't work well.

The following command is an example. In this example, it defines 2 workers and 1 PS, and each worker has 1 GPU. The source code of worker and PS are located in git, and the tensorboard are enabled.

```
$ arena submit tf \
    --name=tf-dist-git \
    --gpus=1 \
    --workers=2 \
    --worker-image=tensorflow/tensorflow:1.5.0-devel-gpu \
    --sync-mode=git \
    --sync-source=https://code.aliyun.com/xiaozhou/tensorflow-sample-code.git \
    --ps=1 \
    --ps-image=tensorflow/tensorflow:1.5.0-devel \
    --gang \
    --tensorboard \
    "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --log_dir=/training_logs --data_dir=code/tensorflow-sample-code/data"

service/tf-dist-git-tensorboard created
deployment.apps/tf-dist-git-tensorboard created
tfjob.kubeflow.org/tf-dist-git created
INFO[0002] The Job tf-dist-git has been submitted successfully
INFO[0002] You can run `arena get tf-dist-git --type tfjob` to check the job status
```

If there are no enough resources, all the instances of the job are ``PENDING``. If it's not gang scheduling, you can see some pods are ``RUNNING`` and others are ``PENDING``.

```bash
$ arena get tf-dist-git
Name:      tf-dist-git
Status:    PENDING
Namespace: default
Priority:  N/A
Trainer:   TFJOB
Duration:  4s

Instances:
  NAME                  STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
  ----                  ------   ---  --------  --------------  ----
  tf-dist-git-ps-0      Pending  4s   false     0               N/A
  tf-dist-git-worker-0  Pending  4s   true      1               N/A
  tf-dist-git-worker-1  Pending  4s   false     1               N/A

Tensorboard:
  Your tensorboard will be available on:
  http://10.0.0.80:31029
```

When there are enough resources, the instances become ``RUNNING``.

```bash
$ arena get tf-dist-git
Name:      tf-dist-git
Status:    RUNNING
Namespace: default
Priority:  N/A
Trainer:   TFJOB
Duration:  50s

Instances:
  NAME                  STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
  ----                  ------   ---  --------  --------------  ----
  tf-dist-git-ps-0      Running  0s   false     0               cn-beijing.10.0.0.84
  tf-dist-git-worker-0  Running  50s  true      1               cn-beijing.10.0.0.83
  tf-dist-git-worker-1  Running  50s  false     1               cn-beijing.10.0.0.85

Tensorboard:
  Your tensorboard will be available on:
  http://10.0.0.80:31029
```
