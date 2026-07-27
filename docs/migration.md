# Migrating from Arena v1 to v2

## Overview

Arena v2 is a complete architectural redesign of the Arena CLI for submitting AI
training and inference workloads to Kubernetes. The key changes from v1 are:

- **YAML-first configuration** — Job definitions are structured YAML files that
  can be versioned in Git, reviewed, and reused. CLI flags remain available for
  quick one-off submissions and for backward compatibility via `arena submit`.
- **No Helm dependency** — v2 generates Operator CRDs directly as Go structs
  (`unstructured.Unstructured`) and submits them through the Kubernetes dynamic
  client API. The entire Helm chart rendering pipeline (23 charts, Helm Go SDK,
  temp files, `kubectl apply` shell-outs) is eliminated.
- **Direct K8s API** — All cluster interaction goes through `client-go`. No
  `helm` or `kubectl` binaries are required on the user's `PATH`.
- **ownerReferences for cleanup** — All resources created by a single job
  submission (ConfigMap, TensorBoard Deployment + Service) are owned by the
  top-level training job CR. Deleting the CR triggers Kubernetes garbage
  collection to cascade-delete all owned resources, making cleanup a single
  atomic operation with no risk of resource leaks.

During the transition period, v1 and v2 coexist as separate binaries (`arena`
and `arena-v2`). v1 code is not modified; v2 packages are self-contained.

For the full YAML schema specification, see [yaml-schema.md](yaml-schema.md).
For recommended workflows, see [best-practices.md](best-practices.md).

---

## Key Behavioral Changes

### worker.replicas Excludes Master

> **This is the #1 migration gotcha.** Getting this wrong will launch one fewer
> or one extra worker Pod than intended.

In v1, `--workers N` means **N total processes, including the master**. The v1
PyTorch submit path internally decrements this by 1 to arrive at the actual
number of worker replicas.

In v2, `worker.replicas: N` means **N worker replicas**. The master is always a
separate, additional Pod (fixed at 1 replica).

**Migration formula:**

```
worker.replicas = --workers - 1
```

Since `--workers` is always >= 1, when `--workers` is 1 the formula yields
`worker.replicas = 0`, which violates the `> 0` constraint. In this single-node
case, omit the `worker` block entirely and specify only the `master` block.

**Example — v1:**

```bash
arena submit pytorch --name my-job --image pytorch:2.1 --gpus 1 --workers 4 "python train.py"
```

**Example — v2:**

```yaml
version: 0.1.0
name: my-job
image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
framework:
  name: pytorch
worker:
  replicas: 3  # --workers 4 minus 1 for master
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
run: |
  python -c "
  import torch, torch.distributed as dist, torch.nn as nn
  dist.init_process_group(backend='gloo')
  rank = dist.get_rank()
  model = nn.Linear(10, 1)
  optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
  for epoch in range(3):
      x, y = torch.randn(64, 10), torch.randn(64, 1)
      loss = nn.MSELoss()(model(x), y)
      optimizer.zero_grad(); loss.backward(); optimizer.step()
      if rank == 0:
          print(f'Epoch {epoch+1}/3 - loss: {loss.item():.4f}')
  if rank == 0:
      print('Training complete.')
  dist.destroy_process_group()
  "
```

Note the `replicas: 3` (not 4). The total GPU-using replicas is
`worker.replicas + 1` (the +1 being the master). If `master` is omitted, the
master inherits the worker's configuration by default. The master can also be
independently configured via the `master` block.

### Resource Lifecycle

**v1:** Resources are tracked via a resource list stored in a ConfigMap.
Deletion follows a 3-step pipeline: read ConfigMap, delete each listed
resource, delete the ConfigMap. If the process crashes mid-delete or the
resource list is incomplete, resources leak.

**v2:** All created resources carry `ownerReferences` pointing to the
top-level training job CR. Deletion is a single API call — delete the CR and
Kubernetes garbage collection cascade-deletes all owned resources:

