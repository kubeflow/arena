# MPI job with specified node tolerations

Arena supports submiting a job and the job tolerates k8s taints nodes(Currently only support mpi job and tf job), the following steps can help you how to use this feature.

1\. query k8s cluster information.

    $ kubectl get nodes
    NAME                       STATUS   ROLES    AGE     VERSION
    cn-beijing.192.168.3.225   Ready    master   2d23h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.226   Ready    master   2d23h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.227   Ready    master   2d23h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.228   Ready    <none>   2d22h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.229   Ready    <none>   2d22h   v1.12.6-aliyun.1
    cn-beijing.192.168.3.230   Ready    <none>   2d22h   v1.12.6-aliyun.1

2\. give some taints for k8s nodes,for example: give taint "gpu_node=invalid:NoSchedule" to node "cn-beijing.192.168.3.228" and node "cn-beijing.192.168.3.229",give taint  "ssd_node=invalid:NoSchedule" to node "cn-beijing.192.168.3.230",now all k8s pod can't schedule to these nodes.

    $ kubectl taint nodes cn-beijing.192.168.3.228 gpu_node=invalid:NoSchedule                                                                            
    node/cn-beijing.192.168.3.228 tainted
    $ kubectl taint nodes cn-beijing.192.168.3.229 gpu_node=invalid:NoSchedule                                                                            
    node/cn-beijing.192.168.3.229 tainted
    $ kubectl taint nodes cn-beijing.192.168.3.230 ssd_node=invalid:NoSchedule                                                                            
    node/cn-beijing.192.168.3.230 tainted


3\. when submitting a job with option ``--toleration``, you can tolerate some nodes which exists taints.

    $ arena submit mpi --name=mpi-dist  \
        --gpus=1 \
        --workers=1 \
        --toleration ssd_node \
        --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
        --tensorboard \
        --loglevel debug \
        "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"


4\. query the job details.

    $ arena get mpi-dist                                                                                                                                 
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 29s

    NAME      STATUS   TRAINER  AGE  INSTANCE                 NODE
    mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-launcher-jgms7  192.168.3.230
    mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-worker-0        192.168.3.230

    Your tensorboard will be available on:
    http://192.168.3.225:30052

the instances of job are running  on node cn-beijing.192.168.3.230(ip is 192.168.3.230,taint is ssd_node=invalid).

5\. you can use ``--toleration`` multiple times,for example: you can use  "--toleration gpu_node --toleration ssd_node" when submitting a job,it represents that the job tolerates nodes which own taint "gpu_node=invalid" and taint "ssd_node=invalid".

    $ arena submit mpi --name=mpi-dist  \
        --gpus=1 \
        --workers=1 \
        --toleration ssd_node \
        --toleration gpu_node \
        --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
        --tensorboard \
        --loglevel debug \
        "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64 --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"

6\. query the job details.

    $ arena get mpi-dist
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 29s

    NAME      STATUS   TRAINER  AGE  INSTANCE                 NODE
    mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-launcher-jgms7  192.168.3.229
    mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-worker-0        192.168.3.230

    Your tensorboard will be available on:
    http://192.168.3.225:30052

!!! note

    you can use "--toleration all" to tolerate all node taints.
