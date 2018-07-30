# Advanced OpenMPI for horovod

OpenMPI is a High Performance Message Passing Library.

For more information,
[visit OpenMPI](https://www.open-mpi.org/).

## Introduction

This chart implements a solution of automatic configuration and deployment for OpenMPI, it deploys the mpi workers as statefulset, and can discover the host list automatically, then start horovod training in one click.

## Prerequisites

- A Kubernetes cluster of Alibaba Cloud Container Service v1.8+ has been created. Refer to [guidance document](https://www.alibabacloud.com/help/doc-detail/53752.html).

- If you should have persistent storage like NAS


## Setup NAS Storage for training

* create `/output/training_logs` in NFS server side, take NFS server `10.244.1.4` 

```
mkdir /nfs
mount -t nfs -o vers=4.0 10.244.1.4:/ /nfs
mkdir -p /nfs/output/training_logs
umount /nfs
```

## Installing the Chart

* To install the chart with the release name `horovod`:

```bash
$ helm install --name horovod aliyun-incubator/horovod
```

* To install with custom values via file:


  ```
  $ helm install --name horovod aliyun-incubator/horovod
  ```
  
  Below is an example of the custom value file horovod.yaml.
  
  ```
mpiWorker:
  number: 5
  podManagementPolicy: Parallel
  image:
    repository: registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/ali-horovod
    tag: gpu-tf-1.6.0
    pullPolicy: Always
  sshPort: 22
  resources:
    limits:
      nvidia.com/gpu: 1
    requests:
      nvidia.com/gpu: 1
mpiMaster:
  image:
    repository: registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/ali-horovod
    tag: gpu-tf-1.6.0
    pullPolicy: Always
  args:
    - /root/hvd-distribute.sh
    - "3"
    - "1"

tensorboard:
  enabled: true
  # NodePort, LoadBalancer
  serviceType: NodePort
  logDir: /output/training_logs
  image:
    repository: registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/tensorboard
    tag: 1.1.0
    pullPolicy: Always

volumeMounts:
   - name: training
     mountPath: /output/training_logs

volumes:
   - name: training
     nfs:
      server: "10.244.1.4"
      path: /output/training_logs

```


## Run OpenMPI

* As the sample above, there are 5 mpi workers

```
# check the statefulset
$ kubectl get sts
NAME                           DESIRED   CURRENT   AGE
horovod-horovod   5         5         1h      5         20s
# wait until the  job done
$ kubectl get job
NAME                               DESIRED   SUCCESSFUL   AGE
horovod-horovod-job   1         1            1h
# check the job's pod status
$ kubectl get po -a -l app=horovod,job-name=horovod-horovod-job
NAME                                     READY     STATUS      RESTARTS   AGE
horovod-horovod-job-j98rm   0/1       Completed   1          1h

```


* Check logs now via 'kubectl logs' on master

```
$ kubectl logs horovod-horovod-job-j98rm 
```


## Uninstalling the Chart

To uninstall/delete the `horovod` deployment:

```bash
$ helm delete horovod
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

## Configuration

The following tables lists the configurable parameters of the Service Tensorflow training
chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `mpiWorker.image.repository` | horovod image | `registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/tf-horovod-k8s` |
| `mpiWorker.number`|  The mpi worker's number | `5` |
| `mpiWorker.image.pullPolicy` | `pullPolicy` for the service mpi worker | `IfNotPresent` |
| `mpiWorker.image.tag` | `tag` for the service mpi worker | `IfNotPresent` |
| `mpiWorker.sshPort` | mpiWorker's sshPort | `22` |
| `mpiWorker.env` | mpiWorker's environment varaibles | `{}` |
| `mpiMaster.image.repository` | horovod image | `registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/tf-horovod-k8s` |
| `mpiMaster.image.pullPolicy` | `pullPolicy` for the service mpi master | `IfNotPresent` |
| `mpiMaster.args` | mpiMaster's args | `{}` |
| `mpiMaster.env` | mpiMaster's environment varaibles | `{}` |
| `tensorboard.enabled` | if the tensorboard is enabled | `true` |
| `tensorboard.image.repository` | `repository` for tensorboard | `registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/tensorboard` |
| `tensorboard.image.pullPolicy` | `pullPolicy` for tensorboard | `IfNotPresent` |
| `tensorboard.serviceType` | `service type` for tensorboard | `LoadBalancer` |
| `volumes`| volume configuration | `{}` |
| `volumeMount`| volume mount configuration | `{}` |



