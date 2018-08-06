## Setup

This documentation assumes you have a Kubernetes cluster already available.

If you need help setting up a Kubernetes cluster please refer to [Kubernetes Setup](https://kubernetes.io/docs/setup/).

If you want to use GPUs, be sure to follow the Kubernetes [instructions for enabling GPUs](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/).

Arena doesn't have to run can be run within Kubernetes cluster. It can also be run in your laptop. If you can run `kubectl` to manage the Kubernetes cluster there, you can also use `arena`  to manage Training Jobs.

### Requirements

  * Kubernetes >= 1.8 [tf-operator requriements](https://github.com/kubeflow/tf-operator#requirements)
  * helm version [v2.8.2](https://docs.helm.sh/using_helm/#installing-helm) or later 
  * tiller with ths same version of helm should be also installed (https://docs.helm.sh/using_helm/#installing-tiller)

### Steps

1. Prepare kubeconfig file by using `export KUBECONFIG=/etc/kubernetes/admin.conf` or creating a `~/.kube/config`

2\. Install kubectl client

Please follow [kubectl installation guide](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

3\. Install Helm client

- Download Helm client from [github.com](https://github.com/helm/helm/releases)  
- Unpack it (tar -zxvf helm-v2.8.2-linux-amd64.tgz)
- Find the `helm` binary in the unpacked directory, and move it to its desired destination (mv linux-amd64/helm /usr/local/bin/helm)

Then run `helm list` to check if the the kubernetes can be managed successfully by helm.

```
# helm list
# echo $?
0
```

4\. Download the charts

```
mkdir /charts
git clone https://github.com/AliyunContainerService/arena.git
cp -r arena/charts/* /charts
```

5\. Install TFJob Controller

```
kubectl create -f arena/kubernetes-artifacts/jobmon/jobmon-role.yaml
kubectl create -f arena/kubernetes-artifacts/tf-operator/tf-operator.yaml
```

6\. Install Dashboard

```
kubectl create -f arena/kubernetes-artifacts/dashboard/dashboard.yaml
```

7\. Install arena

Prerequisites:

- Go >= 1.8

```
mkdir -p $GOPATH/src/github.com/kubeflow
cd $GOPATH/src/github.com/kubeflow
git clone https://github.com/AliyunContainerService/arena.git
cd arena
make
```

`arena` binary is located in directory `arena/bin`. You may want add the dirctory to `$PATH`.
