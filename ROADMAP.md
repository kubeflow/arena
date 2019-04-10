# Arena 2019 Roadmap

This document defines a high level roadmap for Arena development.

### 2019

#### Core CUJs

Objectives: "Make Arena easily to be integrated with External System."

* Integration
	* Provide the Java, Python and C++ API for system interaction
	* Integrate with Pipelines to support Standalone Job, MPI Job, Estimator Job 

Objectives: "Simplify the user experience of the data scientists and provide a low barrier to handle different kind of  training jobs and serve different models."

* High Level Interfaces:
	* Submit and manage other data processing and machine learning jobs, like Spark, Flink, [XDL](https://github.com/alibaba/x-deeplearning), PyTorch, MXNet
	* Submit and manage Model Serving with [KF Serving](https://github.com/kubeflow/kfserving)


Objectives: "Make Arena support the same Operator compatiable with different API version, so the upgrade of operator doesn't impact the existing users' experiences."

* Compatibility:
	* v1aphla2 and v1 TFJob
	* v1alpha1 and v1aphla2 MPIJob

Objectives: "Enchance the software quality of Arena so it can be in the quick iteration"

* Refactor the source code
	* Move Training implementation from `cmd` into `pkg`

* Automatic Test Enhancement: 
	* Unit test
	* Integration test
