# Quick Start

This guide walks you through submitting and managing your first training job with Arena v2.

## Prerequisites

- A Kubernetes cluster with `kubectl` configured
- Training CRDs installed: PyTorchJob, TFJob, MPIJob (from the [Kubeflow Training Operator](https://github.com/kubeflow/trainer))
- Arena binary installed (see [README](../README.md#installation))

## 1. Verify Your Environment

Check that the required CRDs are installed:

```shell
$ arena check
✓ PyTorchJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ TFJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ MPIJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
  compatible: ✓ (storage version v1 supported by arena)
```

If any CRD is missing, install the corresponding operator on your cluster.

Check your Arena version:

```shell
$ arena version
Arena v2
  Version:    0.1.0
  Git Commit: abc123
  Build Date: 2026-07-01T00:00:00Z
```

## 2. Submit Your First Training Job

Arena v2 uses YAML-first configuration. The example file `examples/v2/quickstart/pytorch-simple.yaml` defines a distributed PyTorch training job with 2 workers:

```yaml
version: 0.1.0
name: pytorch-example
image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
framework:
  name: pytorch
  options:
    nproc_per_node: auto
worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
envs:
  NCCL_DEBUG: INFO
run: |
  python -c "
  import torch, torch.nn as nn
  model = nn.Linear(10, 1)
  optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
  for epoch in range(3):
      x, y = torch.randn(64, 10), torch.randn(64, 1)
      loss = nn.MSELoss()(model(x), y)
      optimizer.zero_grad(); loss.backward(); optimizer.step()
      print(f'Epoch {epoch+1}/3 - loss: {loss.item():.4f}')
  print('Training complete.')
  "
```

Submit the job:

```shell
$ arena job run -f examples/v2/quickstart/pytorch-simple.yaml
Job pytorch-example submitted successfully
```

You can customize values at submit time using `--set`:

```shell
$ arena job run -f examples/v2/quickstart/pytorch-simple.yaml --set worker.replicas=2
Job pytorch-example submitted successfully
```

Use `--dry-run` to inspect the generated CRD without submitting:

```shell
$ arena job run -f examples/v2/quickstart/pytorch-simple.yaml --dry-run
{
  "apiVersion": "kubeflow.org/v1",
  "kind": "PyTorchJob",
  "metadata": {
    "name": "pytorch-example",
    "namespace": "default"
  },
  "spec": {
    "pytorchReplicaSpecs": {
      "Worker": {
        "replicas": 2,
        "template": {
          "spec": {
            "containers": [
              {
                "image": "pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime",
                "name": "pytorch",
                "resources": {
                  "limits": {
                    "cpu": "2",
                    "memory": "8Gi"
                  }
                }
              }
            ]
          }
        }
      }
    }
  }
}
```

## 3. List Jobs

View all training jobs in the current namespace:

```shell
$ arena job list
NAME              STATUS    REPLICAS  AGE
pytorch-example   Running   2/2       1m
```

Use `-o wide` for more details:

```shell
$ arena job list -o wide
NAME             NAMESPACE  STATUS   APIVERSION       FRAMEWORK  GPU  REPLICAS  AGE
pytorch-example  default    Running  kubeflow.org/v1  pytorch    0    2/2       1m
```

JSON and YAML output formats are also available:

```shell
$ arena job list -o json
$ arena job list -o yaml
```

## 4. Get Job Details

View detailed information about a specific job:

```shell
$ arena job get pytorch-example
Name:      pytorch-example
Namespace: default
Status:    Running
Replicas:  2/2
Age:       5m

Pods:
  NAME                    STATUS    IP         NODE
  pytorch-example-worker-0  Running   10.0.0.1   node-1
  pytorch-example-worker-1  Running   10.0.0.2   node-2
```

Add `--details` to include the full YAML configuration:

```shell
$ arena job get pytorch-example --details
```

## 5. View Logs

Stream logs from the training job:

```shell
$ arena job logs pytorch-example
Epoch 1/3 - loss: 1.1032
Epoch 2/3 - loss: 0.9417
Epoch 3/3 - loss: 0.8124
Training complete.
```

Follow logs in real-time:

```shell
$ arena job logs pytorch-example --follow
```

View logs from a specific pod:

```shell
$ arena job logs pytorch-example --pod pytorch-example-worker-0
```

Show only the last N lines:

```shell
$ arena job logs pytorch-example --tail 50
```

## 6. Suspend and Resume

Pause a running job (scales pods to zero while preserving the job object):

```shell
$ arena job suspend pytorch-example
pytorchjob/pytorch-example suspended
```

Resume a suspended job:

```shell
$ arena job resume pytorch-example
pytorchjob/pytorch-example resumed
```

## 7. Clean Up

Delete the job and all associated resources:

```shell
$ arena job delete pytorch-example
pytorchjob/pytorch-example deleted
```

You can also delete by file:

```shell
$ arena job delete -f examples/v2/quickstart/pytorch-simple.yaml
pytorchjob/pytorch-example deleted
```

## Next Steps

- **YAML schema reference**: [yaml-schema.md](yaml-schema.md) — all available fields and options
- **More examples**: [examples/v2/](../examples/v2/) — TensorFlow, MPI, standalone, TensorBoard, mounts
- **Try `--set` overrides**: `arena job run -f examples/v2/quickstart/pytorch-simple.yaml --set worker.replicas=2`
- **Try dry-run**: `arena job run -f examples/v2/quickstart/pytorch-simple.yaml --dry-run`
- **Design document**: [keps/arena-v2/README.md](../keps/arena-v2/README.md) — v2 architecture and migration guide
