## Arena Java SDK

### 1.introduction

arena-java-sdk is used to invoke [arena](https://github.com/kubeflow/arena) to manage training jobs，currently we only support mpi job and tf job.

### 2.requirement 

* os type: we only support linux system and mac OS.
* kubeconfig: the kubeconfig file of cluster is required,and you should put it to ~/.kube with name 'config'.
* arena: the arena-java-sdk will invoke arena command tool,please refer [arena installation](https://github.com/kubeflow/arena/blob/master/docs/installation/INSTALL_FROM_BINARY.md) to install arena.
* jdk 1.8: it is required for arena-java-sdk.

### 3.installation

(1). make sure jdk is ok:

```
# java -version
java version "1.8.0_241"
Java(TM) SE Runtime Environment (build 1.8.0_241-b07)
Java HotSpot(TM) 64-Bit Server VM (build 25.241-b07, mixed mode)
```
(2). get the arena installation package:

from the [arena download page](https://github.com/kubeflow/arena/releases) to download package,assume that we download the package 'arena-installer-0.3.1-b96e1ac-linux-amd64.tar.gz
' and put it to /tmp/arena-installer-0.3.1-b96e1ac-linux-amd64.tar.gz

(3). excute installation script

use following commands to install arena-java-sdk:

```
# cd arena-java-sdk
# bash install.sh /tmp/arena-installer-0.3.1-b96e1ac-linux-amd64.tar.gz
```
the install.sh requires the arena installation package.

(4). test arena-java-sdk is ok:

```
# source ~/.bashrc
# cd arena-java-sdk
# cd examples
# javac MPIJobTest.java
# java MPIJobTest
``` 
