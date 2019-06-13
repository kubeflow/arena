## Volcano

Volcano is a batch system built on Kubernetes. It provides a suite of mechanisms currently missing from
Kubernetes that are commonly required by many classes of batch & elastic workload including:

1. machine learning/deep learning,
2. bioinformatics/genomics, and 
3. other "big data" applications.

## Introduction

This chart bootstraps [volcano](https://github.com/volcano-sh/volcano) components like controller, scheduler and admission controller deployments using [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.13+ with CRD support

## Installing Chart

To install the chart with the release name `volcano-release`:

```bash
$ helm install --name volcano-release kubernetes-artifacts/volcano-operator
```

This command deploys volcano in kubernetes cluster with default configuration.  The [configuration](#configuration) section lists the parameters that can be configured during installation.


## Uninstalling the Chart

```bash
$ helm delete volcano-release --purge
```

## Configuration

The following are the list configurable parameters of Volcano Chart and their default values.

| Parameter|Description|Default Value|
|----------------|-----------------|----------------------|
|`basic.image_tag_version`| Docker image version Tag | `default`|
|`basic.controller_image_name`|Controller Docker Image Name|`volcanosh/vk-controllers`|
|`basic.scheduler_image_name`|Scheduler Docker Image Name|`volcanosh/vk-kube-batch`|
|`basic.admission_image_name`|Admission Controller Image Name|`volcanosh/vk-admission`|
|`basic.admission_secret_name`|Volcano Admission Secret Name|`volcano-admission-secret`|
|`basic.scheduler_config_file`|Configuration File name for Scheduler|`kube-batch.conf`|
|`basic.image_pull_secret`|Image Pull Secret|`""`|
|`basic.image_pull_policy`|Image Pull Policy|`IfNotPresent`|
|`basic.admission_app_name`|Admission Controller App Name|`volcano-admission`|
|`basic.controller_app_name`|Controller App Name|`volcano-controller`|
|`basic.scheduler_app_name`|Scheduler App Name|`volcano-scheduler`|

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ helm install --name volcano-release --set basic.image_pull_policy=Always kubernetes-artifacts/volcano-operator
```

The above command set image pull policy to `Always`, so docker image will be pulled each time.


Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ helm install --name volcano-release -f values.yaml kubernetes-artifacts/volcano-operator
```

> **Tip**: You can use the default [values.yaml](values.yaml)

TO verify all deployments are running use the below command

```bash
    kubectl get deployment --all-namespaces | grep {release_name}
```
We should get similar output like given below, where three deployments for controller, admission, scheduler should be running.


NAME                       READY  UP-TO-DATE  AVAILABLE  AGE
{release_name}-admission    1/1    1           1          4s
{release_name}-controllers  1/1    1           1          4s
{release_name}-scheduler    1/1    1           1          4s

TO verify all pods are running use the below command

```bash
    kubectl get pods --all-namespaces | grep {release_name}
```

We should get similar output like given below, where three pods for controller, admission,admissioninit, scheduler should be running.

NAMESPACE     NAME                                          READY    STATUS             RESTARTS   AGE
default       volcano-release-admission-cbfdb8549-dz5hg      1/1     Running            0          33s
default       volcano-release-admission-init-7xmzd           0/1     Completed          0          33s
default       volcano-release-controllers-7967fffb8d-7vnn9   1/1     Running            0          33s
default       volcano-release-scheduler-746f6557d8-9pfg6     1/1     Running            0          33s
