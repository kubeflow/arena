# Arena V1 → V2 Flag Audit

This document maps all arena v1 training CLI flags to arena v2 YAML schema coverage.

**Scope:** `arena submit` training subcommands  
**Source of truth:** arena v2 YAML Schema + arena v1 CLI

---

## 1. Coverage Summary

| Category | v2 Covered | v2 Missing |
|----------|-----------|------------|
| Identity | ✅ name, image, labels, annotations, namespace | — |
| Resources | ✅ cpus, memory, shm, nvidia.com/gpu, extended resources (`--device`) | — |
| Scheduling | ✅ node_selector, tolerations, priority, priority_class_name, gang, scheduler_name, affinity (policy/constraint), selector, toleration, queue | — |
| Data | ✅ storages (PVC) | ❌ hostPath, config-file (ConfigMap mount) |
| Env | ✅ envs (plain/secretKeyRef/configMapKeyRef), per-role independent declaration | — |
| Execution | ✅ restart, image_pull_policy, image_pull_secrets, shell, working_dir | — |
| Lifecycle | ✅ clean_pod_policy, active_deadline, ttl_after_finished, backoff_limit, success_policy | — |
| Sync | ✅ sync (git/rsync/hdfs) | — |
| Model | — | ❌ model-name, model-source |
| TFJob roles | — | ❌ ps, evaluator, chief independent config |
| MPIJob | ✅ mounts_on_launcher, gpu_topology, run_launcher_as_worker |— |
| PyTorch | ✅ nproc_per_node | — |
| Horovod | ✅ ssh_port | — |
| DeepSpeed | — | (provider not yet implemented) |
| Ray | — | ❌ entire Ray provider (Phase 3) |

**Legend:** ✅ = designed in arena v2 schema, ❌ = not in schema, — = N/A

---

## 2. Common Submit Flags (shared by most frameworks)

### 2a. Identity

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `--name` | string | required | `name` | `--name` | ✅ |
| `--image` | string | required | `image` | `--image` | ✅ |
| `--image-pull-policy` | string | `"Always"` | `image_pull_policy` | `--image-pull-policy` | ✅ |
| `--image-pull-secret` | string[] | `[]` | `image_pull_secrets` | `--image-pull-secret` | ✅ (reference-only, no auto-create) |

### 2b. Resources

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `--gpus` | int | `0` | `worker.resources.nvidia.com/gpu` | `--gpus` | ✅ |
| `--cpu` | string | `""` | `worker.resources.cpus` | `--cpus` | ✅ |
| `--memory` | string | `""` | `worker.resources.memory` | `--mem` | ✅ |
| `--device` | string[] | `[]` | `worker.resources.<name>` (flat) | `--device` | ✅ e.g. `--device hugepages-2Mi=32Gi` |
| `--workers` | int | `1` | `worker.replicas` | `--workers` | ✅ |

**Note:** v1 `--device vendor.com/device=count` maps to v2 flat `resources.vendor.com/device: count`.

### 2c. Scheduling

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `--selector` | string[] | `[]` | `scheduling.node_selector` | `--selector` | ✅ |
| `--toleration` | string[] | `[]` | `scheduling.tolerations` | `--toleration` | ✅ |
| `-p, --priority` | int | `0` | `scheduling.priority` | `--priority` | ✅ Pod spec priority field |
| `--priority-class-name` | string | `""` | `scheduling.priority_class_name` | `--priority-class-name` | ✅ Pod spec priorityClassName field |
| `--gang` | bool | `false` | `scheduling.gang` | `--gang` | ✅ |
| `--scheduler` | string | `""` | `scheduling.scheduler_name` | `--scheduler-name` | ✅ |
| `--affinity-policy` | string | `""` | `scheduling.affinity.policy` | `--affinity-policy` | ✅ `none` / `spread` / `binpack` |
| `--affinity-constraint` | string | `""` | `scheduling.affinity.constraint` | `--affinity-constraint` | ✅ `preferred` / `required` |
| `--hostNetwork` | bool | `false` | `host_network` | `--host-network` | ✅ (top-level runtime field, not under scheduling) |
| `--hostIPC` | bool | `false` | `host_ipc` | `--host-ipc` | ✅ (top-level runtime field, not under scheduling) |
| `--hostPID` | bool | `false` | `host_pid` | `--host-pid` | ✅ (top-level runtime field, not under scheduling) |
| `--queue` | string | `""` | `scheduling.queue` | — | ✅ (YAML field exists) |
| `--rdma` | bool | `false` | — | — | ❌ NOT PLANNED |

