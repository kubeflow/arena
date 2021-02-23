# Submit Tensorflow Job with specified node selectors

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


3\. because there is four roles("PS","Worker","Evaluator","Chief") in tf job,you can use ``--selector`` to assgin nodes, it is effective for all roles. for example:

    $ arena submit tfjob \
        --name=tfjob-with-selector \
        --gpus=1              \
        --workers=1              \
        --selector ssd_node=true \
        --worker-image=cheyang/tf-mnist-distributed:gpu \
        --ps-image=cheyang/tf-mnist-distributed:cpu \
        --ps=1              \
        --tensorboard \
        --loglevel debug \
        "python /app/main.py"

4\. check the job status.

    $ arena get tfjob-with-selector                                                                                                                                       
    STATUS: PENDING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 24s

    NAME                   STATUS   TRAINER  AGE  INSTANCE     NODE
    tfjob-with-selector    RUNNING  TFJOB    24s  tf-ps-0      192.168.3.230
    tfjob-with-selector    PENDING  TFJOB    24s  tf-worker-0  192.168.3.230

    Your tensorboard will be available on:
    http://192.168.3.230:31867


the job(includes "PS" and "Worker") have been running on cn-beijing.192.168.3.230(ip is 192.168.3.230,label is ``ssd_node=true``).

## Roles are running with the different node selectors


5\. you also can assign node to run single role job,for example: if you want to run a job whose role is "PS" on nodes which own label ``ssd_node=true`` and run "Worker" job on nodes which own label ``gpu_node=true``,you can use option ``--ps-selector`` and ``--worker-selector``. 

    $ arena submit tfjob \
        --name=tfjob-with-selector \
        --gpus=1 \
        --workers=1 \
        --ps-selector ssd_node=true \
        --worker-selector gpu_node=true \
        --worker-image=cheyang/tf-mnist-distributed:gpu \
        --ps-image=cheyang/tf-mnist-distributed:cpu \
        --ps=1              \
        --tensorboard \
        --loglevel debug \
        "python /app/main.py"


6\. check the jobs's status.

    $ arena get tf 

    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 23s

    NAME                   STATUS   TRAINER  AGE  INSTANCE     NODE
    tfjob-with-selector    RUNNING  TFJOB    23s  tf-ps-0      192.168.3.230
    tfjob-with-selector    RUNNING  TFJOB    23s  tf-worker-0  192.168.3.228

    Your tensorboard will be available on:
    http://192.168.3.225:30162


the "PS" job is running on cn-beijing.192.168.3.230(ip is 192.168.3.230,label is ``ssd_node=true`` ) and the "Worker" job is running on  cn-beijing.192.168.3.228(ip is 192.168.3.228,label is ``gpu_node=true`` ). 


7\. if you use ``--selector`` in ``arena submit tf`` command and also use ``--ps-selector`` (or ``--worker-selector`` , ``--evaluator-selector`` , ``chief-selector`` ),the value of ``--ps-selector`` would cover value of ``--selector`` ,for example:

    $ arena submit tfjob \
        --name=tfjob-with-selector \
        --gpus=1              \
        --workers=1              \
        --ps-selector ssd_node=true \
        --selector gpu_node=true \
        --worker-image=cheyang/tf-mnist-distributed:gpu \
        --ps-image=cheyang/tf-mnist-distributed:cpu \
        --ps=1              \
        --tensorboard \
        --loglevel debug \
        "python /app/main.py"


"PS" job will be running on nodes whose label is ``ssd_node=true`` ,other jobs will be running on nodes whose label is ``gpu_node=true`` . now verify our conclusions,use follow command to check job status.

    $ arena get tfjob-with-selector                                                                                                                                      
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 39s

    NAME                   STATUS   TRAINER  AGE  INSTANCE     NODE
    tfjob-with-selector    RUNNING  TFJOB    39s  tf-ps-0      192.168.3.230
    tfjob-with-selector    RUNNING  TFJOB    39s  tf-worker-0  192.168.3.228

    Your tensorboard will be available on:
    http://192.168.3.225:32105

As you can see, "PS" job is running on nodes which own label ``ssd_node=true`` ,other jobs are running on nodes which own label ``gpu_node=true``.
