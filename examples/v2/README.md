# Arena v2 Examples

This directory contains YAML example files for Arena v2, the next-generation AI workload
CLI for Kubernetes. Each file follows the Arena v2 schema (`version: 0.1.0`) and can be
submitted with `arena job run -f <file>`.

## Quick Start

The examples in `quickstart/` are self-contained and immediately runnable with
`--dry-run` (no cluster required):

```bash
# Dry-run a single-node PyTorch job (prints the rendered CRD to stdout)
arena job run -f examples/v2/quickstart/pytorch-standalone.yaml --dry-run

# Dry-run a distributed TensorFlow job
arena job run -f examples/v2/quickstart/tensorflow-simple.yaml --dry-run

# Dry-run an MPI/Horovod job
arena job run -f examples/v2/quickstart/mpi-simple.yaml --dry-run
```

To actually submit a job to a cluster (requires a running Kubernetes cluster with the
relevant operators installed):

```bash
arena job run -f examples/v2/quickstart/pytorch-standalone.yaml
```

## Directory Structure

| Directory      | Purpose                                                                              |
| -------------- | ------------------------------------------------------------------------------------ |
| `quickstart/`  | Self-contained, immediately runnable examples. Inline scripts, public images.        |
| `reference/`   | Real-world examples adapted from the Kubeflow trainer repo. May need PVC setup.      |
| `pretrain/`    | Pre-training examples. See `pretrain/README.md` for details.                         |
| `posttrain/`   | Post-training templates (planned). See `posttrain/README.md` for details.            |
| `rl/`          | Reinforcement learning templates (planned). See `rl/README.md` for details.          |
| `inference/`   | Inference templates (planned). See `inference/README.md` for details.             |

### Notes by category

- **`quickstart/`** examples are self-contained: they embed the training script inline in
  the `run:` field and use public container images. They work out of the box with
  `--dry-run` and can be submitted to any cluster with the matching operator installed.

- **`reference/`** examples use real Kubeflow trainer container images and may require
  PersistentVolumeClaims (PVCs) to be created before submission. Most are e2e verified on
  an ACK cluster. Two files — `tf-with-roles.yaml` and `pytorch-tensorboard-mounts.yaml` —
  are schema demos that pass `--dry-run` but have not been e2e verified; they are the only
  examples demonstrating TF chief/ps/evaluator roles and TensorBoard integration
  respectively.

- **`pretrain/`** contains verified pre-training examples. See `pretrain/README.md` for
  details.

- **`posttrain/`, `rl/`, `inference/`** are placeholders for future templates. See the
  `README.md` in each directory for planned examples.

- **Elastic examples** (`pytorch-elastic-*`) run as fixed-replica jobs because
  `elasticPolicy` is not part of Arena v2's current schema. They demonstrate the torchrun
  / torchelastic entrypoint pattern but do not support elastic scaling yet.

## Example Catalog

### quickstart/

| File                          | Framework  | Description                                                        | Status        |
| ----------------------------- | ---------- | ------------------------------------------------------------------ | ------------- |
| `pytorch-standalone.yaml`     | PyTorch    | Single-node PyTorch training, inline script, synthetic data.       | E2E verified  |
| `pytorch-simple.yaml`         | PyTorch    | Distributed PyTorch (2 workers) with Gloo backend, inline script.  | E2E verified  |
| `tensorflow-simple.yaml`      | TensorFlow | Distributed TF (2 workers), MultiWorkerMirroredStrategy.           | E2E verified  |
| `mpi-simple.yaml`             | MPI        | Horovod (2 workers) with DistributedOptimizer, inline script.      | E2E verified  |

Run any quickstart example:

```bash
arena job run -f examples/v2/quickstart/pytorch-standalone.yaml --dry-run
```

### reference/

| File                                | Framework  | Description                                                              | Status       |
| ----------------------------------- | ---------- | ------------------------------------------------------------------------ | ------------ |
| `pytorch-mnist.yaml`                | PyTorch    | PyTorch MNIST, Master+Worker, CPU.                                       | E2E verified |
| `pytorch-mnist-nccl.yaml`           | PyTorch    | PyTorch MNIST, NCCL backend adapted to Gloo for CPU.                  | E2E verified |
| `pytorch-mnist-gloo.yaml`           | PyTorch    | PyTorch MNIST with Gloo backend, CPU.                                    | E2E verified |
| `pytorch-mnist-mpi-backend.yaml`    | PyTorch    | PyTorch MNIST with MPI backend, CPU.                                     | E2E verified |
| `pytorch-smoke-test.yaml`           | PyTorch    | PyTorch send/recv smoke test, 3 workers.                                 | E2E verified |
| `pytorch-elastic-echo.yaml`         | PyTorch    | PyTorch torchelastic echo test, Worker-only (fixed-replica).             | E2E verified |
| `pytorch-elastic-imagenet.yaml`     | PyTorch    | PyTorch torchelastic ImageNet, Worker-only (fixed-replica).              | E2E verified |
| `pytorch-tensorboard-mounts.yaml`   | PyTorch    | PyTorch with TensorBoard integration and volume mounts.                  | Schema demo  |
| `tf-with-roles.yaml`                | TensorFlow | TFJob with Chief, PS, and Evaluator roles.                               | Schema demo  |
| `tf-simple.yaml`                    | TensorFlow | TFJob simple MNIST, 2 workers (CPU, MultiWorkerMirroredStrategy).        | E2E verified |
| `tf-dist-mnist.yaml`                | TensorFlow | TFJob distributed MNIST, Worker-only (CPU).                          | E2E verified |
| `tf-multi-worker-gpu.yaml`          | TensorFlow | TFJob MultiWorkerMirroredStrategy, CPU + PVC.                         | E2E verified |
| `tf-mnist-summaries.yaml`           | TensorFlow | TFJob MNIST with TensorBoard summaries (CPU, no PVC).                    | E2E verified |
| `mpi-horovod-mnist.yaml`            | MPI        | MPIJob Horovod MNIST with mpirun.                                        | E2E verified |
| `mpi-horovod-elastic.yaml`          | MPI        | MPIJob elastic Horovod with horovodrun.                                  | E2E verified |

