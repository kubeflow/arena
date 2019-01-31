

Arena supports and simplifies distributed TensorFlow Training(ps/worker mode). 


1. To run a distributed Tensorflow Training, you need to specify:

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
              --tensorboard \
              "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --logdir /training_logs"
configmap/tf-dist-git-tfjob created
configmap/tf-dist-git-tfjob labeled
service/tf-dist-git-tensorboard created
deployment.extensions/tf-dist-git-tensorboard created
tfjob.kubeflow.org/tf-dist-git created
INFO[0001] The Job tf-dist-git has been submitted successfully
INFO[0001] You can run `arena get tf-dist-git --type tfjob` to check the job status
```

2\. Get the details of the specific job

```
# arena get tf-dist-git
NAME         STATUS   TRAINER  AGE  INSTANCE                            NODE                   
tf-dist-git  RUNNING  tfjob    55s  tf-dist-git-tfjob-594d59789c-lrfsk  192.168.1.119
tf-dist-git  RUNNING  tfjob    55s  tf-dist-git-tfjob-ps-0              192.168.1.118
tf-dist-git  RUNNING  tfjob    55s  tf-dist-git-tfjob-worker-0          192.168.1.119
tf-dist-git  RUNNING  tfjob    55s  tf-dist-git-tfjob-worker-1          192.168.1.120

Your tensorboard will be available on:
192.168.1.117:32298
```

3\. Check the tensorboard

![](3-tensorboard.jpg)


4\. Get the TFJob dashboard

```
# arena logviewer tf-dist-git
Your LogViewer will be available on:
192.168.1.120:8080/tfjobs/ui/#/default/tf-dist-git-tfjob
```


![](4-tfjob-logviewer-distributed.jpg)

Congratulations! You've run the distributed training job with `arena` successfully. 