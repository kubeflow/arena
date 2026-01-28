# Arena

**Arena** is a command-line interface (CLI) designed for data scientists to efficiently manage machine learning workloads on Kubernetes clusters. It abstracts complex Kubernetes concepts, allowing users to focus on training and serving models without deep Kubernetes expertise.

## Key Features

### Simplified Training Management

Support for multiple training frameworks and distributed training orchestration:

- **TensorFlow** - Single and distributed training with TensorFlow jobs.
- **PyTorch** - Distributed training and elastic training support.
- **MPI** - High-performance computing with MPI jobs.
- **Spark** - Distributed data processing with Spark.
- **Ray** - Distributed machine learning with Ray.
- **Elastic Training** - Fault-tolerant distributed training that scales dynamically.
- **Horovod** - Distributed deep learning training.
- **Volcano** - High-performance computing workloads.

### Model Serving & Inference

Deploy and manage inference services:

- **TensorFlow Serving** - Production-grade model serving.
- **NVIDIA Triton** - Multi-framework inference server.
- **KServe** - Kubernetes-native model serving.
- **KFServing** - Kubeflow model serving framework.
- **Custom Serving** - Deploy custom inference services.
- **Distributed Serving** - Multi-node inference deployments.

### Resource Management

- **GPU Resource Monitoring** - Real-time GPU utilization tracking.
- **Node Management** - View and manage cluster resources.
- **Auto-scaling** - Scale training jobs in and out dynamically.
- **Multiple Users** - Multi-tenant support with namespace isolation.

### Model Lifecycle

- **Model Registration** - Track model versions and metadata.
- **Model Tagging** - Organize models with custom tags.
- **Training Integration** - Automatically register models during training.
- **Serving Integration** - Reference models in serving deployments.

## Platform Support

- **Operating Systems**: Linux and macOS.
- **Kubernetes**: Version 1.11 and above.
- **Frameworks**: TensorFlow, PyTorch, MPI, Spark, Ray, Horovod, and more.

## Getting Help

- Check [Installation Guide](./installation/index.md) for setup issues.
- Review [FAQ & Troubleshooting](./faq/index.md) for common problems.
- See [CLI Reference](./cli/arena.md) for command documentation.
