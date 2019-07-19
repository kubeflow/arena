
# Arena 支持MPIJob任务抢占的示例

## 前提条件

- k8s > 1.11

1.利用下列yaml创建`PriorityClass`对象，这里定义了两个优先级`critical`和`medium`:

```yaml
apiVersion: scheduling.k8s.io/v1
description: Used for the critical app
kind: PriorityClass
metadata:
  name: critical
value: 1100000

---

apiVersion: scheduling.k8s.io/v1
description: Used for the medium app
kind: PriorityClass
metadata:
  name: medium
value: 1000000
```

将上述内容保存到`pc.yaml`文件，并且通过下列命令创建:

```
kubectl create -f pc.yaml
```

2.通过arena命令可以看到：在当前Kubernetes集群中只有一张可用GPU卡:

```
# arena top node
NAME          IPADDRESS     ROLE    GPU(Total)  GPU(Allocated)
192.168.0.20  192.168.0.20  master  0           0
192.168.0.21  192.168.0.21  master  0           0
192.168.0.22  192.168.0.22  master  0           0
192.168.0.23  192.168.0.23  <none>  1           0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
0/1 (0%)
```

3.提交一个MPI训练任务，该任务的优先级为`medium`:

参考如下例子 

```
# arena submit mpi          \
    --name=medium           \
    --priority=medium       \
    --gpus=1                \
    --workers=1             \
    --image=registry.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5               \
    "mpirun tail -f /dev/null"
configmap/medium-mpijob created
configmap/medium-mpijob labeled
mpijob.kubeflow.org/medium created
INFO[0000] The Job medium has been submitted successfully
INFO[0000] You can run `arena get medium --type mpijob` to check the job status
```

4.查看该任务的运行状态

```
# arena get medium
STATUS: RUNNING
NAMESPACE: default
PRIORITY: medium
TRAINING DURATION: 58s

NAME    STATUS   TRAINER  AGE  INSTANCE               NODE
medium  RUNNING  MPIJOB   58s  medium-launcher-sz5xj  192.168.0.23
medium  RUNNING  MPIJOB   58s  medium-worker-0        192.168.0.23
```

5.可以看到该任务占用了唯一的一张GPU卡

```
# arena top node -d

NAME:       cn-hangzhou.192.168.0.23
IPADDRESS:  192.168.0.23
ROLE:       <none>

NAMESPACE  NAME             GPU REQUESTS  GPU LIMITS
default    medium-worker-0  1             1

Total GPUs In Node cn-hangzhou.192.168.0.23:      1
Allocated GPUs In Node cn-hangzhou.192.168.0.23:  1 (100%)
-----------------------------------------------------------------------------------------

Allocated/Total GPUs In Cluster:  1/1 (100%)
```

6.再提交一个MPI训练任务，该任务的优先级为`critical`:

```
# arena submit mpi          \
    --name=critical           \
    --priority=critical       \
    --gpus=1                \
    --workers=1             \
    --image=registry.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5               \
    "mpirun tail -f /dev/null"
```

7.检查MPI训练任务`medium`的相关事件，可以发现它被驱逐了。而它被驱逐的原因是由于被更重要的任务`critical`下的Pod也在申请GPU资源，而集群内只有一个可用的GPU资源，所以较低优先级的任务`medium`的`medium-worker-0`被驱逐

```
# kubectl get events --field-selector involvedObject.name=medium-worker-0
LAST SEEN   TYPE     REASON      OBJECT                MESSAGE
15m         Normal   Scheduled   pod/medium-worker-0   Successfully assigned default/medium-worker-0 to 192.168.0.23
14m         Normal   Pulled      pod/medium-worker-0   Container image "registry.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5" already present on machine
14m         Normal   Created     pod/medium-worker-0   Created container mpi
14m         Normal   Started     pod/medium-worker-0   Started container mpi
2m32s       Normal   Preempted   pod/medium-worker-0   by default/critical-worker-0 on node 192.168.0.23
2m32s       Normal   Killing     pod/medium-worker-0   Stopping container mpi
```

8.查看MPI训练任务`medium`的细节信息，发现这个任务已经处于失败状态。

```
# arena get medium
STATUS: FAILED
NAMESPACE: default
PRIORITY: medium
TRAINING DURATION: 12m

NAME    STATUS  TRAINER  AGE  INSTANCE               NODE
medium  FAILED  MPIJOB   20m  medium-launcher-sz5xj  192.168.0.23
```

9.查看MPI训练任务`critical`的细节信息，发现这个任务已经处于运行状态。

```
# arena get critical
STATUS: RUNNING
NAMESPACE: default
PRIORITY: critical
TRAINING DURATION: 10m

NAME      STATUS   TRAINER  AGE  INSTANCE                 NODE
critical  RUNNING  MPIJOB   10m  critical-launcher-mfffs  192.168.0.23
critical  RUNNING  MPIJOB   10m  critical-worker-0        192.168.0.23
```

10.而且也可以通过`arena top node -d`发现这个GPU已经被MPI训练任务`critical`占用。

```
# arena top node -d
NAME:       cn-hangzhou.192.168.0.23
IPADDRESS:  192.168.0.23
ROLE:       <none>

NAMESPACE  NAME               GPU REQUESTS  GPU LIMITS
default    critical-worker-0  1             1

Total GPUs In Node cn-hangzhou.192.168.0.23:      1
Allocated GPUs In Node cn-hangzhou.192.168.0.23:  1 (100%)
-----------------------------------------------------------------------------------------
```

恭喜! 你已经可以通过arena实现对于MPIJob优先级抢占。
