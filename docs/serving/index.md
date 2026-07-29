# Model Serving Management Guide

Welcome to the Arena Model Serving Guide! This comprehensive guide covers how to use the Arena CLI to deploy, manage, and monitor inference services on Kubernetes.

## Overview

Arena simplifies model serving by providing a unified CLI for deploying and managing inference services. You can:

- Deploy trained models for inference
- Support multiple serving frameworks
- Manage model versions and traffic routing
- Monitor inference service performance
- Scale serving instances automatically

## Who Should Use This Guide?

This guide is for you if you want to:

- Deploy trained models for production inference
- Manage multiple model versions
- Route traffic between model versions
- Monitor serving infrastructure
- Integrate model serving with training pipelines

## Quick Start: Deploy Your First Model

Deploy a TensorFlow model in minutes:

```bash
# Assuming you have a trained model in a PVC
# Deploy the model
arena serve tensorflow \
  --name=my-model-serving \
  --model-name=my-model \
  --data=model-storage:/model \
  --model-path=/model

# Monitor the serving job
arena serve list
arena serve get my-model-serving
arena serve logs my-model-serving
```

## Common Serving Operations

Start with these essential operations:

- [List serving jobs](./common/list_jobs.md) - View all deployed models.
- [Get serving job details](./common/get_job.md) - Inspect serving configuration and status.
- [Get serving job logs](./common/get_job_logs.md) - View serving container logs.
- [Attach to serving job](./common/attach_job.md) - Connect to serving container for debugging.
- [Delete serving jobs](./common/delete_jobs.md) - Undeploy serving jobs.

## Serving Frameworks

Choose the serving framework that fits your needs.

### TensorFlow Serving (Production-Grade Model Serving)

Highly optimized C++ server for serving TensorFlow models with gRPC and REST APIs.

**Best For:**

- TensorFlow models (SavedModel format).
- Production high-throughput inference.
- Model versioning and canary deployments.
- Multi-model serving.

**Features:**

- Model version policy control.
- Dynamic model loading.
- gRPC and REST APIs.
- Prometheus metrics.
- Model warming.

**Getting Started:**

- [Basic TensorFlow Serving](./tfserving/serving.md) - Deploy a TensorFlow model.
- [GPU Support](./tfserving/gpu.md) - Use GPUs for inference.
- [GPU Sharing Mode](./tfserving/gpushare.md) - Share GPUs across models.
- [Prometheus Monitoring](./tfserving/monitor.md) - Monitor serving metrics.
- [Update Serving Job](./tfserving/update-serving.md) - Modify deployed models.

### NVIDIA Triton Inference Server (Multi-Framework)

Flexible inference server supporting TensorFlow, PyTorch, ONNX, and custom models.

**Best For:**

- Multi-framework model serving.
- Custom inference logic.
- Ensemble models.
- Advanced batching strategies.

**Features:**

- Support for TensorFlow, PyTorch, ONNX formats.
- Ensemble models.
- Dynamic batching.
- Model broadcasting.
- Custom backend support.

**Getting Started:**

- [Triton Serving](./triton/serving.md) - Deploy models with Triton.
- [Update Triton Serving](./triton/update-serving.md) - Modify deployments.

### KServe (Kubernetes-Native Model Serving)

Kubernetes-native serving system with support for multiple frameworks and protocols.

**Best For:**

- Kubernetes-native deployments.
- AutoML and explainability.
- Batch prediction.
- Distributed inference.

**Features:**

- Out-of-the-box predictors (Scikit-learn, XGBoost, LightGBM, etc.).
- Model transformers.
- Explainability (SHAP, Integrated Gradients).
- Auto-scaling.
- Request/response logging.

**Getting Started:**

- [Scikit-learn Models](./kserve/sklearn.md) - Deploy Scikit-learn models.
- [Custom Models](./kserve/custom.md) - Deploy custom inference code.

### KFServing (Kubeflow Model Serving)

Kubeflow's model serving framework with advanced ML serving capabilities.

**Best For:**

- Kubeflow ecosystems.
- Advanced feature engineering.
- A/B testing.
- Canary deployments.

**Getting Started:**

- [Custom KFServing](./kfserving/custom.md) - Deploy custom models.

### TensorRT Serving (NVIDIA Optimization)

NVIDIA's inference optimization framework for maximum GPU performance.

**Best For:**

- NVIDIA GPU optimization.
- Low-latency requirements.
- High throughput inference.
- Model optimization and quantization.

**Getting Started:**

- [TensorRT Serving](./arena_serve_tensorrt.md) - Deploy TensorRT models.

### Custom Serving (Your Own Framework)

Deploy custom inference services using any framework or programming language.

**Best For:**

- Custom inference logic.
- Non-standard models.
- Framework not listed above.
- Specialized preprocessing/postprocessing.

**Features:**

- Use any Docker image.
- Custom ports and protocols.
- Full control over inference logic.
- Environment variable configuration.

