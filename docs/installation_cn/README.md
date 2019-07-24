## 部署

本文档假设您已经有可用的 Kubernetes 集群。

如果您需要有关 Kubernetes 集群设置的帮助，请参阅 [Kubernetes 设置](https://kubernetes.io/docs/setup/)。

如果您希望使用 GPU，请务必按照 Kubernetes [GPU 启用说明](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/) 操作。

Arena 并非必需在 Kubernetes 集群内运行。它也可以在您的笔记本电脑中运行。如果您可以运行 `kubectl` 以管理 Kubernetes 集群，那么也可以使用 `arena` 管理训练作业。

### 要求

  * Kubernetes >= 1.11, kubectl >= 1.11
  * helm 版本 [v2.8.2](https://docs.helm.sh/using_helm/#installing-helm) 或更新版本 
  * 此外还要部署与 helm 版本相同的 tiller(https://docs.helm.sh/using_helm/#installing-tiller)

### 步骤

1\.通过使用 `export KUBECONFIG=/etc/kubernetes/admin.conf` 或创建一个 `~/.kube/config` 来准备 kubeconfig 文件

2\.安装 kubectl 客户端

请按照 [kubectl 安装指南] 操作(https://kubernetes.io/docs/tasks/tools/install-kubectl/)

3\.安装 Helm 客户端

- 从 [github.com] 下载 Helm 客户端(https://github.com/helm/helm/releases)  
- 将下载到的文件解压缩 (tar -zxvf helm-v2.8.2-linux-amd64.tgz)
- 在解压缩目录中找到 `helm` 二进制文件，将其移到所需目标位置 (mv linux-amd64/helm /usr/local/bin/helm)

然后运行 `helm list` 以检查 helm 能否成功管理 kubernetes。

```
#helm list
#echo $?
0
```

4\.下载 Chart

```
mkdir /charts
git clone https://github.com/kubeflow/arena.git
cp -r arena/charts/* /charts
```

5\.安装 TFJob 控制器

```
kubectl create -f arena/kubernetes-artifacts/jobmon/jobmon-role.yaml
kubectl create -f arena/kubernetes-artifacts/tf-operator/tf-crd.yaml
kubectl create -f arena/kubernetes-artifacts/tf-operator/tf-operator.yaml
```

6\.安装控制台 (可选)

```
kubectl create -f arena/kubernetes-artifacts/dashboard/dashboard.yaml
```

7\.安装 MPIJob 控制器

```
kubectl create -f arena/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
```

8\.安装 arena

先决条件：

- Go >= 1.8

```
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/kubeflow/arena.git
cd arena
make
```

`arena` 二进制文件位于 `arena/bin` 目录下。您可能希望将目录添加到 `$PATH`。


9\.安装并为群调度配置 kube-arbitrator（可选）

```
kubectl create -f arena/kubernetes-artifacts/kube-batchd/kube-batched.yaml
```

10\.启用 shell 自动完成

在 Linux 上，请使用 bash

在 CentOS Linux 上，您可能需要安装默认并未安装的 bash-completion 包。

```
yum install bash-completion -y
```

要为当前 shell 添加 arena 自动完成，请运行 source <(arena completion bash)。

通过如下方法向您的配置文件添加 arena 自动完成功能，以便将来 shell 运行时可以自动加载此功能：

```
echo "source <(arena completion bash)" >> ~/.bashrc
```

然后，你可以使用 [TAB] 来自动完成命令

```
#arena list
NAME STATUS TRAINER AGE NODE
tf1 PENDING TFJOB 0s N/A
caffe-1080ti-1 RUNNING HOROVOD 45s 192.168.1.120
#arena get [tab]
caffe-1080ti-1 tf1
```


11\.为训练启用主机网络（可选）

默认情况下，训练并非 `useHostNetwork`。如果您希望在 HostNetwork 中运行训练。可以运行如下命令：

```
find /charts/ -name values.yaml | xargs sed -i "/useHostNetwork/s/false/true/g"
```

12\.在公共云中启用 Loadbalancer

 Kubernetes 可在 AWS、GCE、Azure 和阿里云中运行，其云提供商支持 `LoadBalancer`。如果您希望在互联网上直接访问 tensorboard，可以运行如下代码：

```
find /charts/ -name "*.yaml" | xargs sed -i "s/NodePort/LoadBalancer/g"
```

> 警告：我们不鼓励将服务公开给互联网，因为这种做法会导致服务受黑客攻击。

13\. 在公共云中启用 Ingress

Kubernetes 可在 AWS、GCE、Azure 和阿里云中运行，其云提供商支持 `Ingress`。如果您希望在互联网上直接通过统一入口访问 tensorboard，可以运行如下代码：

```
find /charts/ -name values.yaml | xargs sed -i "/ingress/s/false/true/g"
```

> 警告：我们不鼓励将服务公开给互联网，因为这种做法会导致服务受黑客攻击。

14\. 将 imagePullPolicy 策略由 `Always` 修改为 `IfNotPresent` (可选)

```
find /charts/ -name values.yaml| xargs sed -i "s/Always/IfNotPresent/g"
```

> 警告: 这会导致容器镜像可能不是最新更新版本。
