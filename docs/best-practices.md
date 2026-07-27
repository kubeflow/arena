# Best Practices

Practical guidance for using Arena v2 effectively.

## Configuration Management

### Use YAML files as the source of truth

YAML files are version-controllable, reviewable, and reusable across experiments. Commit them to Git so your team can track changes, review pull requests, and reproduce any past run without retyping flags.

```yaml
version: 0.1.0
name: llm-finetune
framework:
  name: pytorch
  options:
    nproc_per_node: auto
image: nvcr.io/nvidia/pytorch:23.10
run: torchrun train.py --epochs 10
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: 8
    cpu: "32"
    memory: 128Gi
```

### Use --set for quick experiments

`--set` applies Helm-style dot-notation overrides on top of a YAML file before parsing. This lets you sweep parameters without editing the file. When a key contains dots (such as a Kubernetes resource name), wrap it in single quotes so it is treated as a single segment.

```shell
# Override replica count
arena job run -f train.yaml --set worker.replicas=8

# Override GPU count — quote the resource key because it contains dots
arena job run -f train.yaml --set worker.resources.'nvidia.com/gpu'=2

# Override multiple fields at once
arena job run -f train.yaml \
  --set worker.replicas=2 \
  --set worker.resources.'nvidia.com/gpu'=1 \
  --set envs.NCCL_DEBUG=INFO
```

See [yaml-schema.md](yaml-schema.md) for the full field reference.

### Always dry-run before submitting

`--dry-run` builds the CRD (and TensorBoard resources if enabled) and prints them as JSON without touching the cluster. Use it to validate field values, verify provider mapping, and catch schema errors early.

```shell
arena job run -f train.yaml --dry-run
```

### Organize YAML files by project and experiment

Keep a directory structure that mirrors your team's workflow so configs are easy to find and reuse.

```
training-configs/
  project-a/
    base.yaml              # shared defaults
    exp-01-lr1e-4.yaml     # experiment variants
    exp-02-lr3e-5.yaml
  project-b/
    pretrain.yaml
    finetune.yaml
```

## Resource Management

### Set requests equal to limits for Guaranteed QoS

Arena v2 writes every `resources` value to **both** `requests` and `limits` in the Pod spec. When you specify both `cpu` and `memory`, Kubernetes assigns the **Guaranteed QoS** class, which protects long-running training jobs from priority eviction under node contention. Specifying only GPUs (without CPU/memory) does **not** qualify for Guaranteed QoS.

```yaml
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: 8        # GPU request
    cpu: "32"                # CPU request  → also becomes the limit
    memory: 128Gi            # Memory request → also becomes the limit
```

### Select GPU type via nodeSelector

Pin jobs to specific GPU hardware (e.g. A100 vs V100) using `scheduling.node_selector`. The `nvidia.com/gpu.product` label is the standard node label set by the NVIDIA device plugin.

```yaml
scheduling:
  node_selector:
    nvidia.com/gpu.product: "NVIDIA-A100-SXM4-40GB"
```

You can also combine multiple selectors to constrain by zone, disk type, or other node labels:

```yaml
scheduling:
  node_selector:
    nvidia.com/gpu.product: "NVIDIA-A100-SXM4-40GB"
    topology.kubernetes.io/zone: us-west-2a
```

### Use shared memory for multi-process training

Multi-process frameworks (PyTorch DDP, DeepSpeed) rely on `/dev/shm` for inter-process communication. Add a `shm` storage entry to mount a tmpfs-backed shared memory volume instead of relying on the default 64 MB.

```yaml
storages:
  - name: shm
    shm: 64Gi                # Optional. Default: 2Gi
    # mount_path: /dev/shm   # Optional. Default: /dev/shm
```

### Right-size resources to avoid waste

Over-allocating GPUs or CPU/memory wastes cluster capacity and increases queue times. Start with a modest allocation, profile the job, then adjust. A single-GPU experiment rarely needs 32 CPUs — 8 is often sufficient.

```yaml
# Good: right-sized for a single-GPU experiment
worker:
  replicas: 1
  resources:
    nvidia.com/gpu: 1
    cpu: "8"
    memory: 32Gi
```

### Full resource example

```yaml
version: 0.1.0
name: resnet-training
framework:
  name: pytorch
  options:
    nproc_per_node: auto
image: nvcr.io/nvidia/pytorch:23.10
run: torchrun train.py --epochs 50 --batch-size 256
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: 8
    cpu: "48"
    memory: 256Gi
envs:
  NCCL_DEBUG: INFO
  NCCL_IB_DISABLE: "0"
storages:
  - name: dataset
    pvc: training-data-pvc
    mount_path: /data
  - name: shm
    shm: 64Gi
  - name: checkpoints
    pvc: ckpt-pvc
    mount_path: /ckpts
scheduling:
  node_selector:
    nvidia.com/gpu.product: "NVIDIA-A100-SXM4-40GB"
lifecycle:
  active_deadline: 24h
  ttl_after_finished: 7d
  backoff_limit: 3
```

