## Setup

This documentation assumes you have a Kubernetes cluster already available.

If you need help setting up a Kubernetes cluster please refer to [Kubernetes Setup](https://kubernetes.io/docs/setup/).

If you want to use GPUs, be sure to follow the Kubernetes [instructions for enabling GPUs](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/).

Arena doesn't have to run can be run within Kubernetes cluster. It can also be run in your laptop. If you can run `kubectl` to manage the Kubernetes cluster there, you can also use `arena`  to manage Training Jobs.

### Requirements

  * Linux OS
  * Kubernetes >= 1.11, kubectl >= 1.11
  * helm version [v2.9.2](https://docs.helm.sh/using_helm/#installing-helm) or later 
  * tiller with ths same version of helm should be also installed (https://docs.helm.sh/using_helm/#installing-tiller)

### Steps

1\. Prepare kubeconfig file by using `export KUBECONFIG=/etc/kubernetes/admin.conf` or creating a `~/.kube/config`

2\. Download the latest installer from [Release Page](https://github.com/kubeflow/arena/releases), and rename it to `arena-installer.tar.gz`

3\. Untar the installer package

```
# tar -xvf arena-installer.tar.gz 
```

4\. Setup Environment Varaibles for customization

4.1\. If you'd like to train and serving in hostNetwork

```
export USE_HOSTNETWORK=true
```

4.2\. If you'd like to customize Kubernetes namespace of arena infrastructure  

```
export NAMESPACE={your namespace}
```

4.3\. If you'd like to use your private docker registry instead of `ACR(Alibaba Cloud Container Registry)`:

```
export DOCKER_REGISTRY={your docker registry}
```

4.4\. If you'd like to deploy prometheus in `ACK(Alibaba Container Service for Kubernetes)`

```
export USE_PROMETHEUS=true
export PLATFORM=ack
```

4.5\. If you'd like to use Cloud loadbalancer

```
export USE_LOADBALANCER=true
```

5\. Install arena

```
# cd arena-installer
# sudo ./install.sh
```

6\. Enable shell autocompletion

On Linux, please use bash

On CentOS Linux, you may need to install the bash-completion package which is not installed by default.

```
yum install bash-completion -y
```

On Debian or Ubuntu Linux you may need to install with 

```
apt-get install bash-completion
```

To add arena autocompletion to your current shell, run `source <(arena completion bash)`.

On MacOS, please use bash

You can install it with Homebrew:

```
brew install bash-completion@2
```

To add arena autocompletion to your profile, so it is automatically loaded in future shells run:

```
echo "source <(arena completion bash)" >> ~/.bashrc
chmod u+x ~/.bashrc
```

For MacOS, add the following to your `~/.bashrc` file: 

```
echo "source $(brew --prefix)/etc/profile.d/bash_completion.sh" >> ~/.bashrc
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
