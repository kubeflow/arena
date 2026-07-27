# Quick Start

This guide walks you through submitting and managing your first training job with Arena v2.

## Prerequisites

- A Kubernetes cluster with `kubectl` configured
- Training CRDs installed: PyTorchJob, TFJob, MPIJob (from [Kubeflow Training Operator](https://github.com/kubeflow/trainer) and [MPI Operator](https://github.com/kubeflow/mpi-operator))
- Arena binary installed (see [README](../README.md#installation))

## 1. Verify Your Environment

Check that the required CRDs are installed:

```shell
$ arena check
✓ PyTorchJob: installed (expected: v1)
  versions: v1 (served, storage)
✓ TFJob: installed (expected: v1)
  versions: v1 (served, storage)
✓ MPIJob: installed (expected: v2beta1)
  versions: v2beta1 (served, storage)
  compatible: ✓ (storage version v2beta1 supported by arena)
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

Arena v2 uses YAML-first configuration. The example file `examples/v2/pytorch-simple.yaml` defines a distributed PyTorch training job with 4 workers:

```yaml
version: 0.1.0
name: pytorch-example
image: pytorch/pytorch:1.13-cuda11.6-cudnn8-runtime
framework:
  name: pytorch
  options:
    nproc_per_node: auto
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: "1"
    cpu: "2"
    memory: "8Gi"
envs:
  NCCL_DEBUG: INFO
run: python train.py --epochs 10
```

Submit the job:

```shell
$ arena job run -f examples/v2/pytorch-simple.yaml
Job pytorch-example submitted successfully
```

You can customize values at submit time using `--set`:

```shell
$ arena job run -f examples/v2/pytorch-simple.yaml --set worker.replicas=2
Job pytorch-example submitted successfully
```

Use `--dry-run` to inspect the generated CRD without submitting:

```shell
$ arena job run -f examples/v2/pytorch-simple.yaml --dry-run
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
        "replicas": 4,
        "template": {
          "spec": {
            "containers": [
              {
                "image": "pytorch/pytorch:1.13-cuda11.6-cudnn8-runtime",
                "name": "pytorch",
                "resources": {
                  "limits": {
                    "cpu": "2",
                    "memory": "8Gi",
                    "nvidia.com/gpu": "1"
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
pytorch-example   Running   2/4       1m
```

Use `-o wide` for more details:

```shell
$ arena job list -o wide
NAME             NAMESPACE  STATUS   APIVERSION       FRAMEWORK  GPU  REPLICAS  AGE
pytorch-example  default    Running  kubeflow.org/v1  pytorch    4    2/4       1m
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
Replicas:  4/4
Age:       5m

Pods:
  NAME                    STATUS    IP         NODE
  pytorch-example-worker-0  Running   10.0.0.1   node-1
  pytorch-example-worker-1  Running   10.0.0.2   node-2
  pytorch-example-worker-2  Running   10.0.0.3   node-3
  pytorch-example-worker-3  Running   10.0.0.4   node-4
```

Add `--details` to include the full YAML configuration:

```shell
$ arena job get pytorch-example --details
```

## 5. View Logs

Stream logs from the training job:

```shell
$ arena job logs pytorch-example
[Epoch 0] Loss: 2.3026
[Epoch 1] Loss: 1.9453
[Epoch 2] Loss: 1.6234
...
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
$ arena job delete -f examples/v2/pytorch-simple.yaml
pytorchjob/pytorch-example deleted
```

## Next Steps

- **YAML schema reference**: [keps/arena-v2/yaml-schema.md](../keps/arena-v2/yaml-schema.md) — all available fields and options
- **More examples**: [examples/v2/](../examples/v2/) — TensorFlow, MPI, standalone, TensorBoard, mounts
- **Try `--set` overrides**: `arena job run -f examples/v2/pytorch-simple.yaml --set worker.replicas=2`
- **Try dry-run**: `arena job run -f examples/v2/pytorch-simple.yaml --dry-run`
- **Design document**: [keps/arena-v2/README.md](../keps/arena-v2/README.md) — v2 architecture and migration guide