## Job Lifecycle

### Use descriptive job names

The job name becomes the Kubernetes resource name for the CRD, the ConfigMap that stores your YAML, and the TensorBoard Service (`<jobname>-tensorboard`). Choose names that are DNS-compliant and self-documenting.

```yaml
# Good
name: resnet50-imagenet-exp01

# Avoid
name: job1
```

### Suspend and resume to pause without losing state

Suspend patches `spec.runPolicy.suspend` to `true`, which stops pods while preserving the job CRD and its configuration. Resume sets it back to `false`. This is useful for freeing GPUs temporarily or debugging.

```shell
arena job suspend resnet50-imagenet-exp01
arena job resume  resnet50-imagenet-exp01
```

Suspend/resume are **operations**, not YAML configuration — they do not belong in the task spec.

### OwnerReferences auto-cleanup on deletion

When you submit a job, Arena v2 creates the CRD first, then creates the ConfigMap and TensorBoard resources with `ownerReferences` pointing to that CRD. Deleting the CRD triggers Kubernetes garbage collection, which cascades to all owned resources automatically — no manual cleanup needed.

```shell
# Delete by name
arena job delete resnet50-imagenet-exp01

# Delete by file (reads the name from the YAML)
arena job delete -f train.yaml
```

This is a single API call in v2, compared to v1's three-step delete process (read ConfigMap, delete each resource, delete ConfigMap).

## Data Preparation

### Use sync containers for code and data injection

The `sync` block generates init containers that run before the main training container. Use `git` for source code, `rsync` for remote files, and `hdfs` for HDFS-stored datasets. Each entry creates a named init container (`arena-sync-0`, `arena-sync-1`, ...).

```yaml
sync:
  - git: https://github.com/org/training-code.git
    branch: main
    local_path: /workspace/code

  - rsync: 10.88.29.56::backup/data/dataset.zip
    local_path: /workspace/dataset

  - hdfs: hdfs://namenode:8020/models/resnet
    local_path: /workspace/models
```

### Use init containers for custom setup

The `init` block lets you define arbitrary init containers with `run` and `shell` semantics identical to the main container. Use them for `pip install`, data preprocessing, or any one-time setup. `sync` entries always run before `init` entries.

```yaml
init:
  - name: install-deps
    image: pytorch:2.1
    run: pip install transformers accelerate
    mounts:
      - name: code
        mount_path: /workspace/code

  - name: preprocess
    image: custom-etl:latest
    run: /prepare.sh --input /data --output /processed
    shell: /bin/bash
    mounts:
      - name: dataset
        mount_path: /data
```

### Understand the mount resolution pattern

Sync and init containers inherit volumes from the top-level `storages` block. The `mounts` field controls which storages are actually mounted in each container, following an all/subset/none pattern:

- **No storages defined**: no volumes, no mounts.
- **Storages defined, no `mounts` field**: all storages are mounted at their default paths.
- **Storages defined, `mounts` field present**: all storages become volumes, but only the listed storages get volume mounts (mount fields override storage defaults).

```yaml
storages:
  - name: dataset
    pvc: training-data-pvc
    mount_path: /data
  - name: code
    tmp: 1Gi
    mount_path: /code

sync:
  - git: https://github.com/org/training-code.git
    local_path: /code
    mounts:
      - name: code           # Override: mount code storage at /code
        mount_path: /code
      # dataset is NOT listed here, so it is not mounted in this init container
```

### Watch for local_path / mount mismatches

If a sync entry's `local_path` does not match any resolved mount path, the synced data is written to the container's ephemeral storage and will be lost when the pod restarts. Arena prints a warning to stderr when this happens. Always ensure `local_path` matches a `mount_path` in the same entry's `mounts` list or in the referenced storage.

```yaml
# Correct: local_path matches the mount_path
sync:
  - git: https://github.com/org/code.git
    local_path: /code
    mounts:
      - name: code
        mount_path: /code      # ← matches local_path
```

## TensorBoard Integration

### Enable TensorBoard in YAML

Set `logging.tensorboard.enabled: true` and specify the `logdir` that your training script writes checkpoints/events to. Arena creates a Deployment and a Service (`<jobname>-tensorboard`) with `ownerReferences` to the training job CRD.

