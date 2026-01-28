# Frequently Asked Questions (FAQ)

This page provides answers to common questions about Arena. Browse by category or search for your issue.

## Installation

- [Failed to Install Arena on Mac](./installation/failed-install-arena.md) - Troubleshoot macOS installation issues

## Training Jobs

- [No matches for kind "TFJob"](./training/operator-not-deploy.md) - Operator not deployed or not found

## Serving Jobs

### Model Deployment Issues

**Q: My serving job is stuck in PENDING status**
A: Check cluster resources with `arena top node`. Serving jobs require sufficient CPU, memory, and GPU. Increase replicas or adjust resource requests.

**Q: Model fails to load**
A: Verify model path exists: `kubectl exec <pod> -- ls -la <model-path>`. Check pod logs: `arena serve logs <job-name>`

## Common Issues

### Resource Constraints

**Q: "No resources available" error when submitting jobs**
A: Use `arena top node` to check available resources. Either:

- Wait for other jobs to complete
- Use fewer GPUs in your submission
- Submit to a different namespace with dedicated resources

### Kubeconfig and Authentication

**Q: Arena cannot connect to cluster**
A:

1. Verify kubeconfig: `kubectl cluster-info`
2. Set explicit path: `arena --config=/path/to/kubeconfig list`
3. Check KUBECONFIG env: `echo $KUBECONFIG`

**Q: "Permission denied" when submitting jobs**
A: Your kubeconfig user doesn't have required permissions. Contact cluster admin to grant:

- `create tfjobs`, `pytorchjobs`, etc.
- `get pods` and `logs`
- `create services` and `deployments`

### Job Status Issues

**Q: Job status shows FAILED immediately**
A:

1. Check logs: `arena logs <job-name>`
2. Check pod events: `kubectl describe pod <pod-name>`
3. Common causes: bad image, invalid command, missing data volumes

**Q: Job stuck in PENDING for long time**
A:

1. Check node status: `kubectl get nodes`
2. Check pod events: `kubectl describe pod <pod-name> -n <namespace>`
3. Might be waiting for:
   - GPU availability
   - Image pull
   - Scheduler decisions

### Data and Volume Issues

**Q: Cannot mount data volume**
A:

1. Verify volume exists: `kubectl get pvc`
2. Check access mode: needs ReadWriteOnce or ReadWriteMany
3. Verify data path: `kubectl get pvc <name> -o yaml`

**Q: "No space left on device" error**
A: Volume is full. Either:

- Increase volume size
- Clean up old data
- Use a different volume

### GPU Issues

**Q: GPU not detected in job**
A:

1. Check GPU driver: `kubectl get nodes -L nvidia.com/gpu`
2. Verify node has GPU: `kubectl describe node <node-name>`
3. Check GPU allocation: `arena top node`

**Q: GPU out of memory**
A:

1. Reduce batch size in training script
2. Reduce model size
3. Enable gradient checkpointing
4. Use multiple GPUs with distributed training

## Model Management

**Q: Cannot connect to MLflow tracking server**
A:

1. Verify server running: `curl http://<server>:5000/api/2.0/mlflow/version`
2. Set MLFLOW_TRACKING_URI: `export MLFLOW_TRACKING_URI=http://server:5000`
3. Check network connectivity from Arena pod

**Q: Model not appearing in registry**
A:

1. Verify job completed: `arena get <job-name>`
2. Check model submission flags: `--model-name` and `--model-source`
3. Check model directory permissions and contents

## Monitoring and Logging

**Q: Cannot see job logs**
A:

1. Job might not have started: `arena get <job-name>`
2. Check pod status: `kubectl get pods`
3. Use `--follow` flag: `arena logs <job-name> --follow`

**Q: "arena top" shows no GPU usage**
A:

1. Prometheus might not be running: `kubectl get pods -n arena-system`
2. GPU operator might not be installed
3. Jobs might be waiting for resources

## CLI Issues

**Q: "command not found: arena"**
A:

1. Verify installation: `which arena`
2. Check PATH: `echo $PATH`
3. Reinstall: `cp /usr/local/bin/arena to PATH`

**Q: Autocompletion not working**
A:

1. Source completion: `source <(arena completion bash)`
2. Add to ~/.bashrc: `echo "source <(arena completion bash)" >> ~/.bashrc`
3. Reload shell: `source ~/.bashrc`

## Performance and Optimization

**Q: Training is very slow**
A:

1. Check GPU utilization: `arena top job <job-name>`
2. Check network: distributed training requires good network
3. Use faster data loading: prefetch, cache, smaller batch

**Q: High latency in serving**
A:

1. Check pod load: `arena top job <serving-job>`
2. Increase replicas: `arena serve delete` + resubmit with more replicas
3. Check model complexity and input size

## Troubleshooting Workflow

When something goes wrong, follow this checklist:

1. **Check Job Status**: `arena get <job-name>`
2. **View Logs**: `arena logs <job-name> --tail 100`
3. **Inspect Pod**: `kubectl describe pod <pod-name>`
4. **Check Resources**: `arena top node` and `arena top job <job-name>`
5. **Verify Configuration**: `arena get <job-name>` shows full config
6. **Check Events**: `kubectl get events -n <namespace>`
7. **Review Operator Logs**: `kubectl logs -n arena-system deployment/<operator>`

## Getting More Help

- **Search Documentation**: Check [Training Guide](../training/index.md), [Serving Guide](../serving/index.md)
- **Check CLI Help**: `arena <command> --help`
- **View Examples**: Check [Training Examples](../training/tfjob/standalone.md)
- **GitHub Issues**: [Arena Issues](https://github.com/kubeflow/arena/issues)
- **GitHub Discussions**: [Arena Discussions](https://github.com/kubeflow/arena/discussions)

## Performance Tuning

### Training Performance

1. **Enable GPU**: Always use `--gpus` flag
2. **Increase Batch Size**: Larger batches = better GPU utilization
3. **Distributed Training**: Use multiple nodes/GPUs for large models
4. **Data Pipeline**: Optimize data loading with prefetching
5. **Mixed Precision**: Reduce memory and accelerate training

### Inference Performance

1. **Model Quantization**: Reduce model size with INT8/FP16
2. **Batch Inference**: Process multiple requests together
3. **Caching**: Cache model predictions when possible
4. **Load Balancing**: Distribute traffic across replicas
5. **GPU Memory**: Monitor and optimize GPU allocation

## Related Resources

- [Installation Guide](../installation/index.md)
- [Training Jobs Guide](../training/index.md)
- [Model Serving Guide](../serving/index.md)
- [CLI Reference](../cli/arena.md)
- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug-application-cluster/)
- [MLflow Documentation](https://mlflow.org/docs/latest/index.html)
