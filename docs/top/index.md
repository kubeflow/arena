# Resource Monitoring and Usage Guide

Welcome to the Arena Resource Monitoring Guide! This guide covers how to use the `arena top` command to monitor GPU resources, cluster utilization, and job performance.

## Overview

Monitoring is critical for:

- **Resource Planning** - Understand cluster capacity and utilization
- **Debugging** - Identify performance bottlenecks
- **Optimization** - Improve training and serving efficiency
- **Troubleshooting** - Diagnose job failures

Arena provides two main monitoring commands:

- `arena top node` - View cluster node resources and GPU allocation
- `arena top job` - View individual job resource usage

## Quick Start

### Monitor Node Resources

```bash
# View all nodes and their GPU resources
arena top node

# View specific node details
arena top node <node-name>
```

### Monitor Job Resources

```bash
# View resource usage of a specific job
arena top job <job-name>

# View GPU memory usage
arena top job <job-name> --gpu-memory
```

## Usage Guides

This guide covers three main parts:

- [Node Resource Monitoring](./top_node.md) - Monitor cluster node resources
- [Job Resource Monitoring](./top_job.md) - Monitor training and serving job resource usage
- [Prometheus Integration](./prometheus.md) - Advanced monitoring with Prometheus metrics

## Who Should Use This Guide?

This guide is for you if you:

- Want to understand cluster GPU availability
- Need to debug slow training jobs
- Require detailed performance metrics
- Are optimizing resource utilization
- Need to set up monitoring dashboards

## Key Metrics

### Node Level Metrics

- **GPU Count** - Total GPUs per node
- **GPU Memory** - Available and allocated GPU memory
- **Allocation Status** - Whether GPUs are free or allocated
- **GPU Topology** - GPU interconnect topology (relevant for distributed training)
- **Temperature** - GPU temperature (if available)

### Job Level Metrics

- **GPU Usage** - Which GPUs are used
- **Memory Usage** - GPU memory consumption
- **GPU Utilization** - Percentage of GPU compute capacity used
- **Power Consumption** - GPU power draw
- **Execution Time** - Job duration and time in PENDING state

## Common Monitoring Scenarios

### Scenario 1: Check If GPUs Are Available

```bash
# View GPU availability
arena top node

# Check if specific node has free GPUs
arena top node <node-name>
```

**What to look for:**

- "Free" vs "Allocated" GPU count
- Free GPU memory
- If all GPUs allocated, you may need to wait or add more nodes

### Scenario 2: Debug Slow Training

```bash
# Check job resource usage
arena top job <job-name>

# Look for:
# - Low GPU utilization (< 80%) indicates code/data bottleneck
# - High GPU memory but low compute indicates memory bottleneck
# - Alternating high/low indicates data loading issue
```

### Scenario 3: Optimize Multi-GPU Training

```bash
# Monitor distributed training job
arena top job <distributed-job-name>

# Check:
# - Are all GPUs being used equally?
# - Is there GPU-to-GPU communication overhead?
# - Is one GPU lagging behind others?
```

### Scenario 4: Monitor Serving Deployment

```bash
# Check serving job resource usage
arena top job <serving-job-name>

# Monitor:
# - GPU memory baseline (model loading)
# - Memory growth over time (memory leak?)
# - GPU utilization during requests
```

## Monitoring Best Practices

### 1. Regular Monitoring

```bash
# Set up periodic monitoring
watch -n 5 'arena top node'

# Or in separate terminal
arena top node --follow  # if available
```

### 2. Correlate Metrics

When investigating issues, check both:

```bash
arena list  # Job status
arena top job <job-name>  # Resource usage
arena logs <job-name> --tail 20  # Job logs
```

### 3. Baseline Establishment

Understand normal resource usage:

```bash
# Record baseline for healthy job
arena top job <known-good-job>

# Compare with problematic job
arena top job <problematic-job>
```

### 4. Alerting

Set up alerts based on metrics:

- GPU memory not decreasing (potential memory leak)
- GPU utilization < 50% for extended period (inefficient job)
- Long PENDING time (resource contention)

## Advanced Monitoring with Prometheus

For production deployments, use Prometheus for:

- Historical metrics
- Custom alerts
- Dashboards (Grafana)
- Metrics aggregation

See [Prometheus Integration](./prometheus.md) for detailed setup.

## Troubleshooting Monitoring

### "arena top" shows no data

```bash
# Check if monitoring is enabled
kubectl get pods -n arena-system

# Check Prometheus status (if using Prometheus backend)
kubectl get pods -n monitoring

# Jobs might be too new, wait a few seconds
arena top job <new-job-name>
```

### GPU metrics not available

```bash
# Verify GPU driver
kubectl get nodes -L nvidia.com/gpu

# Check node GPU availability
arena top node

# If showing "N/A", GPU driver might not be installed
```

## Performance Tuning Based on Metrics

### High Memory Usage

1. Reduce batch size
2. Enable gradient checkpointing
3. Use mixed precision (FP16)
4. Switch to a smaller model

### Low GPU Utilization

1. Increase batch size
2. Optimize data pipeline (prefetch, cache)
3. Review code for inefficiencies
4. Check for CPU bottlenecks

### Uneven GPU Usage Across Nodes

1. Check network bandwidth between nodes
2. Verify MPI/collective operation efficiency
3. Check for node performance imbalances
4. Verify data is balanced across workers

## Next Steps

1. **Learn Node Monitoring**: [Node Resource Guide](./top_node.md)
2. **Learn Job Monitoring**: [Job Resource Guide](./top_job.md)
3. **Set up Prometheus**: [Prometheus Integration](./prometheus.md)
4. **Optimize Training**: Use metrics to improve [Training Jobs](../training/index.md)
5. **Optimize Serving**: Use metrics to improve [Model Serving](../serving/index.md)

## Related Resources

- [Training Jobs Guide](../training/index.md)
- [Model Serving Guide](../serving/index.md)
- [CLI Reference](../cli/arena.md)
- [Kubernetes Metrics](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-metrics-pipeline/)
