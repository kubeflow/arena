### 0.14.0
* add annotations, for cloud provider customization

### 0.13.0
* Support RoCE by using https://github.com/Mellanox/k8s-rdma-sriov-dev-plugin, only support hostNetwork now.

### 0.12.0
* support standalone training without sshd requirement

### 0.11.0

* disable privileged by default


### 0.10.0

* support multiple dataDirs 

### 0.9.0

* Add shmSize and privileged

### 0.8.0

* Add dig

### 0.7.0

* retry time > 300
* sleep 1 seconds before checking ssh

### 0.6.0

* fix nvidia path

### 0.5.0

* support rsync init container, working dir and merge device plugin and hard code GPU

### 0.4.0

* support pvc list

### 0.3.0

* Add job monitor

### 0.2.0

* Master also participates computing

### 0.1.0

* Master is a job, and the workers are statefulset