### 2d. Data and Volumes

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `-d, --data` | string[] | `[]` | `storages[].pvc` + `storages[].mount_path` | `-d, --data` | ✅ |
| `--data-dir` | string[] | `[]` | `storages[].hostpath` + `storages[].mount_path` | `-d, --data` | ✅ |
| `--config-file` | string[] | `[]` | — | — | ❌ ConfigMap file mount not in schema |

### 2e. Environment and Labels

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `-e, --env` | string[] | `[]` | `envs` | `-e, --env` | ✅ (plain, secretKeyRef, configMapKeyRef) |
| `-a, --annotation` | string[] | `[]` | `annotations` | `-a, --annotation` | ✅ |
| `-l, --label` | string[] | `[]` | `labels` | `-l, --label` | ✅ |

### 2f. Execution

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `--working-dir` | string | image default | `working_dir` | `--working-dir` | ✅ |
| `--shell` | string | `/bin/sh` | `shell` | `--shell` | ✅ (invalid values fallback to /bin/sh) |
| `--retry` | int | `0` | `lifecycle.backoff_limit` | `--backoff-limit` | ✅ (v1 `--retry` → v2 `lifecycle.backoff_limit`) |

### 2g. Model Registry

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--model-name` | string | `""` | — | ❌ NOT PLANNED (Phase 3+) |
| `--model-source` | string | `""` | — | ❌ NOT PLANNED (Phase 3+) |

---

## 3. Sync Code Flags

| v1 Flag | Type | Default | v2 YAML | Status |
|---------|------|---------|---------|--------|
| `--sync-mode` | string | `""` | `sync[].git/rsync/hdfs` (type key) | ✅ |
| `--sync-source` | string | `""` | `sync[].git/rsync/hdfs` (value) | ✅ |
| `--sync-image` | string | `""` | — | ❌ NOT PLANNED (sync uses fixed init container images per type) |

---

## 4. Tensorboard Flags

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `--tensorboard` | bool | `false` | `logging.tensorboard.enabled` | `--tensorboard` | ✅ |
| `--tensorboard-image` | string | `""` | `logging.tensorboard.image` | `--tensorboard-image` | ✅ |
| `--logdir` | string | `"/training_logs"` | `logging.tensorboard.logdir` | `--tensorboard-logdir` | ✅ |

---

## 5. Job Lifecycle (common across frameworks, varying defaults)

| v1 Flag | Type | Default | v2 YAML | v2 CLI Flag | Status |
|---------|------|---------|---------|-------------|--------|
| `--clean-task-policy` | string | varies | `lifecycle.clean_pod_policy` | `--clean-pod-policy` | ✅ (`None` / `Running` / `All`) |
| `--running-timeout` | duration | `0` | `lifecycle.active_deadline` | `--active-deadline` | ✅ |
| `--ttl-after-finished` | duration | `0` | `lifecycle.ttl_after_finished` | `--ttl-after-finished` | ✅ |
| `--backoff-limit` | int | `6` | `lifecycle.backoff_limit` | `--backoff-limit` | ✅ |
| `--success-policy` | string | varies | `lifecycle.success_policy` | `--success-policy` | ✅ (`ChiefWorker` / `AllWorkers`, TFJob only) |
| `--share-memory` | string | `"2Gi"` | `storages[].shm` | `--shm` | ✅ (emptyDir medium: Memory at /dev/shm) |

---

## 6. Framework-Specific Flags

### 6a. TFJob

| v1 Flag | v2 YAML | v2 CLI Flag | Status |
|---------|---------|-------------|--------|
| `--ps` (count) | `ps.replicas` | `--ps-count` | ✅ YAML-only |
| `--ps-image` | `ps.image` | — | ✅ YAML-only |
| `--ps-cpu`, `--ps-cpu-limit`, `--ps-memory`, `--ps-memory-limit` | `ps.resources` | — | ✅ YAML-only |
| `--ps-gpus` | `ps.resources.nvidia.com/gpu`) | — | ✅ YAML-only |
| `--ps-selector` | - | — | ❌ NOT IN SCHEMA |
| `--ps-affinity-policy` | - | — | ❌ NOT IN SCHEMA |
| `--chief` (bool) | `chief` | - | ✅ YAML-only |
| `--chief-cpu`, `--chief-memory`, etc. | chief.resources` | — | ✅ YAML-only |
| `--chief-selector` | - | — | ❌ NOT IN SCHEMA |
| `--evaluator` (bool) | `evaluator` | - | ✅ YAML-only |
| `--evaluator-*` | - | — | ❌ NOT IN SCHEMA |
| `--worker-image` | `worker.image` | — | ✅ YAML-only |
| `--worker-port` | — | — | ❌ NOT IN SCHEMA |
| `--worker-cpu`, `--worker-memory`, etc. | `worker.resources` | — | ✅ YAML-only |
| `--worker-selector` | - | — | ❌ NOT IN SCHEMA |
| `--worker-affinity-policy` | - | — | ❌ NOT IN SCHEMA |
| `--success-policy` | `lifecycle.success_policy` | `--success-policy` | ✅ (moved to lifecycle block) |
| `--role-sequence` | — | — | ❌ Implementation detail, not user-facing |

