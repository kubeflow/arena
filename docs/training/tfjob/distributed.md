# Submit a distributed tensorflow job

Arena supports and simplifies distributed TensorFlow Training (PS/worker mode). 

!!! warning
    
    To run a distributed Tensorflow Training, you need to specify:

    - GPUs of each worker (only for GPU workload)
    - The number of workers (required)
    - The number of PS (required)
    - The docker image of worker (required)
    - The docker image of PS (required)
    - The Port of Worker (default is 22222)
    - The Port of PS (default is 22223)

The following command is an example. In this example, it defines 2 workers and 1 PS, and each worker has 1 GPU. The source code of worker and PS are located in git, and the tensorboard are enabled.

1\. Submit the tensorflow training job.

    $ arena submit tf \
        --name=tf-dist-git \
        --gpus=1 \
        --workers=2 \
        --worker-image=tensorflow/tensorflow:1.5.0-devel-gpu \
        --sync-mode=git \
        --sync-source=https://code.aliyun.com/xiaozhou/tensorflow-sample-code.git \
        --ps=1 \
        --ps-image=tensorflow/tensorflow:1.5.0-devel \
        --tensorboard \
        "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --log_dir=/training_logs --data_dir=code/tensorflow-sample-code/data"

    configmap/tf-dist-git-tfjob created
    configmap/tf-dist-git-tfjob labeled
    service/tf-dist-git-tensorboard created
    deployment.extensions/tf-dist-git-tensorboard created
    tfjob.kubeflow.org/tf-dist-git created
    INFO[0001] The Job tf-dist-git has been submitted successfully
    INFO[0001] You can run `arena get tf-dist-git --type tfjob` to check the job status

!!! note

    If you saw the job or pod is failed, and then look at the logs, you may find out it is due to the reason that git code is not be able to cloned, especially if you are runing container insider some countries like China. This is not caused by arena, but cross-border network connectivity. 

2\. Get the details of the specific job.

    $ arena get tf-dist-git

    Name:        tf-dist-git
    Status:      SUCCEEDED
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    6m

    Instances:
    NAME                  STATUS       AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                  ------       ---  --------  --------------  ----
    tf-dist-git-ps-0      Terminating  6m   false     0               cn-hongkong.192.168.2.109
    tf-dist-git-worker-0  Completed    6m   true      1               cn-hongkong.192.168.2.108
    tf-dist-git-worker-1  Terminating  6m   false     1               cn-hongkong.192.168.2.110

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.2.107:30600

3\. Check the tensorboard.

![image](3-tensorboard.jpg)

4\. Get the TFJob dashboard.

    $ arena logviewer tf-dist-git

    Your LogViewer will be available on:
    172.20.0.197:8080/tfjobs/ui/#/default/tf-dist-git


![image](4-tfjob-logviewer-distributed.jpg)

Congratulations! You've run the distributed training job with ``arena`` successfully. 