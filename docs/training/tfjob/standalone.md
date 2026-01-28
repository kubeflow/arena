# Submit a Standalone TensorFlow Training Job

This guide shows how to submit a standalone (single-node, single-GPU) TensorFlow training job using Arena.

## Prerequisites

Before proceeding, ensure you have:

- Arena installed and configured (see [Installation Guide](../../installation/index.md))
- Access to a Kubernetes cluster with GPUs
- A compatible TensorFlow Docker image
- Basic familiarity with `arena` commands (see [Training Guide](../index.md))

## Overview: What You'll Learn

This guide covers:

1. Checking available GPU resources
2. Submitting a TensorFlow training job
3. Monitoring job status and logs
4. Deleting completed jobs

## Step 1: Check Available Resources

Before submitting a job, verify that your cluster has available GPUs:

```bash
arena top node
```

Expected output:

```
NAME                       IPADDRESS      ROLE    STATUS  GPU(Total)  GPU(Allocated)
cn-hongkong.192.168.2.107  47.242.51.160  <none>  Ready   0           0
cn-hongkong.192.168.2.108  192.168.2.108  <none>  Ready   1           0
cn-hongkong.192.168.2.109  192.168.2.109  <none>  Ready   1           0
cn-hongkong.192.168.2.110  192.168.2.110  <none>  Ready   1           0
------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
0/3 (0.0%)
```

This shows 3 nodes with GPU availability. If you see "GPU(Allocated)" all used, wait for existing jobs to complete or submit to a different namespace.

## Step 2: Submit the TensorFlow Job

Submit a training job that downloads source code from GitHub:

```bash
arena submit tfjob \
    --name=tf-standalone-test \
    --gpus=1 \
    --image=tensorflow/tensorflow:2.9-gpu \
    --sync-mode=git \
    --sync-source=https://github.com/happy2048/tensorflow-sample-code.git \
    --env=TEST_TMPDIR=/root/code/tensorflow-sample-code/ \
    "python /root/code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 5000"
```

Expected output:

```
configmap/tf-standalone-test created
configmap/tf-standalone-test labeled
tfjob.kubeflow.org/tf-standalone-test created
INFO[0000] The Job tf-standalone-test has been submitted successfully
INFO[0000] You can run `arena get tf-standalone-test --type tfjob` to check the job status
```

### Command Options Explained

| Option | Description |
|--------|-------------|
| `--name` | Job name for identification |
| `--gpus=1` | Number of GPUs (1 for standalone) |
| `--image` | Docker image containing TensorFlow |
| `--sync-mode=git` | Sync source code from git repository |
| `--sync-source` | Git repository URL |
| `--env` | Environment variables |
| Last argument | Training command to execute |

### Important Notes

- **Shell**: By default, commands execute with `sh`. To use `bash`, add `--shell=bash`
- **Working Directory**: Code downloads to `/root/code/` by default. Change with `--working-dir`
- **Git Branch**: Specify branch with `--env GIT_SYNC_BRANCH=main`
- **Private Repositories**: Use credentials with `--env GIT_SYNC_USERNAME=<user> --env GIT_SYNC_PASSWORD=<password>`

### Alternative: Local Code

If your code is already in a container image:

```bash
arena submit tfjob \
    --name=tf-training \
    --gpus=1 \
    --image=my-registry/my-tensorflow-image:latest \
    "python /workspace/train.py"
```

## Step 3: Monitor Job Status

### List All TensorFlow Jobs

```bash
arena list --type tfjob
```

Output:

```
NAME                  STATUS    TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
tf-standalone-test    PENDING   TFJOB    3s        1               0               N/A
```

### Job Status Values

| Status | Meaning |
|--------|----------|
| `PENDING` | Waiting for resources or image pull |
| `RUNNING` | Training in progress |
| `SUCCEEDED` | Training completed successfully |
| `FAILED` | Training job failed |

### Get Detailed Job Information

```bash
arena get tf-standalone-test
```

Output:

```
Name:        tf-standalone-test
Status:      RUNNING
Namespace:   default
Priority:    N/A
Trainer:     TFJOB
Duration:    2m30s

Instances:
NAME                              STATUS    AGE  IS_CHIEF  GPU(Requested)  NODE
----                              ------    ---  --------  --------------  ----
tf-standalone-test-chief-0        Running   2m   true      1               192.168.1.100
```

