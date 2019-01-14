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

* Avoid PS to use GPU during training when the PS image is CUDA images for nvidia