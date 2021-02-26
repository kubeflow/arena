# Only Install Arena Binary

Sometimes, we don’t need to install arena completely, we just want to install the arena executable file on our laptop. This document can help you.

1\. Prepare kubeconfig file with name `config` and place it  to `~/.kube`.

2\. Download the latest installer from [Release Page](https://github.com/kubeflow/arena/releases), and rename it to `arena-installer.tar.gz`

3\. execute the following command to untar the package:

```
$ tar -xvf arena-installer.tar.gz
```

4\. copy the executable files to `/usr/local/bin` and rename them:

```
$ chmod +x bin/*
$ cp bin/helm /usr/local/bin/arena-helm
$ cp bin/kubectl /usr/local/bin/arena-kubectl
$ cp /bin/arena /usr/local/bin/arena
```

5\. copy the `charts` directory to the home directory of current user.

```
$ cp -a charts ~/
```

Then you can use [tab] to auto complete the command:

```
$ arena list
NAME            STATUS   TRAINER  AGE  NODE
tf1             PENDING  TFJOB    0s   N/A
caffe-1080ti-1  RUNNING  HOROVOD  45s  192.168.1.120

$ arena get [tab]
caffe-1080ti-1  tf1
```
