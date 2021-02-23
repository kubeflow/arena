
这个示例展示了如何使用 `Arena` 提交一个 pytorch 分布式的作业，并通过 tensorboard 可视化。该示例将从 git url 下载源代码。

1. 第一步是检查可用的GPU资源
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
    有 3 个包含 GPU 的可用节点用于运行训练作业

2. 提交一个 pytorch 2 机 1 卡的训练作业，本示例从 [阿里云 code](https://code.aliyun.com/370272561/mnist-pytorch.git) 下载源代码
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

    > 这会下载源代码，并将其解压缩到工作目录的 `code/` 目录。默认的工作目录是 `/root`，您也可以使用 `--workingDir` 加以指定。
    
    > workers 为参与计算的节点总数（必须为正数且大于等于 1），包括用于建立通信的 rank0 节点（对应 pytorch operator 中的 master 节点），默认为 1，可以不填写，即为单机作业。
    
    > `logdir` 指示 Tensorboard 在何处读取 PyTorch 的事件日志

3. 列出所有作业
    ```
    ➜ arena list
    NAME                          STATUS     TRAINER     AGE  NODE
    pytorch-dist-tensorboard      SUCCEEDED  PYTORCHJOB  22h  N/A
    ```

4. 获取作业的详细信息
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
    > 注意：您可以使用 `172.16.0.205:30583` 访问 Tensorboard。如果您通过笔记本电脑无法直接访问 Tensorboard，则可以考虑使用 `sshuttle`。例如：
    
    ```
    # you can install sshuttle==0.74 in your mac with python2.7
    ➜ pip install sshuttle==0.74
    # 0/0 -> 0.0.0.0/0
    ➜ sshuttle -r root@39.104.17.205  0/0
    ```
    ![](19-pytorchjob-tensorboard.png)
