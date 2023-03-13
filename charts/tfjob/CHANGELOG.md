### 0.1.0

* chief, worker, ps, evaluator

### 0.2.0

* support Tensorboard

### 0.3.0

* support git, rsync

### 0.4.0

* set default git and rsync image


### 0.5.0

* Add cleanup policy


### 0.6.0

* support multiple dataDirs


### 0.7.0

* worker and ps soft affinity
* standalone worker without host network


### 0.8.0

* support gang scheduler

### 0.9.0

* Fix tensorboard affinity issue

### 0.10.0

* Support tensorboard loading event log from hdfs path

### 0.11.0

* Support RoCE by using https://github.com/Mellanox/k8s-rdma-sriov-dev-plugin, only support hostNetwork now.

### 0.12.0

* Support Estimator Mode, Chief for new version esitmator(>1.4), Master for old version esitmator(<= 1.4)

### 0.13.0

* add annotations, for cloud provider customization

### 0.14.0

* Make Hostnetwork as false by default

### 0.15.0

* Fix hostnetwork issue which is introduced by ENI

### 0.16.0

* Remove -tfjob from tfjob name

### 0.17.0

* Support Tensorboard for Ingress


### 0.18.0

* Update git-sync image to fix could not lock config file issue

### 0.19.0

* Add resource limits/requests to tfjob initContainer and tensorboard

### 0.20.0

* Add PodSecurityContext support for RunAsUser, RunAsGroup, RunAsNonRoot, SupplementalGroups

### 0.21.0

* Add Priority Class

### 0.21.1

* Upgrade TFJob to kubeflow.org/v1

### 0.22.0

* Support node selector labels and tolerations

### 0.23.0

* Upgrade deployment to apps/v1

### 0.24.0

* Support adding configuration files when submitting jobs

### 0.25.0

* Fix annotations issue


### 0.26.0

* chief and evaluator should have replicas

### 0.27.0

* PS cpu resource variable should be psCPU

### 0.28.0

* PS annotations should be a attribute of metadata yaml node.

### 0.29.0

* support assgining gpu resources for PS when submitting tfjobs

### 0.30.0

* Support imagePullSecrets

### 0.31.0

* Support job level annotation

### 0.32.0

* Support job level labels

### 0.33.0

* Fix: disable nvidia env for none gpu request job

### 0.34.0

* add --shell to specify sh or bash

### 0.35.0

* Support gang scheduling

### 0.35.1

* add evaluator to pod group

### 0.36.0

* add resource limit to pod

### 0.37.0

* change image repo from kube-ai to acs

### 0.38.0

* Support activeDeadlineSeconds,startingDeadlineSeconds

### 0.39.0

* Support TTLSecondsAfterFinished