**Getting Started:**

- [Custom Serving](./customserving/gpu.md) - Deploy with GPU support.
- [Update Custom Serving](./customserving/update-serving.md) - Modify deployments.

### Seldon Core (Open-Source Model Serving)

Open-source serving platform with advanced ML serving capabilities.

**Best For:**

- Open-source deployments.
- Model ensembles.
- Outlier detection.
- A/B testing and canary deployments.

**Getting Started:**

- [Pre-packaged Model Server](./seldon-core/pre-packaged-model-server.md) - Use Seldon pre-packaged servers.

### Distributed Serving (Multi-Node)

Deploy models across multiple nodes for high availability and scalability.

**Best For:**

- High-availability requirements.
- Large model serving.
- Multi-region deployments.
- Load-balanced inference.

**Getting Started:**

- [Distributed Serving](./distributedserving/serving.md) - Deploy across multiple nodes.

## Advanced Features

### Model Versioning and Routing

- Deploy multiple model versions.
- Gradual traffic migration (canary deployments).
- Version-specific routing.
- Model rollback capabilities.

### Traffic Management

- [Traffic Splitting](./common/arena_serve_traffic-split.md) - Split traffic between versions.
- Weight-based routing.
- Shadow traffic routing.
- A/B testing support.

### GPU and Resource Management

- Single and multi-GPU serving.
- GPU memory allocation.
- CPU and memory limits.
- Resource quotas.

### Monitoring and Observability

- Real-time serving logs.
- Prometheus metrics integration.
- Model performance tracking.
- Inference latency monitoring.

### Data Management

- Mount persistent volumes for models.
- Cloud storage integration (S3, GCS).
- Model artifact management.
- Version control integration.

## Workflow Examples

### Example 1: Simple Model Serving

```bash
# Deploy a single model version
arena serve tensorflow \
  --name=mnist-serving \
  --model-name=mnist \
  --data=models:/models \
  --model-path=/models/mnist/1

# Check status
arena serve list
arena serve get mnist-serving
```

### Example 2: Multi-Version Deployment with Traffic Splitting

```bash
# Deploy v1
arena serve tensorflow \
  --name=mnist-serving \
  --version=v1 \
  --model-name=mnist \
  --data=models:/models \
  --model-path=/models/mnist/1

# Deploy v2 (new version)
arena serve tensorflow \
  --name=mnist-serving \
  --version=v2 \
  --model-name=mnist \
  --data=models:/models \
  --model-path=/models/mnist/2

# Route 80% traffic to v1, 20% to v2 (canary)
arena serve traffic-split \
  --name=mnist-serving \
  --version=v1:80 \
  --version=v2:20
```

### Example 3: Custom Serving with GPUs

```bash
# Deploy custom inference service with GPU
arena serve custom \
  --name=my-inference-service \
  --image=my-org/inference:latest \
  --gpus=1 \
  --replicas=2 \
  --restful-port=8080 \
  --data=models:/models
```

## Common Patterns

### Using Model Registry

```bash
# Reference a model from model registry
arena serve tensorflow \
  --name=serving-job \
  --model-name=my-model \
  --model-version=1 \
  --data=models:/models
```

### Multi-Model Serving

```bash
# Deploy multiple models with a single command
arena serve custom \
  --name=multi-model-serving \
  --image=triton-inference-server \
  --data=models:/models
```

### Monitoring Serving Performance

```bash
# Check serving job status
arena serve get my-serving-job

# View serving logs
arena serve logs my-serving-job --tail 100

# Monitor resource usage
arena top job my-serving-job
```

## Next Steps

1. **Start Simple**: Deploy [TensorFlow Serving](./tfserving/serving.md).
2. **Add Monitoring**: Set up [Prometheus Monitoring](./tfserving/monitor.md).
3. **Scale Up**: Deploy [Multi-Node Serving](./distributedserving/serving.md).
4. **Advanced Routing**: Implement [Traffic Splitting](./common/arena_serve_traffic-split.md).
5. **Integrate**: Link with [Training Jobs](../training/index.md).

## Troubleshooting

### Model Not Loading

```bash
# Check serving logs
arena serve logs my-serving-job

# Verify model path
kubectl exec <serving-pod> -- ls -la /model
```

### Out of Memory

```bash
# Check resource usage
arena top job my-serving-job

# Scale down replicas or increase memory
arena serve delete my-serving-job
arena serve tensorflow --name=my-serving --replicas=1 ...
```

### Inference Timeout

```bash
# Check serving pod logs
arena serve logs my-serving-job --tail 50

# Monitor serving metrics
kubectl port-forward svc/my-serving-job 8000:8000
```

## See Also

- [CLI Reference](../cli/arena.md)
- [Training Jobs Guide](../training/index.md)
- [Model Management](../model/index.md)
- [Monitoring Guide](../top/index.md)
- [FAQ & Troubleshooting](../faq/index.md)
