
这个示例展示了如何使用 `Arena` 提交一个 pytorch 分布式的作业。该示例将从 git url 下载源代码。

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
            --name=pytorch-dist-git \
            --gpus=1 \
            --workers=2 \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            "python /root/code/mnist-pytorch/mnist.py --backend gloo"
    configmap/pytorch-dist-git-pytorchjob created
    configmap/pytorch-dist-git-pytorchjob labeled
    pytorchjob.kubeflow.org/pytorch-dist-git created
    INFO[0000] The Job pytorch-dist-git has been submitted successfully
    INFO[0000] You can run `arena get pytorch-dist-git --type pytorchjob` to check the job status
    ```

    > 这会下载源代码，并将其解压缩到工作目录的 `code/` 目录。默认的工作目录是 `/root`，您也可以使用 `--workingDir` 加以指定。

    > workers 为参与计算的节点总数（必须为正整数且大于等于 1），包括用于建立通信的 rank0 节点（对应 pytorch operator 中的 master 节点），默认为 1，可以不填写，即为单机作业。

3. 列出所有作业
    ```
    ➜ arena list
    NAME                          STATUS     TRAINER     AGE  NODE
    pytorch-dist-git              SUCCEEDED  PYTORCHJOB  23h  N/A
    ```

4. 获取作业的详细信息, 可以看到这个作业有 2 个实例，其中 `pytorch-dist-git-master-0` 即为 rank0 的节点。 Arena 借助 `PyTorch-Operator` 简化提交分布式作业的流程，
在 `PyTorch-Operator` 中，会为这个 `master` 实例创建一个 `Service` 便于其他节点通过 `Service` 的 name 访问, 在每个实例中注入环境变量 `MASTER_PORT`、`MASTER_ADDR`、`WORLD_SIZE`、`RANK`，用于
pytorch 建立分布式进程组的初始化工作(dist.init_process_group)。其中 `MASTER_PORT` 自动分配，`MASTER_ADDR` 在 `master` 实例中是 localhost, 其他实例是 `master` 的 `Service` name,
`WORLD_SIZE` 总实例数，`RANK` 当前计算节点的序号，`master` 实例 为 0，`worker` 为实例名尾缀的下标加一，例如，下面的例子中的实例 `pytorch-dist-git-worker-0`，其 `RANK` 为 `0+1=1`。
在 Arena 中，参数 --workers 填写的值包含了 1 个 `master` 节点，因为 `master` 节点也参与训练。
    ```
    ➜ arena get pytorch-local-git
    STATUS: SUCCEEDED
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 1m
    
    NAME              STATUS     TRAINER     AGE  INSTANCE                   NODE
    pytorch-dist-git  SUCCEEDED  PYTORCHJOB  23h  pytorch-dist-git-master-0  172.16.0.210
    pytorch-dist-git  SUCCEEDED  PYTORCHJOB  23h  pytorch-dist-git-worker-0  172.16.0.210
    ```

5. 检查日志
    ``` 
    ➜ arena logs pytorch-dist-git
    WORLD_SIZE: 2, CURRENT_RANK: 0
    args: Namespace(backend='gloo', batch_size=64, data='/root/code/mnist-pytorch', dir='/root/code/mnist-pytorch/logs', epochs=1, log_interval=10, lr=0.01, momentum=0.5, no_cuda=False, save_model=False, seed=1, test_batch_size=1000)
    Using CUDA
    Using distributed PyTorch with gloo backend
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:541: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint8 = np.dtype([("qint8", np.int8, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:542: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint8 = np.dtype([("quint8", np.uint8, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:543: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint16 = np.dtype([("qint16", np.int16, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:544: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint16 = np.dtype([("quint16", np.uint16, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:545: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint32 = np.dtype([("qint32", np.int32, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:550: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      np_resource = np.dtype([("resource", np.ubyte, 1)])
    Train Epoch: 1 [0/60000 (0%)]	loss=2.3000
    Train Epoch: 1 [640/60000 (1%)]	loss=2.2135
    Train Epoch: 1 [1280/60000 (2%)]	loss=2.1705
    Train Epoch: 1 [1920/60000 (3%)]	loss=2.0767
    Train Epoch: 1 [2560/60000 (4%)]	loss=1.8681
    Train Epoch: 1 [3200/60000 (5%)]	loss=1.4142
    Train Epoch: 1 [3840/60000 (6%)]	loss=1.0009
    ...
    ```

    > 对于多实例的分布式作业，默认输出 rank0 (实例为 master 节点)的日志，如果想查看某一个实例的日志，可以通过 -i 实例名查看，例如:
    
    ```
    ➜ arena logs pytorch-dist-git -i pytorch-dist-git-worker-0
    WORLD_SIZE: 2, CURRENT_RANK: 1
    args: Namespace(backend='gloo', batch_size=64, data='/root/code/mnist-pytorch', dir='/root/code/mnist-pytorch/logs', epochs=1, log_interval=10, lr=0.01, momentum=0.5, no_cuda=False, save_model=False, seed=1, test_batch_size=1000)
    Using CUDA
    Using distributed PyTorch with gloo backend
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:541: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint8 = np.dtype([("qint8", np.int8, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:542: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint8 = np.dtype([("quint8", np.uint8, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:543: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint16 = np.dtype([("qint16", np.int16, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:544: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint16 = np.dtype([("quint16", np.uint16, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:545: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint32 = np.dtype([("qint32", np.int32, 1)])
    /opt/conda/lib/python3.7/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:550: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      np_resource = np.dtype([("resource", np.ubyte, 1)])
    Train Epoch: 1 [0/60000 (0%)]	loss=2.3000
    Train Epoch: 1 [640/60000 (1%)]	loss=2.2135
    Train Epoch: 1 [1280/60000 (2%)]	loss=2.1705
    Train Epoch: 1 [1920/60000 (3%)]	loss=2.0767
    Train Epoch: 1 [2560/60000 (4%)]	loss=1.8681
    Train Epoch: 1 [3200/60000 (5%)]	loss=1.4142
    ```

    > 另外，用户查看日志可以通过参数 -t 行数，可以查看尾部倒数几行的日志，如:
    
    ```
    ➜ arena logs pytorch-dist-git -i pytorch-dist-git-worker-0 -t 5
    Train Epoch: 1 [58880/60000 (98%)]	loss=0.2048
    Train Epoch: 1 [59520/60000 (99%)]	loss=0.0646
    
    accuracy=0.9661
    
    ```
    > 更多参数见 `arena logs --help`