### 6b. PyTorchJob

| v1 Flag | v2 YAML | v2 CLI Flag | Status |
|---------|---------|-------------|--------|
| `--cpu` | `resources.cpus` | `--cpus` | ✅ |
| `--memory` | `resources.memory` | `--mem` | ✅ |
| `--nproc-per-node` | `framework.options.nproc_per_node` | `--nproc-per-node` | ✅ (`auto` / `gpu` / `cpu` / int) |

### 6c. MPIJob

| v1 Flag | v2 YAML | v2 CLI Flag | Status |
|---------|---------|-------------|--------|
| `--cpu` | `resources.cpus` | `--cpus` | ✅ |
| `--memory` | `resources.memory` | `--mem` | ✅ |
| `--gputopology` | `framework.options.gpu_topology` | `--gpu-topology` | ✅ |
| `--mounts-on-launcher` | `framework.options.mounts_on_launcher` | `--mounts-on-launcher` | ✅ |
| (no v1 flag) | `framework.options.run_launcher_as_worker` | — | ❌ CLI flag missing (YAML field exists) |

### 6d. Horovod

| v1 Flag | v2 YAML | v2 CLI Flag | Status |
|---------|---------|-------------|--------|
| `--ssh-port` | `framework.options.ssh_port` | — | ✅ (YAML field exists) |
| `--cpu` | `resources.cpus` | `--cpus` | ✅ |
| `--memory` | `resources.memory` | `--mem` | ✅ |

### 6e. DeepSpeed

| v1 Flag | v2 YAML | v2 CLI Flag | Status |
|---------|---------|-------------|--------|
| `--cpu` | `resources.cpus` | `--cpus` | ✅ |
| `--memory` | `resources.memory` | `--mem` | ✅ |
| `--launcher-selector` | `roles[]` (name: launcher, `scheduling.node_selector`) | — | ✅ YAML-only |
| `--job-restart-policy` | `restart` | `--restart` | ✅ |
| `--job-backoff-limit` | `lifecycle.backoff_limit` | `--backoff-limit` | ✅ |
| `--ssh-secret` | — | — | ❌ NOT IN SCHEMA (No ssh_secret is required because the mpi-operator already handles it) |
| `--launcher-annotation` | `roles[]` (name: launcher, `annotations`) | — | ✅ YAML-only |
| `--worker-annotation` | `roles[]` (name: worker, `annotations`) | — | ✅ YAML-only |

### 6f. Elastic Training (ETJob)

| v1 Flag | v2 YAML | v2 CLI Flag | Status |
|---------|---------|-------------|--------|
| `--max-workers` | — | — | ❌ NOT IN SCHEMA (Phase 3) |
| `--min-workers` | — | — | ❌ NOT IN SCHEMA (Phase 3) |
| `--spot-instance` | — | — | ❌ NOT IN SCHEMA (use PriorityClass) |
| `--max-wait-time` | — | — | ❌ NOT IN SCHEMA |
| `--launcher-selector` | `roles[]` (name: launcher, `scheduling.node_selector`) | — | ✅ YAML-only |
| `--job-restart-policy` | `restart` | `--restart` | ✅ |
| `--worker-restart-policy` | `roles[]` (name: worker, `restart`) | — | ✅ YAML-only |
| `--job-backoff-limit` | `lifecycle.backoff_limit` | `--backoff-limit` | ✅ |
| `--ssh-secret` | — | — | ❌ NOT IN SCHEMA |

### 6g. RayJob (Phase 3)

Not mapped here — will be designed when Ray provider is added.
