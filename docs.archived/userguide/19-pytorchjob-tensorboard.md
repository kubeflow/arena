This example shows how to use `Arena` to submit a python distributed job and visualize by `Tensorboard`. The sample downloads the source code from git URL.

1. The first step is to check the available resources.
    ```
    ➜ arena top node
    NAME                       IPADDRESS     ROLE    STATUS  GPU(Total)  GPU(Allocated)
    cn-huhehaote.172.16.0.205  172.16.0.205  master  ready   0           0
    cn-huhehaote.172.16.0.206  172.16.0.206  master  ready   0           0
    cn-huhehaote.172.16.0.207  172.16.0.207  master  ready   0           0
    cn-huhehaote.172.16.0.208  172.16.0.208  <none>  ready   4           0
    cn-huhehaote.172.16.0.209  172.16.0.209  <none>  ready   4           0
    cn-huhehaote.172.16.0.210  172.16.0.210  <none>  ready   4           0
    -----------------------------------------------------------------------------------------
    Allocated/Total GPUs In Cluster:
    0/12 (0%)
    ```
    There are 3 available nodes with GPU for running training jobs.

2. Submit a pytorch distributed training job with 2 nodes and one gpu card, this example download the source code from [Alibaba Cloud code](https://code.aliyun.com/370272561/mnist-pytorch.git).
    ```
    ➜ arena --loglevel info submit pytorch \
            --name=pytorch-dist-tensorboard \
            --gpus=1 \
            --workers=2 \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            --tensorboard \
            --logdir=/root/logs \
            "python /root/code/mnist-pytorch/mnist.py --epochs 50 --backend gloo --dir /root/logs"
    configmap/pytorch-dist-tensorboard-pytorchjob created
    configmap/pytorch-dist-tensorboard-pytorchjob labeled
    service/pytorch-dist-tensorboard-tensorboard created
    deployment.apps/pytorch-dist-tensorboard-tensorboard created
    pytorchjob.kubeflow.org/pytorch-dist-tensorboard created
    INFO[0000] The Job pytorch-dist-tensorboard has been submitted successfully
    INFO[0000] You can run `arena get pytorch-dist-tensorboard --type pytorchjob` to check the job status
    ```

    > the source code will be downloaded and extracted to the directory `code/` of the working directory. The default working directory is `/root`, you can also specify by using `--workingDir`.

    > `workers` is the total number of nodes participating in the training (must be a positive integer and greater than or equal to 1), including rank0 node used to establish communication (corresponding to the `master` node in the pytorch-operator). The default value of the parameter is 1, which can not be set, as a stand-alone job.

    > `logdir` indicates where the tensorboard reads the event logs of Pytorch.

3. List all the jobs.
    ```
    ➜ arena list
    NAME                          STATUS     TRAINER     AGE  NODE
    pytorch-dist-tensorboard      SUCCEEDED  PYTORCHJOB  22h  N/A
    ```

4. Get the details of the this job.
    ```
    ➜ arena get pytorch-dist-tensorboard
    STATUS: SUCCEEDED
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 15m
    
    NAME                      STATUS     TRAINER     AGE  INSTANCE                           NODE
    pytorch-dist-tensorboard  SUCCEEDED  PYTORCHJOB  22h  pytorch-dist-tensorboard-master-0  172.16.0.210
    pytorch-dist-tensorboard  SUCCEEDED  PYTORCHJOB  22h  pytorch-dist-tensorboard-worker-0  172.16.0.210
    
    Your tensorboard will be available on:
    http://172.16.0.205:30583
    ```
    > Notice: you can access the tensorboard by using `172.16.0.205:30583`. You can consider `sshuttle` if you can't access the tensorboard directly from your laptop. For example: 
    ```
    # you can install sshuttle==0.74 in your mac with python2.7
    ➜ pip install sshuttle==0.74
    # 0/0 -> 0.0.0.0/0
    ➜ sshuttle -r root@39.104.17.205  0/0
    ```
    ![](19-pytorchjob-tensorboard.png)
