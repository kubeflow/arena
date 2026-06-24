# Arena V1 Flag → V2 YAML Schema Audit

This document maps all arena v1 training CLI flags to arena v2 YAML schema coverage.

**Scope:** v1 `arena submit` flags → v2 `arena job run` YAML schema
**Source of truth:** arena v2 YAML Schema + arena v1 CLI

> **Note:** v2 CLI flag-to-YAML override mechanism will be designed separately as a generic `--set`-style approach (similar to Helm). This document tracks v1 flag → v2 YAML schema coverage only.

---

## 1. Coverage Summary

| Category | v2 Covered | v2 Missing / Notes |
|----------|-----------|------------|
| Identity | ✅ name, labels, annotations, namespace | — |
| Resources | ✅ cpu, memory, nvidia.com/gpu, extended resources (`--device`) | — |
| Scheduling | ✅ node_selector, tolerations, priority, priority_class_name, gang, scheduler_name, affinity (policy/constraint/target/rules), queue | not yet implemented |
| Data | ✅ storages (PVC, hostPath, tmp, shm, configmap, secret) | — |
| Env | ✅ envs (plain/secretKeyRef/configMapKeyRef) | — |
| Execution | ✅ restart, image_pull_policy, image_pull_secrets, shell, working_dir | — |
| Lifecycle | ✅ clean_pod_policy, active_deadline, ttl_after_finished, backoff_limit, success_policy | — |
| Sync | ✅ sync (git/rsync/hdfs) | — |
| Model | ❌ | — |
| TFJob roles | ✅ | provider not yet implemented |
| MPIJob | ✅ mounts_on_launcher, run_launcher_as_worker, slots_per_worker | — |
| PyTorch | ✅ nproc_per_node | — |
| Horovod | ✅ | provider not yet implemented |
| DeepSpeed | ✅ | provider not yet implemented |
| Ray | ✅ (framework placeholder only) | provider not yet implemented |

**Legend:** ✅ = designed in arena v2 schema, ❌ = not in schema, — = N/A

---

## 2. Common Job Run Flags (shared by most frameworks)