Run any reference example:

```bash
arena job run -f examples/v2/reference/tf-simple.yaml --dry-run
```

For reference examples that require a PVC, create the PVC first. For example:

```bash
# The PVC name and source manifest are noted in each YAML file's header comment.
kubectl apply -f https://raw.githubusercontent.com/kubeflow/training-operator/master/examples/tensorflow/distribution_strategy/pvc.yaml
```

### pretrain/

| File                    | Framework  | Description                                          | Status            |
| ----------------------- | ---------- | ---------------------------------------------------- | ----------------- |
| `deepspeed-bert.yaml`   | PyTorch    | DeepSpeed BERT pre-training with ZeRO-1 (GPU).       | E2E verified (GPU) |

### posttrain/

Post-training templates (SFT, LoRA, RLHF) are planned. See `posttrain/README.md` for details.

### rl/

Reinforcement learning templates are planned. See `rl/README.md` for details.

### inference/

Inference templates are planned. See `inference/README.md` for details.

## Schema Overview

Every Arena v2 YAML file uses `version: 0.1.0` and supports these top-level fields:

```yaml
version: 0.1.0
name: my-job              # DNS-1123 label, required
image: my-image:tag       # container image, required
framework:
  name: pytorch           # pytorch | tensorflow | mpi | deepspeed | horovod
  options:                # optional, framework-specific
    nproc_per_node: auto  # PyTorch: auto | gpu | cpu | <integer>
    ps_count: 1           # TensorFlow: number of PS replicas (alternative to ps: role)
    slots_per_worker: 1   # MPI/DeepSpeed: number of slots per worker
    mpi_implementation: OpenMPI  # MPI: OpenMPI | Intel | MPICH
run: |                    # command to execute, required
  python -c "print('Training started.')"
restart: OnFailure        # optional: Always, OnFailure, Never
image_pull_policy: Always # optional: Always, IfNotPresent, Never

# Role configuration (framework-dependent)
worker:                   # all frameworks
  replicas: 2
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
master:                   # PyTorch only (replicas forced to 1)
  resources:
    cpu: "2"
    memory: "8Gi"
# chief:                  # TensorFlow only (replicas forced to 1)
#   resources:
#     cpu: "2"
# ps:                     # TensorFlow only (replicas configurable)
#   replicas: 1
#   resources:
#     cpu: "4"
# evaluator:              # TensorFlow only (replicas forced to 1)
#   resources:
#     cpu: "2"
# launcher:               # MPI only (replicas forced to 1)
#   resources:
#     cpu: "1"
#     memory: "2Gi"

storages:                 # volume mounts
  - name: data
    mount_path: /data
    pvc: my-pvc           # PVC-backed storage
  - name: logs
    mount_path: /logs
    pvc: logs-pvc
  - name: shm
    mount_path: /dev/shm  # default for shm
    shm: 64Gi             # shared memory (emptyDir with medium: Memory)

envs:                     # environment variables
  KEY: VALUE

logging:                  # optional logging integrations
  tensorboard:
    enabled: true
    logdir: /tb/logs      # TensorBoard log directory
    image: tensorflow/tensorflow:2.21.0  # optional, defaults to TF image
    mounts:               # optional: selectively mount storages into TensorBoard pod
      - name: logs
        mount_path: /tb/logs

lifecycle:                # optional lifecycle controls
  clean_pod_policy: None  # None, Running, All
  # success_policy: ChiefWorker  # ChiefWorker, AllWorkers (TF only)
```

### Role rules by framework

| Role      | Frameworks        | Replicas                    |
| --------- | ----------------- | --------------------------- |
| `worker`  | PyTorch, TF, MPI  | User-specified (>= 1)       |
| `master`  | PyTorch           | Forced to 1                 |
| `chief`   | TensorFlow        | Forced to 1                 |
| `ps`      | TensorFlow        | User-specified (required)   |
| `evaluator`| TensorFlow       | Forced to 1                 |
| `launcher`| MPI               | Forced to 1                 |

For roles forced to 1 (`master`, `chief`, `evaluator`, `launcher`), `replicas:` is
managed automatically — no need to specify it.

For `ps:` (TensorFlow Parameter Server), `replicas:` is required because it is
unconstrained. Alternatively, use `framework.options.ps_count` to configure PS
replicas without a `ps:` role section.

## Submitting Jobs

```bash
# Dry-run (renders the CRD, no cluster needed)
arena job run -f examples/v2/quickstart/pytorch-simple.yaml --dry-run

# Submit to cluster
arena job run -f examples/v2/quickstart/pytorch-simple.yaml

# Submit with inline overrides
arena job run -f examples/v2/quickstart/pytorch-simple.yaml --set worker.replicas=4
```

For the legacy `submit` interface (backward compatible with Arena v1):

```bash
arena submit pytorch --name my-job --image pytorch:2.1 --gpus 2 "python train.py"
```
