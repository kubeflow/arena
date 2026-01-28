# Training Job Management Guide

Welcome to the Arena Training Job Management Guide! This comprehensive guide covers how to use the Arena CLI to submit, manage, and monitor machine learning training workloads on Kubernetes.

## Overview

Arena simplifies distributed training by abstracting Kubernetes complexity. You can:

- Submit training jobs from the command line
- Monitor job progress in real-time
- Access logs and debugging information
- Scale jobs dynamically
- Manage model outputs and versions

## Who Should Use This Guide?

This guide is for you if you want to:

- Run machine learning training on Kubernetes
- Manage single and distributed training workloads
- Use popular frameworks like TensorFlow, PyTorch, or MPI
- Integrate training with model management

## Quick Start: Your First Training Job

Submit a TensorFlow job in 30 seconds:

```bash
# Submit a training job
arena submit tfjob \
  --name=my-first-job \
  --gpus=1 \
  --image=tensorflow/tensorflow:latest-gpu \
  "python train.py"

# Monitor the job
arena list
arena get my-first-job
arena logs my-first-job
```

## Common Training Operations

Start with these essential operations:

- [List all training jobs](./common/list_jobs.md) - View jobs and their status.
- [Get training job details](./common/get_job.md) - Inspect job configuration and status.
- [Get training job logs](./common/get_job_logs.md) - View training output and debugging info.
- [Attach to a training job](./common/attach_job.md) - Connect to running job for debugging.
- [Delete training jobs](./common/delete_jobs.md) - Remove completed or failed jobs.
- [Clean up finished jobs](./common/prune_jobs.md) - Batch delete old jobs.
- [Image pull secrets](./common/image-pull-secret.md) - Use private Docker registries.

## Training Frameworks

Choose the training framework that fits your needs.

### TensorFlow (Popular for Deep Learning)

TensorFlow is widely used for deep learning, computer vision, and NLP tasks.

**Common Use Cases:**
- Image classification and detection.
- Natural language processing.
- Time series forecasting.
- Transfer learning.

**Getting Started:**
- [Standalone TensorFlow job](./tfjob/standalone.md) - Single machine training.
- [Distributed TensorFlow job](./tfjob/distributed.md) - Multi-machine, multi-GPU training.
- [TensorFlow with TensorBoard](./tfjob/tensorboard.md) - Real-time monitoring and visualization.
- [TensorFlow with datasets](./tfjob/dataset.md) - Manage training data.
- [TensorFlow with estimators](./tfjob/estimator.md) - High-level TensorFlow API.
- [Gang scheduling](./tfjob/gangschd.md) - Ensure all replicas start together.
- [Node selectors](./tfjob/selector.md) - Pin jobs to specific nodes.
- [Node tolerations](./tfjob/toleration.md) - Schedule on tainted nodes.
- [Config files](./tfjob/assign_config_file.md) - Pass configuration files.
- [Role sequence](./tfjob/role-sequence.md) - Control replica startup order.

### PyTorch (Popular for Research)

PyTorch provides flexibility and is favored in research communities.

**Common Use Cases:**
- Computer vision research.
- Natural language processing.
- Reinforcement learning.
- Custom model architectures.

**Getting Started:**
- [Standalone PyTorch job](./pytorchjob/standalone.md) - Single GPU training.
- [Distributed PyTorch job](./pytorchjob/distributed.md) - Data-parallel and model-parallel training.
- [PyTorch with TensorBoard](./pytorchjob/tensorboard.md) - Visualization during training.
- [Distributed data handling](./pytorchjob/distributed-data.md) - Efficient data loading.
- [Node selectors](./pytorchjob/node-selector.md) - Target specific nodes.
- [Node tolerations](./pytorchjob/node-toleration.md) - Use tainted nodes.
- [Config files](./pytorchjob/assign-config-file.md) - Configuration management.
- [Preemption handling](./pytorchjob/preempted.md) - Resume from checkpoints.
- [Clean pod policy](./pytorchjob/clean-pod-policy.md) - Control pod cleanup on completion.

### MPI (High-Performance Computing)

Message Passing Interface is ideal for HPC and distributed algorithms requiring tight synchronization.

**Common Use Cases:**
- Distributed scientific computing.
- Large-scale model training.
- Complex communication patterns.
- RDMA-optimized workloads.

**Getting Started:**
- [Distributed MPI job](./mpijob/distributed.md) - Classic MPI distributed training.
- [GPU topology scheduling](./mpijob/gputopology.md) - Optimize GPU topology for NCCL.
- [Preemption handling](./mpijob/preempted.md) - Checkpoint and resume.
- [Node tolerations](./mpijob/toleration.md) - Schedule on tainted nodes.
- [Node selectors](./mpijob/selector.md) - Select specific nodes.
- [Config files](./mpijob/assign_config_file.md) - Pass MPI configuration.
- [RDMA devices](./mpijob/rdma.md) - Enable RDMA for performance.

### Elastic Training (Dynamic Scaling)

Elastic training automatically adjusts the number of workers during training.