### 2a. Identity

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--name` | string | required | `name` | ✅ |
| `--namespace` | string | optional | `namespace` | ✅ |

### 2b. Resources

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--gpus` | int | `0` | `nvidia.com/gpu` field in `worker.resources` block | ✅ |
| `--cpu` | string | `""` | `worker.resources.cpu` | ✅ |
| `--memory` | string | `""` | `worker.resources.memory` | ✅ |
| `--device` | string[] | `[]` | `<device-name>` field in `worker.resources` block | ✅ e.g. `--device hugepages-2Mi=32Gi` |
| `--workers` | int | `1` | `worker.replicas` | ✅ (see the [Migration Notes](README.md#migration-notes) section) |

**Note:** v1 `--device vendor.com/device=count` maps to v2 flat `vendor.com/device: count` field in `resources` block.

### 2c. Scheduling

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--selector` | string[] | `[]` | `scheduling.node_selector` | ✅ |
| `--toleration` | string[] | `[]` | `scheduling.tolerations` | ✅ |
| `-p, --priority` | int | `0` | `scheduling.priority` | ✅ Pod spec priority field |
| `--priority-class-name` | string | `""` | `scheduling.priority_class_name` | ✅ Pod spec priorityClassName field |
| `--gang` | bool | `false` | `scheduling.gang.enabled` | ✅ |
| `--scheduler` | string | `""` | `scheduling.scheduler_name` | ✅ |
| `--affinity-policy` | string | `""` | `scheduling.affinity.policy` | ✅ `none` / `spread` / `binpack` |
| `--affinity-constraint` | string | `""` | `scheduling.affinity.constraint` | ✅ `preferred` / `required` |
| `--queue` | string | `""` | `scheduling.queue` | ✅ (YAML field exists) |
| `--rdma` | bool | `false` | — | ❌ NOT PLANNED |

### 2d. Data and Volumes

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `-d, --data` | string[] | `[]` | `storages[].pvc` + `storages[].mount_path` | ✅ |
| `--data-dir` | string[] | `[]` | `storages[].hostpath` + `storages[].mount_path` | ✅ |
| `--config-file` | string[] | `[]` | `storages[].configmap + storages[].mount_path` | ✅ |
| `--share-memory` | string | `"2Gi"` | `storages[].shm` | ✅ (emptyDir medium: Memory at /dev/shm) |

### 2e. Environment and Labels

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `-e, --env` | string[] | `[]` | `envs` | ✅ (plain, secretKeyRef, configMapKeyRef) |
| `-a, --annotation` | string[] | `[]` | `annotations` | ✅ |
| `-l, --label` | string[] | `[]` | `labels` | ✅ |

### 2f. Execution

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--working-dir` | string | image default | `working_dir` | ✅ |
| `--shell` | string | `/bin/sh` | `shell` | ✅ (invalid values fall back to /bin/sh) |
| `--image` | string | required | `image` | ✅ |
| `--image-pull-policy` | string | `"Always"` | `image_pull_policy` | ✅ |
| `--image-pull-secret` | string[] | `[]` | `image_pull_secrets` | ✅ (reference-only, no auto-create) |
| `--hostNetwork` | bool | `false` | `host_network` | ✅ |
| `--hostIPC` | bool | `false` | `host_ipc` | ✅ |
| `--hostPID` | bool | `false` | `host_pid` | ✅ |

### 2g. Model Registry

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--model-name` | string | `""` | — | ❌ NOT PLANNED |
| `--model-source` | string | `""` | — | ❌ NOT PLANNED |

---

## 3. Sync Code Flags

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--sync-mode` | string | `""` | `sync[].git/rsync/hdfs` (type key) | ✅ |
| `--sync-source` | string | `""` | `sync[].git/rsync/hdfs` (value) | ✅ |
| `--sync-image` | string | `""` | `sync[].<git/rsync/hdfs>.image` | ✅ |

---

## 4. Tensorboard Flags

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--tensorboard` | bool | `false` | `logging.tensorboard.enabled` | ✅ |
| `--tensorboard-image` | string | `""` | `logging.tensorboard.image` | ✅ |
| `--logdir` | string | `"/training_logs"` | `logging.tensorboard.logdir` | ✅ |

---

## 5. Job Lifecycle

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--clean-task-policy` | string | varies | `lifecycle.clean_pod_policy` | ✅ (`None` / `Running` / `All`) |
| `--running-timeout` | duration | `0` | `lifecycle.active_deadline` | ✅ |
| `--ttl-after-finished` | duration | `0` | `lifecycle.ttl_after_finished` | ✅ |
| `--job-backoff-limit` | int | `6` | `lifecycle.backoff_limit` | ✅ (use v2 `lifecycle.backoff_limit` default value) |
| `--retry` (redundant, functions identically to `--job-backoff-limit`) | int | `0` | `lifecycle.backoff_limit` | ✅ (use v2 `lifecycle.backoff_limit` default value) |
| `--success-policy` | string | varies | `lifecycle.success_policy` | ✅ (`ChiefWorker` / `AllWorkers`, TFJob only) |

---

## 6. Framework-Specific Flags

### 6a. TFJob

| v1 Flag | v2 YAML | Status |
|---------|---------|--------|
| `--ps` (count) | `ps.replicas` | ✅ |
| `--ps-image` | — | ❌ not yet implemented |
| `--ps-cpu`, `--ps-cpu-limit`, `--ps-memory`, `--ps-memory-limit` | `ps.resources` | ✅ |
| `--ps-gpus` | `ps.resources.nvidia.com/gpu` | ✅ |
| `--ps-selector` | — | ❌ not yet implemented |
| `--ps-affinity-policy` | — | ❌ not yet implemented |
| `--chief` (bool) | `chief` | ✅ |
| `--chief-cpu`, `--chief-memory`, etc. | `chief.resources` | ✅ |
| `--chief-selector` | — | ❌ not yet implemented |
| `--evaluator` (bool) | `evaluator` | ✅ |
| `--evaluator-cpu`, `--evaluator-memory`, etc | `evaluator.resources` | ✅ |
| `--evaluator-selector` | — | ❌ not yet implemented |
| `--worker-image` | — | ❌ not yet implemented |
| `--worker-port` | — | ❌ not yet implemented |
| `--worker-cpu`, `--worker-memory`, etc. | `worker.resources` | ✅ |
| `--worker-selector` | — | ❌ not yet implemented |
| `--worker-affinity-policy` | — | ❌ not yet implemented |
| `--success-policy` | `lifecycle.success_policy` | ✅ (moved to lifecycle block) |
| `--role-sequence` | — | ❌ Implementation detail, not user-facing |

### 6b. PyTorchJob

| v1 Flag | v2 YAML | Status |
|---------|---------|--------|
| `--cpu` | worker and master `resources.cpu` | ✅ |
| `--memory` | worker and master `resources.memory` | ✅ |
| `--nproc-per-node` | `framework.options.nproc_per_node` | ✅ (`auto` / `gpu` / `cpu` / int) |

### 6c. MPIJob

| v1 Flag | v2 YAML | Status |
|---------|---------|--------|
| `--cpu` | `worker.resources.cpu` | ✅ |
| `--memory` | `worker.resources.memory` | ✅ |
| `--gputopology` | `worker.resources`, `labels` | ✅ |
| `--mounts-on-launcher` | `framework.options.mounts_on_launcher` | ✅ |
| (no v1 flag) | `framework.options.run_launcher_as_worker` | ✅ (YAML field exists, no v1 equivalent) |
| (no v1 flag) | `framework.options.slots_per_worker` | ✅ (YAML field exists, no v1 equivalent) |

### 6d. Horovod

| v1 Flag | v2 YAML | Status |
|---------|---------|--------|
| `--ssh-port` | — | ❌ NOT IN SCHEMA (always use port 22) |
| `--cpu` | `worker.resources.cpu` | ✅ |
| `--memory` | `worker.resources.memory` | ✅ |

### 6e. DeepSpeed

| v1 Flag | v2 YAML | Status |
|---------|---------|--------|
| `--cpu` | `worker.resources.cpu` | ✅ |
| `--memory` | `worker.resources.memory` | ✅ |
| `--launcher-selector` | — | ❌ not yet implemented |
| `--job-restart-policy` | `restart` | ✅ |
| `--ssh-secret` | — | ❌ NOT IN SCHEMA (No ssh secret is required because the MPI Operator already handles it) |
| `--launcher-annotation` | — | ❌ not yet implemented |
| `--worker-annotation` | — | ❌ not yet implemented |

### 6f. RayJob

Not mapped here — will be designed when Ray provider is added.

### 6g. Elastic Training (ETJob)

Not mapped here and not planned.