```yaml
logging:
  tensorboard:
    enabled: true
    logdir: /training_logs
    image: tensorflow/tensorflow:2.21.0    # Optional. Defaults to tensorflow/tensorflow:2.21.0
```

### Access TensorBoard via port-forward

The TensorBoard Service listens on port 6006. Forward it locally to view the dashboard.

```shell
kubectl port-forward svc/resnet50-imagenet-exp01-tensorboard 6006:6006
```

Then open `http://localhost:6006` in your browser.

### TensorBoard resources are auto-cleaned

Because TensorBoard resources carry `ownerReferences` to the training job CRD, they are garbage-collected automatically when you delete the job. No separate cleanup command is needed.

```shell
arena job delete resnet50-imagenet-exp01
# CRD, ConfigMap, TensorBoard Deployment, and TensorBoard Service are all removed
```

### Selectively mount storages in TensorBoard

By default, TensorBoard mounts all storages. Use `logging.tensorboard.mounts` to mount only the storages that contain your logs, reducing unnecessary volume attachments. All storages still become volumes, but only listed ones get volume mounts — and mount fields override storage defaults.

```yaml
storages:
  - name: data
    pvc: training-data-pvc
    mount_path: /data
  - name: logs
    pvc: tensorboard-logs-pvc
    mount_path: /logs
  - name: cache
    pvc: model-cache-pvc
    mount_path: /cache

logging:
  tensorboard:
    enabled: true
    logdir: /tb/logs
    mounts:
      - name: logs              # Only mount the logs PVC in TensorBoard
        mount_path: /tb/logs    # Override the default /logs path
```

See [frameworks.md](frameworks.md) for framework-specific TensorBoard guidance.

## Migration from v1

### worker.replicas in v2 EXCLUDES master

This is the most important migration difference for PyTorch. In v1, `--workers=N` **includes** the master (Arena internally decrements it by 1). In v2, `worker.replicas=N` **excludes** the master — the master is always an additional Pod.

**Migration formula:** `worker.replicas = --workers - 1`

| v1 (`--workers`) | v2 (`worker.replicas`) | Total GPU-using pods |
|---|---|---|
| 1 | 0 (omit worker, use master block) | 1 |
| 2 | 1 | 2 |
| 5 | 4 | 5 |

When `--workers` is 1, `worker.replicas` would be 0, which is invalid. In that case, omit the `worker` block and specify only `master`:

```yaml
# v1: arena submit pytorchjob --workers 1 --gpus 1 "python train.py"
# v2 equivalent:
version: 0.1.0
name: single-node-training
framework:
  name: pytorch
image: pytorch:2.1
run: python train.py
master:
  replicas: 1
  resources:
    nvidia.com/gpu: 1
```

For non-PyTorch frameworks (MPI, TensorFlow, etc.), `worker.replicas` maps directly with no master adjustment. See [yaml-schema.md](yaml-schema.md) for the full provider mapping table.

### arena job run is the v2 preferred way

`arena job run -f <file>` is the YAML-first entry point. It loads, validates, builds the CRD, and submits it directly via the Kubernetes API — no Helm, no kubectl, no temp files.

```shell
arena job run -f train.yaml
arena job run -f train.yaml --dry-run
arena job run -f train.yaml --set worker.replicas=8
```

### arena submit remains for backward compatibility

The legacy `arena submit` command maps v1-style flags to v2's internal Task model, so existing CI/CD scripts continue to work without changes.

```shell
# v1-style invocation still works
arena submit pytorchjob --name my-job --workers 5 --gpus 2 "python train.py"
```

### --set replaces most v1 CLI flags

Instead of remembering 40+ CLI flags, define your job in YAML and use `--set` for one-off overrides. This is especially useful for CI pipelines that parameterize a single template.

```shell
# v1 style (many flags)
arena submit pytorchjob --name exp01 --workers 5 --gpus 2 --cpu 8 --memory 32Gi \
  --data dataset:/data:training-pvc --tensorboard --tensorboard-logdir /logs \
  "python train.py"

# v2 equivalent (YAML + --set)
arena job run -f train.yaml \
  --set name=exp01 \
  --set worker.replicas=4 \
  --set worker.resources.'nvidia.com/gpu'=2
```

### No Helm dependency in v2

v2 generates Kubernetes resources directly in Go using `client-go` and the dynamic API. It eliminates the Helm SDK, all 23 Helm charts, and all shell-outs to `helm`/`kubectl`. This means:

- No `helm` or `kubectl` binaries required in your `PATH`.
- No temp files that can leak if the process crashes mid-submission.
- Atomic submission — the CRD is either created or it isn't.
- Smaller binary size (only `client-go` + `cobra`).

See [cli-reference.md](cli-reference.md) for the complete v2 command reference.
