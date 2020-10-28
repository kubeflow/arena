## [Release 0.4.0]

### Added

- Add Pytorch
- Support custom serving
- Add tarball installation for Linux and Mac
- Support non-root installation
- Add train init framework
- Add GPU support for PS
- Support GangScheduling Native in MPIJob
- Supoort GangScheduling Native in PytorchJob

### Fixed

- Upgrade Deployment version from extensions/v1beta1 to apps/v1
- Fix the issue of incorrect number of allocated GPUs
- Upgrade Helm to v2.14.1
- Fix evaluator & chief validation
- Fix incorrect cpu resource variable, should be psCPU
- Set exit code as 2 when delete job failed
- Fix the bug of using Estimator
- Fix the bug of deploying Prometheus
- Support Kubernetes 1.18 and above

## [Release 0.3.0]

### Added

- Add Priority class support for MPIJob and TFJob
- Display Unhealthy GPU devices
- Integrate GPUShare capablities
- Upgrade TFJob to V1 (commit id: d746bde)
- Add Customize Serving
- Add GPUsharing features for Serving Job

## [Release 0.2.0]

### Added

- Add spark and volcano Job
- Add multiple users and add PodSecurityContext for Training Job
- Add TensorRT

### Changed

- Refactoring code to remove dependency of helm create
- Enhance cluster management

## [Release 0.1.0]

### Added

- Add TFJob v1alpha2 for Solo/Distributed Training, and support binpack and spread mode
- Add Download Source Code from Git for Training
- Add Tensorboard
- Add top node/job for checking GPU allocations in Kubernetes
- Add MPIJob v1alpha1 for Solo/Distributed Training
- Add gang scheduling support for TFJob
- Add Data
- Add RDMA support

### Changed

### Removed

### Fixed

### Deprecated

- HorovodJob is going to remove when MPIJob is production ready
