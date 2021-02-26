# Submit Tensorflow Job with enabled tensorboard

## Submit the job

Here is an example how you can use ``Arena`` for the machine learning training. It will download the source code from git url, and use Tensorboard to visualize the Tensorflow computation graph and plot quantitative metrics.

1\. the first step is to check the available resources:

    $ arena top node
    NAME                       IPADDRESS      ROLE    STATUS  GPU(Total)  GPU(Allocated)
    cn-hongkong.192.168.2.107  47.242.51.160  <none>  Ready   1           0
    cn-hongkong.192.168.2.108  192.168.2.108  <none>  Ready   1           0
    cn-hongkong.192.168.2.109  192.168.2.109  <none>  Ready   1           0
    cn-hongkong.192.168.2.110  192.168.2.110  <none>  Ready   1           0
    -----------------------------------------------------------------------------------
    Allocated/Total GPUs In Cluster:
    0/4 (0.0%)

There are 3 available nodes with GPU for running training jobs.

2\. Now we can submit a training job with ``arena cli``, it will download the source code from github:

    $ arena submit tf \
        --name=tf-tensorboard \
        --gpus=1 \
        --image=tensorflow/tensorflow:1.5.0-devel-gpu \
        --env=TEST_TMPDIR=code/tensorflow-sample-code/ \
        --sync-mode=git \
        --sync-source=https://code.aliyun.com/xiaozhou/tensorflow-sample-code.git \
        --tensorboard \
        --logdir=/training_logs \
        "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 5000"

    configmap/tf-tensorboard-tfjob created
    configmap/tf-tensorboard-tfjob labeled
    service/tf-tensorboard-tensorboard created
    deployment.extensions/tf-tensorboard-tensorboard created
    tfjob.kubeflow.org/tf-tensorboard created
    INFO[0001] The Job tf-tensorboard has been submitted successfully
    INFO[0001] You can run `arena get tf-tensorboard --type tfjob` to check the job status

!!! note

    * the source code will be downloaded and extracted to the directory ``code/`` of the working directory. The default working directory is ``/root``, you can also specify by using ``--workingDir``.
    * ``logdir`` indicates where the tensorboard reads the event logs of TensorFlow


## List the tensorflow jobs 

When submited the job, you can list all tensorflow training jobs:

    $ arena list -T tfjob
    NAME                         STATUS     TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
    tf-tensorboard               PENDING    TFJOB    2m        1               1               N/A
    tf-standalone-test-with-git  SUCCEEDED  TFJOB    4m        1               N/A             192.168.2.109


## Get the tensorflow job details

1\. If you want to get the training job details,``arena get`` can help you:

    $ arena get tf-tensorboard
    Name:        tf-tensorboard
    Status:      RUNNING
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    5m

    Instances:
    NAME                    STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                    ------   ---  --------  --------------  ----
    tf-tensorboard-chief-0  Running  5m   true      1               cn-hongkong.192.168.2.108

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.2.107:31141

2\. Use ``-g`` can display the gpu utilization of the job(this feature depends on the prometheus):

    $ arena get tf-tensorboard -g
    Name:        tf-tensorboard
    Status:      RUNNING
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    6m

    Instances:
    NAME                    STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                    ------   ---  --------  --------------  ----
    tf-tensorboard-chief-0  Running  6m   true      1               cn-hongkong.192.168.2.108

    GPUs:
    INSTANCE                NODE(IP)       GPU(Requested)  GPU(IndexId)  GPU(DutyCycle)  GPU Memory(Used/Total)
    --------                --------       --------------  ------------  --------------  ----------------------
    tf-tensorboard-chief-0  192.168.2.108  1               N/A           N/A             N/A

    Allocated/Requested GPUs of Job: 1/1

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.2.107:31141


3\. Check the resource usage of the cluster:

    $ arena top node
    NAME                       IPADDRESS      ROLE    STATUS  GPU(Total)  GPU(Allocated)
    cn-hongkong.192.168.2.107  47.242.51.160  <none>  Ready   1           0
    cn-hongkong.192.168.2.108  192.168.2.108  <none>  Ready   1           1
    cn-hongkong.192.168.2.109  192.168.2.109  <none>  Ready   1           0
    cn-hongkong.192.168.2.110  192.168.2.110  <none>  Ready   1           0
    -----------------------------------------------------------------------------------
    Allocated/Total GPUs In Cluster:
    1/4 (25.0%)

!!! note

    * you can access the tensorboard by using ``192.168.1.117:30670``. You can consider ``sshuttle`` if you can't access the tensorboard directly from your laptop. For example: ``sshuttle -r root@47.89.59.51 192.168.0.0/16``


![tensorboard](./2-tensorboard.jpg)

Congratulations! You've run the training job with ``arena`` successfully, and you can also check the tensorboard easily.