```
PyTorchJob CRD
  ├── owns → ConfigMap (stores effective YAML)
  ├── owns → TensorBoard Deployment
  └── owns → TensorBoard Service
```

This also applies to `ttlSecondsAfterFinished`: when the training job CR
expires, all owned resources are cleaned up alongside it.

### V1 Jobs Cannot Be Managed by V2 CLI

V2 detects v1 jobs by their ConfigMap format and refuses to manage them. You
cannot use `arena-v2 job get`, `arena-v2 job delete`, or similar commands on
jobs submitted by v1.

To migrate a running v1 job:

1. Delete the v1 job using the v1 CLI (`arena delete <name>`).
2. Resubmit the job using v2 (`arena-v2 job run -f job.yaml`).

### No Helm Dependency

V2 directly creates CRDs via the Kubernetes API. There are no Helm charts, no
Helm Go SDK, no temp files, and no shell-outs to `kubectl`. This means:

- No `helm` or `kubectl` binaries required on the user's `PATH`.
- No command-line injection risk from user input.
- No temp file leaks (v1 created 3 temp files per submission).
- Submission is atomic — either the CRD is created or it isn't.

---

## Flag Mapping Table

The following tables map v1 CLI flags to v2 YAML fields, organized by category.
Field names in v2 use `snake_case` per the schema specification.

