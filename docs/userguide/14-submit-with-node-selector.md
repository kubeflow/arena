

Arena supports assigning jobs to some k8s particular nodes(Currently only support mpi job and tf job).

some usage examples in here.
 
1.query k8s cluster information:
```
# kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.3.225   Ready    master   2d23h   v1.12.6-aliyun.1
cn-beijing.192.168.3.226   Ready    master   2d23h   v1.12.6-aliyun.1
cn-beijing.192.168.3.227   Ready    master   2d23h   v1.12.6-aliyun.1
cn-beijing.192.168.3.228   Ready    <none>   2d22h   v1.12.6-aliyun.1
cn-beijing.192.168.3.229   Ready    <none>   2d22h   v1.12.6-aliyun.1
cn-beijing.192.168.3.230   Ready    <none>   2d22h   v1.12.6-aliyun.1
```
2.give a label to nodes,for example: give label "gpu_node=ok" to node "cn-beijing.192.168.3.228" and node "cn-beijing.192.168.3.229",give label "ssd_node=ok" to node "cn-beijing.192.168.3.230"
```
# kubectl label nodes cn-beijing.192.168.3.228 gpu_node=ok
node/cn-beijing.192.168.3.228 labeled
# kubectl label nodes cn-beijing.192.168.3.229 gpu_node=ok
node/cn-beijing.192.168.3.229 labeled
# kubectl label nodes cn-beijing.192.168.3.230 ssd_node=ok
node/cn-beijing.192.168.3.230 labeled
``` 
## for MPI job
1.when submit a job,you can assign nodes to run job with operation "--selector" 
```
# arena submit mpi --name=mpi-dist  \
              --gpus=1              \
              --workers=1              \
	      --selector gpu_node=ok \
              --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
              --tensorboard \
              --loglevel debug \
              "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"
```
2.query the job information
```
# arena get mpi-dist                                                                                                                                  
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 21s

NAME      STATUS   TRAINER  AGE  INSTANCE                 NODE
mpi-dist  RUNNING  MPIJOB   21s  mpi-dist-launcher-7jn4q  192.168.3.229
mpi-dist  RUNNING  MPIJOB   21s  mpi-dist-worker-0        192.168.3.229

Your tensorboard will be available on:
http://192.168.3.225:31611
```
the jobs are running  on node cn-beijing.192.168.3.229(ip is 192.168.3.229).

3.you can use "--selector" multiple times,for example you can use  "--selector gpu_node=ok --selector ssd_node=ok" in arena submit command,it represents that the job should be running on nodes which own label "gpu_node=ok" and label "ssd_node=ok".

## for tf job 

1.because there is four roles("PS","Worker","Evaluator","Chief") in tf job,you can use "--selector" to assgin nodes,this is effective for all roles.for example:
```
arena submit tfjob \
      --name=tf \
      --gpus=1              \
      --workers=1              \
      --selector ssd_node=ok \
      --workerImage=cheyang/tf-mnist-distributed:gpu \
      --psImage=cheyang/tf-mnist-distributed:cpu \
      --ps=1              \
      --tensorboard \
      --loglevel debug \
      "python /app/main.py"
```
use follow command to check the job status:

```
# arena get tf                                                                                                                                       
STATUS: PENDING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 24s

NAME  STATUS   TRAINER  AGE  INSTANCE     NODE
tf    RUNNING  TFJOB    24s  tf-ps-0      192.168.3.230
tf    PENDING  TFJOB    24s  tf-worker-0  192.168.3.230

Your tensorboard will be available on:
http://192.168.3.225:31867
```

the jobs(include "PS" and "Worker") have been running on cn-beijing.192.168.3.230(ip is 192.168.3.230,label is "ssd_node=ok").

2.you also can assign node to run single role job,for example: if you want to run a job whose role is "PS" on nodes which own label ssd_node="ok" and run "Worker" job on nodes which own label gpu_node=ok,you can use option "--ps-selector" and "--worker-selector" 
```
arena submit tfjob \
      --name=tf \
      --gpus=1              \
      --workers=1              \
      --ps-selector ssd_node=ok \
      --worker-selector gpu_node=ok \
      --workerImage=cheyang/tf-mnist-distributed:gpu \
      --psImage=cheyang/tf-mnist-distributed:cpu \
      --ps=1              \
      --tensorboard \
      --loglevel debug \
      "python /app/main.py"
```

then check the jobs's status:

```
# arena get tf                                                                                                                                       
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 23s

NAME  STATUS   TRAINER  AGE  INSTANCE     NODE
tf    RUNNING  TFJOB    23s  tf-ps-0      192.168.3.230
tf    RUNNING  TFJOB    23s  tf-worker-0  192.168.3.228

Your tensorboard will be available on:
http://192.168.3.225:30162
```

the "PS" job is running on cn-beijing.192.168.3.230(ip is 192.168.3.230,label is "ssd_node=ok") and the "Worker" job is running on  cn-beijing.192.168.3.228(ip is 192.168.3.228,label is "gpu_node=ok") 

3.if you use "--selector" in "arena submit tf" command and also use "--ps-selector"(or "--worker-selector","--evaluator-selector","chief-selector"),the value of "--ps-selector" would cover value of "--selector",for example:

```
arena submit tfjob \
      --name=tf \
      --gpus=1              \
      --workers=1              \
      --ps-selector ssd_node=ok \
      --selector gpu_node=ok \
      --workerImage=cheyang/tf-mnist-distributed:gpu \
      --psImage=cheyang/tf-mnist-distributed:cpu \
      --ps=1              \
      --tensorboard \
      --loglevel debug \
      "python /app/main.py"
```

"PS" job will be running on nodes whose label is "ssd_node=ok",other jobs will be running on nodes whose label is "gpu_node=ok",now verify our conclusions,use follow command to check job status.
```
# arena get tf                                                                                                                                       
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 39s

NAME  STATUS   TRAINER  AGE  INSTANCE     NODE
tf    RUNNING  TFJOB    39s  tf-ps-0      192.168.3.230
tf    RUNNING  TFJOB    39s  tf-worker-0  192.168.3.228

Your tensorboard will be available on:
http://192.168.3.225:32105
```
as you can see, "PS" job is running on nodes which own label "ssd_node=ok",other jobs are running on nodes which own label "gpu_node=ok"
