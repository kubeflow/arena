
Arena支持给提交的任务指定运行的节点（目前仅支持mpi和tf类型的任务）。

下面展示一些使用例子。

1.查询k8s集群信息：
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
2.为一些k8s节点指定标签。例如，为节点"cn-beijing.192.168.3.228"和节点"cn-beijing.192.168.3.229"指定标签"gpu_node=ok"，为节点"cn-beijing.192.168.3.230"指定标签"ssd_node=ok"。
```
# kubectl label nodes cn-beijing.192.168.3.228 gpu_node=ok
node/cn-beijing.192.168.3.228 labeled
# kubectl label nodes cn-beijing.192.168.3.229 gpu_node=ok
node/cn-beijing.192.168.3.229 labeled
# kubectl label nodes cn-beijing.192.168.3.230 ssd_node=ok
node/cn-beijing.192.168.3.230 labeled
``` 
## MPI类型的job
1.当提交一些job时，可以通过"--selector"选项来确定这些job运行在哪些节点上
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
2.查询job信息
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
可以看到job已经运行在节点cn-beijing.192.168.3.228(ip是192.168.3.229)上了。
the jobs have been running  on node cn-beijing.192.168.3.228(ip is 192.168.3.229).

3.你可以多次使用"--selector"选项，例如：你可以在arena的提交命令中使用"--selector gpu_node=ok --selector ssd_node=ok",这代表你需要将job运行在那些同时拥有标签"gpu_node=ok"和标签"ssd_node=ok"的节点上

## TF类型的job
 
1.因为在tf类型的job当中，存在四种角色（"PS","Worker","Evaluator","Chief"），你可以使用"--selector"来指定job运行在哪些节点上。
```
arena submit tfjob \
      --name=tf \
      --gpus=1              \
      --workers=1              \
      --selector ssd_node=ok \
      --work-image=cheyang/tf-mnist-distributed:gpu \
      --ps-image=cheyang/tf-mnist-distributed:cpu \
      --ps=1              \
      --tensorboard \
      --loglevel debug \
      "python /app/main.py"
```
使用如下命令检查节点状态：

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

可以看到"PS"类型的job和"Worker"类型的job都运行在了节点cn-beijing.192.168.3.230(ip是192.168.3.230,标签是"ssd_node=ok")上了。
the jobs(include "PS" and "Worker") have been running on cn-beijing.192.168.3.230(ip is 192.168.3.230,label is "ssd_node=ok").

2.你也可以单独指定一种角色的job运行在哪些节点上，例如：如果你希望把"PS" job运行在标签为ssd_node="ok"节点上，把"Worker" job运行在标签为"gpu_node=ok"的节点上，可以使用"--ps-selector"和"--worker-selector"。

```
arena submit tfjob \
      --name=tf \
      --gpus=1              \
      --workers=1              \
      --ps-selector ssd_node=ok \
      --worker-selector gpu_node=ok \
      --work-image=cheyang/tf-mnist-distributed:gpu \
      --ps-image=cheyang/tf-mnist-distributed:cpu \
      --ps=1              \
      --tensorboard \
      --loglevel debug \
      "python /app/main.py"
```
检查job的状态:

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
"PS" job运行在节点cn-beijing.192.168.3.230(ip是192.168.3.230,标签是"ssd_node=ok")，"Worker" job运行在节点cn-beijing.192.168.3.228(ip是192.168.3.228,标签是"gpu_node=ok")上。

3.如果你同时使用"--selector"和"--ps-selector"（或者"--worker-selector","--evaluator-selector","chief-selector"），那么"--ps-selector"的值会覆盖"--selector"的值。，例如：

```
arena submit tfjob \
      --name=tf \
      --gpus=1              \
      --workers=1              \
      --ps-selector ssd_node=ok \
      --selector gpu_node=ok \
      --work-image=cheyang/tf-mnist-distributed:gpu \
      --ps-image=cheyang/tf-mnist-distributed:cpu \
      --ps=1              \
      --tensorboard \
      --loglevel debug \
      "python /app/main.py"
```
理论上"--selector"会应用到所有角色的job中，在上面的命令中，所有角色的job将会被调度到标签为gpu_node=ok的节点上，但是因为有"--ps-selector"，那么"PS" job会被调度到标签为ssd_node=ok上，而不是标签为gpu_node=ok的节点上。
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
正如你所看到的，"PS" job被调度到拥有标签为"ssd_node=ok"的节点上，其他节点被调度到标签为"gpu_node=ok"的节点上。