### Identity

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--name` | `name` | Required. Must be a valid DNS-1123 label. |
| `--namespace` | `namespace` | Optional. |

### Resources

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--gpus` | `worker.resources.nvidia.com/gpu` | String value (e.g. `"1"`). |
| `--cpu` | `worker.resources.cpu` | String value. |
| `--memory` | `worker.resources.memory` | String value (e.g. `"8Gi"`). |
| `--device` | `worker.resources.<device-name>` | v1 `--device vendor.com/device=count` maps to v2 flat `vendor.com/device: count` in `resources`. e.g. `--device hugepages-2Mi=32Gi`. |
| `--workers` | `worker.replicas` | **Subtract 1 for master.** See [worker.replicas Excludes Master](#workerreplicas-excludes-master). |

### Scheduling

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--selector` | `scheduling.node_selector` | Array of `key=value` in v1; map in v2. |
| `--toleration` | `scheduling.tolerations` | Array of toleration objects. |
| `-p, --priority` | `scheduling.priority` | Integer. |
| `--priority-class-name` | `scheduling.priority_class_name` | String. |
| `--gang` | `scheduling.gang.enabled` | Boolean. |
| `--scheduler` | `scheduling.scheduler_name` | String. |
| `--affinity-policy` | `scheduling.affinity.policy` | `none` / `spread` / `binpack`. |
| `--affinity-constraint` | `scheduling.affinity.constraint` | `preferred` / `required`. |
| `--queue` | `scheduling.queue` | String. |
| `--rdma` | — | **Not planned.** |

### Data and Volumes

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `-d, --data` | `storages[].pvc` + `storages[].mount_path` | PVC-backed storage. |
| `--data-dir` | `storages[].hostpath` + `storages[].mount_path` | Host path storage. |
| `--config-file` | `storages[].configmap` + `storages[].mount_path` | ConfigMap storage. |
| `--share-memory` | `storages[].shm` | emptyDir medium: Memory at `/dev/shm`. Default `"2Gi"` in v1. |

v2 also supports `storages[].tmp` (emptyDir), `storages[].secret` (Secret
storage), and `storages[].sub_path` for sub-path mounts. Each storage entry
must specify exactly one storage type.

### Environment and Labels

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `-e, --env` | `envs` | Supports plain values, `secretKeyRef`, and `configMapKeyRef`. |
| `-a, --annotation` | `annotations` | Map of string to string. |
| `-l, --label` | `labels` | Map of string to string. |

### Execution

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--image` | `image` | Required. |
| `--working-dir` | `working_dir` | Defaults to image default if omitted. |
| `--shell` | `shell` | Invalid values fall back to `/bin/sh`. |
| `--image-pull-policy` | `image_pull_policy` | `Always` / `IfNotPresent` / `Never`. |
| `--image-pull-secret` | `image_pull_secrets` | Array of secret names. Reference-only; no auto-create. |
| `--hostNetwork` | `host_network` | Boolean. |
| `--hostIPC` | `host_ipc` | Boolean. |
| `--hostPID` | `host_pid` | Boolean. |

### Lifecycle

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--clean-task-policy` | `lifecycle.clean_pod_policy` | `None` / `Running` / `All`. |
| `--running-timeout` | `lifecycle.active_deadline` | Duration string (e.g. `1h`, `7d`). |
| `--ttl-after-finished` | `lifecycle.ttl_after_finished` | Duration string. |
| `--job-backoff-limit` | `lifecycle.backoff_limit` | Integer. |
| `--retry` | `lifecycle.backoff_limit` | Redundant in v1; maps to same field as `--job-backoff-limit`. |
| `--success-policy` | `lifecycle.success_policy` | TFJob only: `ChiefWorker` / `AllWorkers`. |

### Sync

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--sync-mode` | `sync[].git` / `sync[].rsync` / `sync[].hdfs` | Type key determines sync method. |
| `--sync-source` | `sync[].git` / `sync[].rsync` / `sync[].hdfs` | Value of the source URL/path. |
| `--sync-image` | `sync[].image` | Container image for the sync init container. |

Each sync entry also supports `sync[].branch` (for Git), `sync[].local_path`
(required — target path inside the container), and `sync[].mounts` for
overriding storage mount points.

### TensorBoard

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--tensorboard` | `logging.tensorboard.enabled` | Boolean. |
| `--tensorboard-image` | `logging.tensorboard.image` | String. |
| `--logdir` | `logging.tensorboard.logdir` | Default `"/training_logs"` in v1. |

### PyTorch-Specific

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--cpu` | `worker.resources.cpu` and `master.resources.cpu` | Applied to both worker and master. |
| `--memory` | `worker.resources.memory` and `master.resources.memory` | Applied to both worker and master. |
| `--nproc-per-node` | `framework.options.nproc_per_node` | `auto` / `gpu` / `cpu` / positive integer. |

### TFJob-Specific

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--ps` (count) | `ps.replicas` | Integer. |
| `--ps-cpu`, `--ps-cpu-limit` | `ps.resources.cpu` | |
| `--ps-memory`, `--ps-memory-limit` | `ps.resources.memory` | |
| `--ps-gpus` | `ps.resources.nvidia.com/gpu` | |
| `--chief` (bool) | `chief` | Presence of `chief` block enables the role. |
| `--chief-cpu`, `--chief-memory`, etc. | `chief.resources` | |
| `--evaluator` (bool) | `evaluator` | Presence of `evaluator` block enables the role. |
| `--evaluator-cpu`, `--evaluator-memory`, etc. | `evaluator.resources` | |
| `--worker-cpu`, `--worker-memory`, etc. | `worker.resources` | |
| `--success-policy` | `lifecycle.success_policy` | Moved to `lifecycle` block in v2. |
| `--ps-image` | — | Not yet implemented. |
| `--ps-selector` | — | Not yet implemented. |
| `--ps-affinity-policy` | — | Not yet implemented. |
| `--chief-selector` | — | Not yet implemented. |
| `--evaluator-selector` | — | Not yet implemented. |
| `--worker-image` | — | Not yet implemented. |
| `--worker-port` | — | Not yet implemented. |
| `--worker-selector` | — | Not yet implemented. |
| `--worker-affinity-policy` | — | Not yet implemented. |
| `--role-sequence` | — | Implementation detail, not user-facing. |

### MPI-Specific

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--cpu` | `worker.resources.cpu` | |
| `--memory` | `worker.resources.memory` | |
| `--gputopology` | `host_network` + `worker.resources` + `labels` | Sets `gpu-topology` / `gpu-topology-replica` labels. |
| `--mounts-on-launcher` | `framework.options.mounts_on_launcher` | Boolean. |
| (no v1 flag) | `framework.options.run_launcher_as_worker` | v2-only field. |
| (no v1 flag) | `framework.options.slots_per_worker` | v2-only field. |

### Horovod-Specific

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--ssh-port` | — | **Not in schema.** v2 always uses port 22. |
| `--cpu` | `worker.resources.cpu` | |
| `--memory` | `worker.resources.memory` | |

### DeepSpeed-Specific

| v1 Flag | v2 YAML Field | Notes |
|---------|---------------|-------|
| `--cpu` | `worker.resources.cpu` | |
| `--memory` | `worker.resources.memory` | |
| `--job-restart-policy` | `restart` | |
| `--launcher-selector` | — | Not yet implemented. |
| `--ssh-secret` | — | **Not in schema.** MPI Operator handles SSH internally. |
| `--launcher-annotation` | — | Not yet implemented. |
| `--worker-annotation` | — | Not yet implemented. |

---

## Features Not Planned in v2

| v1 Feature | Status | Notes |
|------------|--------|-------|
| `--model-name` | Not planned | Use an external model registry. |
| `--model-source` | Not planned | Use an external model registry. |
| `--rdma` | Not planned | |
| ETJob (Elastic Training) | Not planned | |
| `--ssh-port` (Horovod) | Not in schema | v2 always uses port 22. |
| `--ssh-secret` (DeepSpeed) | Not in schema | MPI Operator handles SSH internally; no secret required. |

---

## Features in Schema but Not Yet Implemented

These features have schema definitions in v2 but are not yet implemented.
They are targeted for future releases.

| Feature | Status |
|---------|--------|
| Ray provider | Framework placeholder only, provider not yet implemented |
| Per-role selectors (`--ps-selector`, `--chief-selector`, `--evaluator-selector`, `--worker-selector`) | Not yet implemented |
| Per-role affinity policies (`--ps-affinity-policy`, `--worker-affinity-policy`) | Not yet implemented |
| Per-role images (`--ps-image`, `--worker-image`) | Not yet implemented |
| `--worker-port` (TFJob) | Not yet implemented |

---

## Conversion Examples

### Example 1: PyTorch Distributed Training

**v1 command:**

```bash
arena submit pytorch \
  --name my-job \
  --image pytorch:2.1 \
  --gpus 1 \
  --workers 4 \
  "python train.py"
```

**v2 YAML:**

```yaml
version: 0.1.0
name: my-job
image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
framework:
  name: pytorch
worker:
  replicas: 3  # --workers 4 minus 1 for master
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
run: |
  python -c "
  import torch, torch.distributed as dist, torch.nn as nn
  dist.init_process_group(backend='gloo')
  rank = dist.get_rank()
  model = nn.Linear(10, 1)
  optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
  for epoch in range(3):
      x, y = torch.randn(64, 10), torch.randn(64, 1)
      loss = nn.MSELoss()(model(x), y)
      optimizer.zero_grad(); loss.backward(); optimizer.step()
      if rank == 0:
          print(f'Epoch {epoch+1}/3 - loss: {loss.item():.4f}')
  if rank == 0:
      print('Training complete.')
  dist.destroy_process_group()
  "
```

**Key change:** `--workers 4` becomes `worker.replicas: 3`. The master is an
additional Pod, making the total GPU-using replicas 4 (3 workers + 1 master).
The master inherits the worker's resource configuration by default since no
`master` block is specified.

### Example 2: TensorFlow with PS, Chief, and Evaluator

**v1 command:**

```bash
arena submit tfjob \
  --name tf-distributed \
  --image tensorflow/tensorflow:2.15.0 \
  --workers 4 \
  --gpus 2 \
  --ps 2 \
  --ps-cpu 4 \
  --ps-memory 16Gi \
  --chief \
  --chief-cpu 2 \
  --chief-memory 4Gi \
  --evaluator \
  --evaluator-gpus 1 \
  "python train.py"
```

**v2 YAML:**

```yaml
version: 0.1.0
name: tf-distributed
image: docker.io/tensorflow/tensorflow:2.15.0
framework:
  name: tensorflow
run: |
  python -c "
  import tensorflow as tf
  strategy = tf.distribute.MultiWorkerMirroredStrategy()
  with strategy.scope():
      model = tf.keras.Sequential([tf.keras.layers.Dense(1, input_shape=(10,))])
      model.compile(optimizer='sgd', loss='mse')
  x = tf.random.normal((128, 10))
  y = tf.random.normal((128, 1))
  for epoch in range(3):
      history = model.fit(x, y, epochs=1, verbose=0)
      print(f'Epoch {epoch+1}/3 - loss: {history.history[\"loss\"][0]:.4f}')
  print('Training complete.')
  "

worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: "2"  # Uncomment for GPU clusters

chief:
  resources:
    cpu: "2"
    memory: 4Gi

ps:
  replicas: 2
  resources:
    cpu: "4"
    memory: 16Gi

evaluator:
  resources:
    # nvidia.com/gpu: "1"  # Uncomment for GPU clusters
    cpu: "2"
    memory: 4Gi
```

**Key changes:**
- TFJob roles (`chief`, `ps`, `evaluator`) become top-level YAML blocks.
- `--ps 2` becomes `ps.replicas: 2`.
- `--chief` and `--evaluator` flags (booleans) become the presence of the
  `chief` and `evaluator` blocks respectively.
- Resource specifications move into per-role `resources` maps.
- For TFJob, `worker.replicas` is NOT decremented by 1 — the `--workers - 1`
  formula applies only to PyTorch. TFJob workers have no separate master role
  (the `chief` role serves that purpose and is declared independently).

### Example 3: MPI Training

**v1 command:**

```bash
arena submit mpijob \
  --name mpi-training \
  --image mpioperator/mpi-test:openmpi-4.1.0 \
  --workers 4 \
  --gpus 4 \
  "mpirun -np 16 python train.py"
```

**v2 YAML:**

```yaml
version: 0.1.0
name: mpi-training
image: docker.io/mpioperator/mpi-test:openmpi-4.1.0
framework:
  name: mpi
run: mpirun -np 16 python -c "print('Hello from rank', __import__('os').environ.get('OMPI_COMM_WORLD_RANK', '0'))"

worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: "4"  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"

launcher:
  resources:
    cpu: "1"
    memory: "2Gi"
```

**Key changes:**
- For MPIJob, `worker.replicas` maps directly from `--workers` with no
  subtraction — MPI has a separate `launcher` role, not a master.
- The `launcher` block is optional. If omitted, a default CPU-only launcher is
  used. When specified, it allows independent resource configuration for the
  launcher Pod.
- v2-only fields `framework.options.mounts_on_launcher`,
  `framework.options.run_launcher_as_worker`, and
  `framework.options.slots_per_worker` are available for advanced MPI
  configurations with no v1 equivalent.

---

## Additional Resources

- [YAML Schema Specification](yaml-schema.md) — Complete field reference,
  validation rules, and per-provider `worker.replicas` mapping.
- [Best Practices](best-practices.md) — Recommended patterns for job
  configuration, resource sizing, and operational workflows.
- [CLI Reference](cli-reference.md) — Full command and flag documentation for
  the `arena-v2` CLI.
- [Arena v2 KEP](https://github.com/kubeflow/arena/blob/v2/keps/arena-v2/README.md) —
  Design rationale, detailed component comparison, and graduation criteria.
- [CLI Flag Mapping Audit](https://github.com/kubeflow/arena/blob/v2/keps/arena-v2/cli-flag-mapping.md) —
  Complete v1 flag to v2 YAML coverage audit.
