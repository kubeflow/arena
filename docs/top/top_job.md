# Display GPU Usage For Training Job

The `arena top job` command allows you to see the gpu resource consumption for training jobs.


1\. display gpu resource consumption of all training jobs:

```
$ arena top job
NAME                           STATUS     TRAINER  AGE  GPU(Requested)  GPU(Allocated)  NODE
dawnbench-1x1-v4               PENDING    MPIJOB   20d  100             0               N/A
dawnbench-1x1-v7               SUCCEEDED  MPIJOB   12h  1               0               192.168.8.3
dawnbench-1x1-v6               SUCCEEDED  MPIJOB   12h  1               0               192.168.8.3
tf-distributed-test            FAILED     TFJOB    11m  0               0               N/A
tf-git                         SUCCEEDED  TFJOB    14m  0               0               N/A
dawnbench-1x1-v3               SUCCEEDED  MPIJOB   1d   0               0               N/A
dawnbench-1x1-v2               SUCCEEDED  MPIJOB   12h  0               0               N/A
dawnbench-1x1-v1               SUCCEEDED  MPIJOB   12h  0               0               N/A
mpi-test                       SUCCEEDED  MPIJOB   12h  0               0               N/A
elastic-training               SUCCEEDED  ETJOB    40d  0               0               N/A
horovod-resnet50-v2-4x8-fluid  SUCCEEDED  MPIJOB   1h   0               0               N/A
horovod-resnet50-v2-4x8-nfs    SUCCEEDED  MPIJOB   2h   0               0               N/A

Total Allocated/Requested GPUs of Training Jobs: 0/0
```

2\. display gpu resource consumption of single training job:

```
$ arena top job dawnbench-1x1-v6
Name:      dawnbench-1x1-v6
Status:    SUCCEEDED
Namespace: default
Priority:  N/A
Trainer:   MPIJOB
Duration:  12h

Instances:
  NAME                             STATUS     GPU(Request)  NODE         GPU(DeviceIndex)  GPU(DutyCycle)  GPU_MEMORY(Used/Total)
  ----                             ------     ------------  ----         ----------------  --------------  ---------------
  dawnbench-1x1-v6-launcher-686cl  Completed  0             192.168.8.3  N/A               N/A             N/A

GPUs:
  Allocated/Requested GPUs of Job: 0/1
```

3\. If you need to monitor the training job in real time, "-r" is required:

```
$ arena top job dawnbench-1x1-v6 -r

Name:      dawnbench-1x1-v6
Status:    SUCCEEDED
Namespace: default
Priority:  N/A
Trainer:   MPIJOB
Duration:  12h

Instances:
  NAME                             STATUS     GPU(Request)  NODE         GPU(DeviceIndex)  GPU(DutyCycle)  GPU_MEMORY(Used/Total)
  ----                             ------     ------------  ----         ----------------  --------------  ---------------
  dawnbench-1x1-v6-launcher-686cl  Completed  0             192.168.8.3  N/A               N/A             N/A

GPUs:
  Allocated/Requested GPUs of Job: 0/1
------------------------------------------- 2021-02-22 17:42:25 ----------------------------------------------------
Name:      dawnbench-1x1-v6
Status:    SUCCEEDED
Namespace: default
Priority:  N/A
Trainer:   MPIJOB
Duration:  12h

Instances:
  NAME                             STATUS     GPU(Request)  NODE         GPU(DeviceIndex)  GPU(DutyCycle)  GPU_MEMORY(Used/Total)
  ----                             ------     ------------  ----         ----------------  --------------  ---------------
  dawnbench-1x1-v6-launcher-686cl  Completed  0             192.168.8.3  N/A               N/A             N/A

GPUs:
  Allocated/Requested GPUs of Job: 0/1
------------------------------------------- 2021-02-22 17:42:27 ----------------------------------------------------
```