### Monitor Resource Usage

```bash
arena top job tf-standalone-test
```

This shows GPU memory and compute utilization. Low utilization might indicate:

- Data loading bottleneck
- CPU-bound preprocessing
- Small batch size

## Step 4: View Training Logs

### Tail Recent Logs

```bash
arena logs tf-standalone-test --tail 20
```

Expected output:

```
Accuracy at step 4910: 0.9820
Accuracy at step 4920: 0.9828
Accuracy at step 4930: 0.9823
Accuracy at step 4940: 0.9827
Accuracy at step 4950: 0.9824
Accuracy at step 4960: 0.983
Accuracy at step 4970: 0.979
Accuracy at step 4980: 0.9821
Accuracy at step 4990: 0.9823
Adding run metadata for 4999
Total Train-accuracy=0.9823
```

### Stream Live Logs

```bash
arena logs tf-standalone-test --follow
```

This shows new log lines as they appear (similar to `tail -f`).

### View in Web UI

For more detailed visualization:

```bash
arena logviewer tf-standalone-test
```

Output:

```
Your LogViewer will be available on:
172.20.0.197:8080/tfjobs/ui/#/default/tf-standalone-test
```

Open this URL in your browser to see:

- Training progress graphs
- Resource utilization charts
- Pod events and logs

## Step 5: Delete the Job

Once training is complete, clean up:

```bash
arena delete tf-standalone-test
```

This removes:

- Training pod
- TensorFlow job
- Associated ConfigMaps

## Troubleshooting

### Job Stuck in PENDING

```bash
# Check resource availability
arena top node

# Check pod events
kubectl describe pod <pod-name> -n default
```

**Solutions:**

- Wait for other jobs to complete
- Use fewer GPUs (`--gpus=1`)
- Reduce CPU/memory requests

### Image Pull Error

```bash
# Try alternative image
arena submit tfjob \
    --name=tf-training \
    --gpus=1 \
    --image=registry.cn-hongkong.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu \
    "python train.py"
```

### No Logs Appearing

```bash
# Check job status first
arena get tf-standalone-test

# If PENDING, wait for RUNNING state
# If RUNNING but no logs, check pod directly
kubectl logs <pod-name> -n default
```

### Out of Memory

```bash
# Reduce batch size in your training script
# Or use mixed precision:
```

In your training code:

```python
# TensorFlow 2.x mixed precision
from tensorflow.keras import mixed_precision
policy = mixed_precision.Policy('mixed_float16')
mixed_precision.set_global_policy(policy)
```

## Next Steps

1. **Distributed Training**: Scale to multiple nodes with [Distributed TensorFlow](./distributed.md)
2. **Monitor GPU Usage**: Track resource utilization with [Resource Monitoring](../../top/index.md)
3. **Deploy Model**: Serve your trained model using [Model Serving](../../serving/tfserving/serving.md)
4. **Register Model**: Track model versions with [Model Management](../../model/index.md)

## Related Guides

- [TensorFlow with TensorBoard](./tensorboard.md) - Visualize training
- [TensorFlow with Datasets](./dataset.md) - Manage training data
- [Distributed TensorFlow](./distributed.md) - Multi-GPU/multi-node training
- [Training Job Operations](../common/list_jobs.md) - Manage training jobs

## Advanced Options

### Using Custom Docker Registry

```bash
arena submit tfjob \
    --name=tf-training \
    --gpus=1 \
    --image=private-registry.example.com/tensorflow:latest \
    --image-pull-policy=IfNotPresent \
    "python train.py"
```

### Specifying Resource Requests

```bash
arena submit tfjob \
    --name=tf-training \
    --gpus=1 \
    --cpu=4 \
    --memory=16Gi \
    --image=tensorflow/tensorflow:2.9-gpu \
    "python train.py"
```

### Environment Variables and Configuration

```bash
arena submit tfjob \
    --name=tf-training \
    --gpus=1 \
    --env=LEARNING_RATE=0.001 \
    --env=BATCH_SIZE=32 \
    --env=MAX_STEPS=10000 \
    --image=tensorflow/tensorflow:2.9-gpu \
    "python train.py"
```

## Congratulations

You've successfully submitted and monitored your first TensorFlow training job with Arena! Next, try distributed training or deploy your model for serving.