**Common Use Cases:**
- Fault-tolerant training on unreliable infrastructure.
- Cost optimization by adding/removing workers.
- Automatic recovery from node failures.
- Dynamic workload scaling.

**Getting Started:**
- [Elastic PyTorch training](./etjob/elastictraining-pytorch-synthetic.md) - Elastic distributed PyTorch.
- [Elastic TensorFlow training](./etjob/elastictraining-tensorflow2-mnist.md) - Elastic TensorFlow jobs.

### Spark (Big Data Processing)

Apache Spark for distributed data processing and machine learning.

**Common Use Cases:**
- Large-scale data preprocessing.
- Distributed SQL queries.
- Machine learning on big data.
- Batch processing pipelines.

**Getting Started:**
- [Distributed Spark job](./sparkjob/distributed.md) - Multi-node Spark cluster.

### Ray (Distributed ML)

Ray framework for distributed machine learning and reinforcement learning.

**Common Use Cases:**
- Reinforcement learning training.
- Distributed hyperparameter tuning.
- Distributed scikit-learn.
- General distributed ML.

**Getting Started:**
- [Ray training job](./rayjob/rayjob.md) - Distributed Ray cluster.

### Volcano (High-Performance Computing)

Volcano is a Kubernetes-native batch scheduling system optimized for HPC.

**Common Use Cases:**
- High-performance computing workloads.
- Batch scheduling with advanced features.
- Gang scheduling for parallel jobs.
- Fair resource allocation.

**Getting Started:**
- [Volcano training job](./volcanojob/volcanojob.md) - Submit jobs via Volcano.

### Horovod (Distributed Deep Learning)

Horovod provides a simplified distributed deep learning framework.

**Common Use Cases:**
- Easier distributed training setup.
- Scalable distributed training.
- Multi-framework support (TensorFlow, PyTorch, etc.)

**Getting Started:**
- [Horovod training job](./../../docs/cli/arena_submit_horovodjob.md) - Distributed Horovod training.

### Cron Training (Scheduled Jobs)

Automatic training job scheduling at specified times.

**Common Use Cases:**
- Daily model retraining.
- Periodic data processing.
- Automated model updates.
- Scheduled experiments.

**Getting Started:**
- [Cron TensorFlow job](./cron/cron-tfjob.md) - Schedule periodic training.

## Advanced Features

### Data Management

- Mount persistent volumes for training data
- Use cloud storage (S3, GCS, etc.)
- Data synchronization and preprocessing
- Multi-mount data volumes

### GPU and Resource Management

- Single GPU training
- Multi-GPU training (data parallelism)
- GPU memory limits
- CPU and memory requests
- Node affinity and topology awareness

### Monitoring and Debugging

- Real-time job logs
- Log viewer UI
- Pod-level debugging
- Performance profiling
- Resource usage tracking

### Model Lifecycle

- Automatic model registration during training
- Model versioning
- Training metadata tracking
- Integration with model serving

## Workflow Examples

### Example 1: Simple GPU Training

```bash
# Start a single-GPU TensorFlow training
arena submit tfjob \
  --name=simple-training \
  --gpus=1 \
  --image=tensorflow/tensorflow:latest-gpu \
  "python /workspace/train.py"

# Monitor progress
arena logs simple-training --follow
```

### Example 2: Distributed Multi-GPU Training

```bash
# Start 2-node distributed PyTorch training
arena submit pytorchjob \
  --name=distributed-training \
  --gpus=2 \
  --workers=2 \
  --image=pytorch/pytorch:latest \
  "python /workspace/distributed_train.py"

# Check status
arena get distributed-training
```

### Example 3: Training with Data and Model Registration

```bash
# Submit training with data volume and model registration
arena submit pytorchjob \
  --name=mnist-training \
  --gpus=1 \
  --data=training-data:/data \
  --model-name=mnist-model \
  --model-source=pvc://default/models/mnist \
  --image=pytorch/pytorch:latest \
  "python /workspace/train_mnist.py --data-dir /data"
```

## Next Steps

1. **Start Simple**: Try [standalone TensorFlow job](./tfjob/standalone.md)
2. **Scale Up**: Learn [distributed training](./tfjob/distributed.md)
3. **Automate**: Set up [cron jobs](./cron/cron-tfjob.md)
4. **Deploy**: Deploy models using [Model Serving Guide](../serving/index.md)
5. **Optimize**: Monitor with [Top/Resource Usage](../top/index.md)

## Troubleshooting

### Job Stuck in PENDING

```bash
# Check resource availability
arena top node

# Check pod events
kubectl describe pod <pod-name> -n <namespace>
```

### GPU Not Available

```bash
# Verify GPU driver
kubectl get nodes -L nvidia.com/gpu

# Check Arena namespace
kubectl get pods -n arena-system
```

### Out of Memory

```bash
# Monitor memory usage
arena top job <job-name>

# Reduce batch size or model size
```

## See Also

- [CLI Reference](../cli/arena.md)
- [Model Serving Guide](../serving/index.md)
- [Model Management](../model/index.md)
- [Monitoring Guide](../top/index.md)
- [FAQ & Troubleshooting](../faq/index.md)
