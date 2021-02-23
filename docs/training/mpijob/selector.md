# MPI job with specified node selectors

Arena supports assigning jobs to some k8s particular nodes(Currently only support mpi job and tf job), the following steps will show how to use this feature.


## Label the nodes

1\. query k8s cluster information.

    $ kubectl get nodes
    NAME                       STATUS   ROLES    AGE     VERSION
    cn-beijing.192.168.3.225   Ready    master   2d23h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.226   Ready    master   2d23h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.227   Ready    master   2d23h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.228   Ready    <none>   2d22h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.229   Ready    <none>   2d22h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.230   Ready    <none>   2d22h   v1.12.6-aliyun.1

2\. label the nodes,for example: label node cn-beijing.192.168.3.228" and node cn-beijing.192.168.3.229 with ``gpu_node=true`` ,label node cn-beijing.192.168.3.230 with ``ssd_node=true``.

    $ kubectl label nodes cn-beijing.192.168.3.228 gpu_node=true
    node/cn-beijing.192.168.3.228 labeled
    $ kubectl label nodes cn-beijing.192.168.3.229 gpu_node=true
    node/cn-beijing.192.168.3.229 labeled
    $ kubectl label nodes cn-beijing.192.168.3.230 ssd_node=true
    node/cn-beijing.192.168.3.230 labeled


## Roles are running with the same node selectors

3\. you can use ``--selector`` to assgin nodes, for example::

    $ arena submit mpi --name=mpi-dist \
        --gpus=1 \
        --workers=1 \
        --selector gpu_node=true \
        --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
        --tensorboard \
        "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"

4\. check the job status.

    $ arena get mpi-dist
        Name:        mpi-dist
        Status:      RUNNING
        Namespace:   default
        Priority:    N/A
        Trainer:     MPIJOB
        Duration:    32s

        Instances:
        NAME               STATUS             AGE  IS_CHIEF  GPU(Requested)  NODE
        ----               ------             ---  --------  --------------  ----
        mpi-dist-worker-0  Running            32s  false     1               cn-beijing.192.168.3.228

        Tensorboard:
        Your tensorboard will be available on:
        http://192.168.3.228:30099


the job have been running on cn-beijing.192.168.3.228(ip is 192.168.3.228,label is ``gpu_node=true``).