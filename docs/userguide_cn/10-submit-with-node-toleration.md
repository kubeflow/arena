
Arena支持将提交的job运行在k8s污点上（目前仅支持mpi和tf类型的 job）

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
2.为k8s节点打上一些污点，例如：为节点"cn-beijing.192.168.3.228"和节点"cn-beijing.192.168.3.229"打上污点"gpu_node=invalid:NoSchedule"，为节点"cn-beijing.192.168.3.230"打上污点"ssd_node=invalid:NoSchedule"。现在，所有pod都不能调度到这些节点了。
```
# kubectl taint nodes cn-beijing.192.168.3.228 gpu_node=invalid:NoSchedule                                                                            
node/cn-beijing.192.168.3.228 tainted
# kubectl taint nodes cn-beijing.192.168.3.229 gpu_node=invalid:NoSchedule                                                                            
node/cn-beijing.192.168.3.229 tainted
# kubectl taint nodes cn-beijing.192.168.3.230 ssd_node=invalid:NoSchedule                                                                            
node/cn-beijing.192.168.3.230 tainted
``` 
3.当提交一个job时，你可以使用"--toleration"来容忍一些带有污点的k8s节点。
```
# arena submit mpi --name=mpi-dist  \
              --gpus=1              \
              --workers=1              \
	      --toleration ssd_node \
              --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
              --tensorboard \
              --loglevel debug \
              "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"
```
查询job信息：
```
# arena get mpi-dist                                                                                                                                 
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 29s

NAME      STATUS   TRAINER  AGE  INSTANCE                 NODE
mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-launcher-jgms7  192.168.3.230
mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-worker-0        192.168.3.230

Your tensorboard will be available on:
http://192.168.3.225:30052
```
job已经运行在节点cn-beijing.192.168.3.230(ip为192.168.3.230,污点为ssd_node=invalid)上了。

4.你可以在同一个命令中多次使用"--toleration"。例如，你可以在命令中使用"--toleration gpu_node --toleration ssd_node"，它代表既可以容忍有污点"gpu_node"的节点，又可以容忍污点"ssd_node"的节点。

```
# arena submit mpi --name=mpi-dist  \
              --gpus=1              \
              --workers=1              \
              --toleration ssd_node \
              --toleration gpu_node \
              --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
              --tensorboard \
              --loglevel debug \
              "mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"
```
查询job状态：

```
# arena get mpi-dist
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 29s

NAME      STATUS   TRAINER  AGE  INSTANCE                 NODE
mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-launcher-jgms7  192.168.3.229
mpi-dist  RUNNING  MPIJOB   29s  mpi-dist-worker-0        192.168.3.230

Your tensorboard will be available on:
http://192.168.3.225:30052
```
5.你可以使用"--toleration all"来容忍所有节点上的所有污点。
