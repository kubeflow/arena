This example shows how to use `Arena` to submit a pytorch distributed job. This example will download the source code from git url.

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

    > the source code will be downloaded and extracted to the directory `code/` of the working directory. The default working directory is `/root`, you can also specify by using `--workingDir`.

    >`workers` is the total number of nodes participating in the training (must be a positive integer and greater than or equal to 1), including rank0 node used to establish communication (corresponding to the `master` node in the pytorch-operator). The default value of the parameter is 1, which can not be set, as a stand-alone job.
    

3. List all the jobs.
    ```
    ➜ arena list
    NAME                          STATUS     TRAINER     AGE  NODE
    pytorch-dist-git              SUCCEEDED  PYTORCHJOB  23h  N/A
    ```

4. Get the details of the this job. There are 2 instances of this job, and instance `pytorch-dist-git-master-0` is the rank0. Arena simplifies the process of submitting distributed jobs with `PyTorch-Operator`.
A `Service` will be created for this `master` instance for other nodes to access through the name of `Service` in `PyTorch-Operator`, and inject environment variables into each instance: `MASTER_PORT`、`MASTER_ADDR`、`WORLD_SIZE`、`RANK`. Initialization of distributed process group for pytorch（ dist.init_ process_ group). `MASTER_PORT` auto assign, `MASTER_ADDR` is "localhost" in the `master` instance, and other instances are `Service` name of the `master`,`WORLD_SIZE` is the total number of instances, and `RANK` is the serial number of the current calculation node, and `master` is 0, `Worker` instance is the index of instance name suffix plus one. For example, in the following example, `RANK` of instance `pytorch-dist-git-worker-0` is `0 + 1 = 1`
In Arena, the value filled in by the parameter `--workers` contains one `master` instance, because `master` is also involved in training.
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

5. Check logs.
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

    > For multi instances of distributed job, the default output is the log of rank0 (the instance is the `master` node). If you want to view the log of the specific instance, you can view it by `-i` instance name, for example:
    
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

    > In addition, user can view the logs of the last few lines through the parameter `-t` lines num, such as:
    
    ```
    ➜ arena logs pytorch-dist-git -i pytorch-dist-git-worker-0 -t 5
    Train Epoch: 1 [58880/60000 (98%)]	loss=0.2048
    Train Epoch: 1 [59520/60000 (99%)]	loss=0.0646
    
    accuracy=0.9661
    
    ```
    > For more parameters, see ` arena logs -- help`
  