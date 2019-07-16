## Setup

This documentation assumes you have a Kubernetes cluster already available.

If you need help setting up a Kubernetes cluster please refer to [Kubernetes Setup](https://kubernetes.io/docs/setup/).

If you want to use GPUs, be sure to follow the Kubernetes [instructions for enabling GPUs](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/).

Arena doesn't have to run can be run within Kubernetes cluster. It can also be run in your laptop. If you can run `kubectl` to manage the Kubernetes cluster there, you can also use `arena`  to manage Training Jobs.

### Requirements

  * Kubernetes >= 1.11, kubectl >= 1.11
  * helm version [v2.8.2](https://docs.helm.sh/using_helm/#installing-helm) or later 
  * tiller with ths same version of helm should be also installed (https://docs.helm.sh/using_helm/#installing-tiller)

### Steps

1\. Prepare kubeconfig file by using `export KUBECONFIG=/etc/kubernetes/admin.conf` or creating a `~/.kube/config`

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
git clone https://github.com/kubeflow/arena.git
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

7\. Install MPIJob Controller

```
kubectl create -f arena/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
```

8\. Install arena

Prerequisites:

- Go >= 1.8

```
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/kubeflow/arena.git
cd arena
make
```

`arena` binary is located in directory `arena/bin`. You may want add the directory to `$PATH`.


9\. Install and configure kube-arbitrator for gang scheduling(optional)

```
kubectl create -f arena/kubernetes-artifacts/kube-batchd/kube-batched.yaml
```

10\. Enable shell autocompletion

On Linux, please use bash

On CentOS Linux, you may need to install the bash-completion package which is not installed by default.

```
yum install bash-completion -y
```

To add arena autocompletion to your current shell, run source <(arena completion bash).

To add arena autocompletion to your profile, so it is automatically loaded in future shells run:

```
echo "source <(arena completion bash)" >> ~/.bashrc
```

Then you can use [tab] to auto complete the command

```
# arena list
NAME            STATUS   TRAINER  AGE  NODE
tf1             PENDING  TFJOB    0s   N/A
caffe-1080ti-1  RUNNING  HOROVOD  45s  192.168.1.120
# arena get [tab]
caffe-1080ti-1  tf1
```


11\. Enable Host network for training (optional)

The training is not `useHostNetwork` by default. If you'd like to run the training in HostNetwork. You can run the command below:

```
find /charts/ -name values.yaml | xargs sed -i "/useHostNetwork/s/false/true/g"
```

12\. Enable Loadbalancer in the public cloud (optional)

 Kubernetes can be run on AWS, GCE, Azure and Alibaba Cloud, and `LoadBalancer` is supported in their cloud provider. If you want to access tensorboard on the internet directly, you can run the command below:


```
find /charts/ -name "*.yaml" | xargs sed -i "s/NodePort/LoadBalancer/g"
```

> Warning: it's not encouraged to expose the service to the internet, because the service can be attacked by hacker easily.


13\. Enable Ingress in the public cloud (optional)

If you have ingress controller configured, you are able to access tensorboard through ingress. You can run the command below:

```
find /charts/ -name values.yaml | xargs sed -i "/ingress/s/false/true/g"
```

> Warning: it's not encouraged to expose the service to the internet, because the service can be attacked by hacker easily.


14\. Change imagePullPolicy from `Always` to `IfNotPresent` (optional)

```
find /charts/ -name values.yaml| xargs sed -i "s/Always/IfNotPresent/g"
```

> Warning: this may cause the docker images are not up to date if it's already downloaded in node.
