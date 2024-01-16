# Kubeflow Arena Roadmap

## Kubeflow Arena 2024 Roadmap

This document defines a high level roadmap for Arena development.

* Objective：Simplify the user experience by deeply integrating with the Kubeflow Ecosystem.
    * Kubeflow Integration
        * Prepare Arena for release v1.0.0 alongside kubeflow v1.10.
        * Develop a seamless integration with the Training Operator to help simplify model training using command line.
        * Integrate with Kubeflow Pipelines to facilitate model training from a Pipeline.
        * Enable mode serving with KServe.
        * Add documentation to Kubeflow website:
            * Installation, uninstallation, and upgrade processes.
            * Guide for tfjob, mpijob, pytorchJob.

* Objective：Amplify the Extensibility of the Arena for Different ML frameworks, AIGC models and platforms.
    * Support DeepSpeed Training Job.
    * Support for submitting and managing llm fine-tuning jobs, like DeepSpeed.
    * Enable model serving for an expanded set of models like Baichuan, LLaMA, ChatGLM, Falcon, and more.
    * Extend platform support to include ARM.
    * Integrate [Fluid project](https://github.com/fluid-cloudnative/fluid) for efficient data management.
    * Add support for Ray Job management with [Kuberay](https://github.com/ray-project/kuberay).

* Objective: Boost Performance and Stability.
    * Regularly publish recommended practices documentation.
    * Enhancements on Arena SDK.
    * Enhance code quality by:
        * Reduce repetitive code.
        * Enhance unit test.
    * Implement automated End-to-End Test:
        * Add integration tests using GitHub Actions.
        * Strive for more than 60% Test Coverage of Supported Features.

## Kubeflow Arena 2019 Roadmap

### Core CUJs

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
