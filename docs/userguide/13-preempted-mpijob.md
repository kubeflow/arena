
# Arena supports Priority and Preemption for MPIJob

## prerequisites

- k8s > 1.11

1.Create `PriorityClass` with the yaml below:

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

Save the template that applies in a file named `pc.yaml`, and create the `PriorityClass`:

```
kubectl create -f pc.yaml
```

2.There is only 1 GPU available in the Kubernetes cluster

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

3.Run the MPI training Job with `medium` priority:


The following command is an example. 

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

4.Get the details of the specific job

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

5.The only one GPU is used by MPI training Job `medium`

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

6.Run the MPI training Job with `critical` priority:

```
# arena submit mpi          \
    --name=critical           \
    --priority=critical       \
    --gpus=1                \
    --workers=1             \
    --image=registry.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5               \
    "mpirun tail -f /dev/null"
```

7.Check MPI Training Job `medium`, and find it's preempted by critical-worker-0

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

8.Check the details of the MPI Training Job `medium`, and it's turned to fail

```
# arena get medium
STATUS: FAILED
NAMESPACE: default
PRIORITY: medium
TRAINING DURATION: 12m

NAME    STATUS  TRAINER  AGE  INSTANCE               NODE
medium  FAILED  MPIJOB   20m  medium-launcher-sz5xj  192.168.0.23
```

9.And check the details of the MPI Training Job `critical`, it's running.

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

10.And we can find the only GPU is used by the MPI Training Job `critical`

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

Congratulations! You've run the the job in priorities and preemptions with `arena` successfully.