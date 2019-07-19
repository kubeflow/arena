## [Release 0.3.0]

### Added

- Add Priority class support for MPIJob and TFJob


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