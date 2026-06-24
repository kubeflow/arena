
## YAML Schema

### Minimal Example

```yaml
version: 0.1.0
name: my-job
framework: 
  name: pytorch
image: pytorch/pytorch:2.1
run: python train.py --epochs 10
worker:
  replicas: 1
  resources:
    nvidia.com/gpu: 1
```

### Full Schema

```yaml
# ─── Schema Version ───
version: 0.1.0                       # Required. Schema version for forward compatibility.

# ─── Identity ───
name: llm-finetune                   # Required. K8s resource name.
namespace: ml-team                   # Optional. Default: kubeconfig context.
description: "experiment run"        # Optional
labels:                              # Optional
  team: platform
annotations:                         # Optional
  note: "experiment run"

image: nvcr.io/nvidia/pytorch:23.10  # Required. Container image.
run: torchrun train.py --epochs 10   # Required. Training command, executed via shell.
shell: /bin/sh                       # Optional. Shell interpreter path. Default: /bin/sh. Invalid values fall back to default.
working_dir: /root                   # Optional. Container working directory (default: image default)

# ─── Task ───
framework:                           # Required. String shorthand or object with provider config.
  name: pytorch                      # Required. Provider name as key (pytorch | tensorflow | mpi | deepspeed | ray)
  options:                            # Optional
    nproc_per_node: auto             # auto | gpu | cpu | int → torchrun --nproc-per-node

envs:                                # Optional. Common envs.
  NCCL_DEBUG: INFO
  NCCL_IB_DISABLE: "0"

# ─── Scale ───
worker:                             # Required
  replicas: 4                       # Required. Worker replicas must be greater than 0
  envs:                             # Optional
    NCCL_DEBUG: DEBUG               # Worker: inherited env + NCCL_DEBUG (override/merge)
  resources:                        # Required
    nvidia.com/gpu: 8               # Worker: manual specification required

# ─── Sync & Init ───
sync:                               # Optional
  - git: https://github.com/org/training-code.git   # git clone
    branch: main                     # Optional. Override default branch main
    local_path: /code                # Optional. Override default local_path
    mounts:                          # Optional. Override storages by name
    - name: code 
      mount_path: /workspace
      sub_path: /

  - rsync: 10.88.29.56::backup/data/logoRecoTrain.zip   # rsync from job's remote source
    local_path: /workdir             # Optional. Override default local_path
    mounts:                          # Optional
    - mount_path: /dataset
      name: dataset


  - hdfs: hdfs://namenode:8020/models/resnet # download from remote hdfs
    local_path: /workdir             # Optional. Override default local_path
    mounts:                          # Optional
    - mount_path: /models
      name: checkpoints  

# ─── storages ───
# inject into all pods
storages:                 # Optional
  - name: dataset         # Required if storages field is specified.
    mount_path: /data     # Required if storages field is specified and storage is not shm.
    sub_path: /data       # Optional
    pvc: dataset-pvc
  - name: checkpoints
    mount_path: /ckpts
    pvc: ckpt-pvc
  - name: shm
    shm: 64Gi
    # mount_path: /dev/shm # Optional. Default to /dev/shm
  - name: tmp
    tmp: 128Gi
    mount_path: /tmp
  - name: host
    hostpath: /runtime-mnt
    mount_path: /runtime

# ─── Scheduling (K8s-specific) ───
# inject into all pods
scheduling:                            # Optional
  priority: 100                        # Integer → pod spec priority field
  priority_class_name: high-priority   # String → pod spec priorityClassName field
  gang: false                          # Gang scheduling (Volcano/coscheduling)
  scheduler_name: default              # Custom scheduler (optional)
  queue: low-priority                  # Queue name (optional)
  node_selector:
    disktype: ssd
    zone: us-west-2a
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
  affinity:                          # future feature
    policy: spread                   # none | spread | binpack
    constraint: preferred            # preferred | required
    target: pod                      # pod | node
    rules:
    - topology_key: nvidia.com/gpu
      weight: 1
      match_expressions:
      - key: foo
        operator: In
        values:
        - bar
      match_fields:
      - key: foo
        operator: In
        values:
        - bar
      match_labels:
        a: b
      namespaces: ["a", "b"]
      namespace_selector:            # Only for target: pod (filter namespaces by label)
        match_labels:
          team: platform
        match_expressions:
        - key: foo
          operator: In
          values:
          - bar

# ─── Lifecycle ───
# inject into all pods 
lifecycle:                           # Optional
  clean_pod_policy: Running          # None | Running | All
  active_deadline: 2h                # Max active duration
  ttl_after_finished: 7d             # Auto-delete after this duration
  backoff_limit: 6                   # Max restart count
  success_policy: ChiefWorker        # ChiefWorker | AllWorkers (TFJob only)

# ─── Runtime ───
# inject into all pods 
image_pull_policy: Always            # Optional. Always | IfNotPresent | Never
image_pull_secrets:                  # Optional
  - registry-secret                  # Reference existing Secret
service_account: training-sa         # Optional
restart: OnFailure                   # Optional. Always | OnFailure | Never
host_network: false                  # Optional. true | false
host_ipc: false                      # Optional. true | false
host_pid: false                      # Optional. true | false

# ─── Logging ───
# create additional pods
logging:                             # Optional
  tensorboard:
    enabled: false
    logdir: /training_logs             # Required if tensorboard enabled. TensorBoard --logdir arg
    image: tensorflow/tensorflow:2.15  # Optional. Override default image
  # wandb:                             # future feature
  #   enabled: true
  #   project: my-project
  #   entity: my-team
```

### resources

Fields under `worker.resources` are all **Scalar** values, applied identically to both `requests` and `limits` (**Guaranteed QoS**).

```yaml
worker:
  resources:
    nvidia.com/gpu: 8
    cpu: 32
    memory: 128Gi
```

**Why Guaranteed only**: Training workloads are long-running, GPU-intensive jobs that require stable resource guarantees. Burstable QoS (requests < limits) leads to priority eviction under node contention, which is unsuitable for training scenarios.

**Implementation**: When building the CRD, scalar values are written to both `requests` and `limits` in the Pod spec with identical values, ensuring K8s assigns the Guaranteed QoS class.

### image_pull_secrets

`image_pull_secrets` references existing `imagePullSecret` resources by name:

| Form | YAML | Behavior |
|---|---|---|
| String | `- registry-secret` | Reference existing `imagePullSecret` by name |

```yaml
image_pull_secrets:
  - nvcr-registry                  # Must already exist in the namespace
  - gcr-creds
```

Users create imagePullSecrets externally:

```bash
kubectl create secret docker-registry nvcr-registry \
  --docker-server=nvcr.io \
  --docker-username=$NVCR_USER \
  --docker-password=$NVCR_TOKEN
```

### Secrets Design Principles

arena-v2 **does not create K8s Secrets** — it only references existing ones. Rationale:

1. **arena-v2 is a client-side tool** (like kubectl). Creating Secrets is a state-mutating operation that conflicts with the "CRD generation only" design.
2. **Standard K8s ecosystem practice**: External Secrets Operator, Vault Agent, Sealed Secrets, and similar tooling all operate around the reference model.
3. **v1 also lacks Secret creation** — v1 users are already accustomed to managing Secrets via `kubectl create secret` first.
4. **Security**: arena-v2 never handles sensitive credentials (read, encode, or transmit). Secrets are managed by cluster administrators or external systems.

**Secret reference via `envs` field (secretKeyRef)**:

```yaml
envs:
  HF_TOKEN:
    secret: my-hf-creds            # Reference existing K8s Secret "my-hf-creds"
    key: token
  WANDB_KEY:
    secret: my-hf-creds
    key: wandb-key
  DB_HOST:
    configmap: db-config
    key: host
```

This keeps the YAML schema pure (reference-only declarations) while providing a convenient Secret creation tool.

### envs Value Forms

`envs` values accept three forms:

| Form | YAML | Behavior |
| --- | --- | --- |
| Plain | `BATCH_SIZE: "32"` | Literal value in Pod spec |
| Secret ref | `DB_PASSWORD: {secret: name, key: k}` | `secretKeyRef` to existing K8s Secret |
| ConfigMap ref | `CONFIG_PATH: {configmap: name, key: k}` | `configMapKeyRef` to existing K8s ConfigMap |

### framework (Provider Selection & Config)

`framework` accepts one form: an **object** where the provider name is the key and its value holds provider-specific settings.


**Object form** — provider name as key, config as value:
```yaml
framework:
  name: pytorch
  options:
    nproc_per_node: auto             # auto | gpu | cpu | int → torchrun --nproc-per-node
```

```yaml
framework:
  name: mpi
  options:
    slots_per_worker: 4              # MPI slots per worker node
    mounts_on_launcher: true         # Launcher also mounts PVCs (default: false)
    run_launcher_as_worker: false    # Optional: Launcher also run as worker (default: false)
    gpu_topology: true               # GPU topology scheduling annotation
```

```yaml
framework:
  name: tensorflow
  options:
    ps_count: 2                      # Parameter Server replica count
    chief: true                      # Enable Chief (default: true for standalone)
    evaluator: false                 # Enable Evaluator
```

```yaml
framework:
  name: horovod
  options:
    ssh_port: 22                     # SSH port for MPI launcher
```

**Implementation**: `UnmarshalYAML` handles the object form, parsing into a unified `FrameworkConfig` struct. Only one provider key is allowed. Provider-specific fields are strongly typed per provider (no cross-provider field leakage).

**Mapping to replicas**: `framework.tensorflow.options.ps_count` creates PS replicas. `chief` and `evaluator` are booleans that enable/disable those role types entirely.

### lifecycle (Job Lifecycle Policies)

```yaml
lifecycle:
  clean_pod_policy: Running          # None | Running | All — which pods to clean after completion
  active_deadline: 2h                # Max wall-clock duration → activeDeadlineSeconds
  ttl_after_finished: 7d             # Auto-delete Job after this duration → ttlSecondsAfterFinished
  backoff_limit: 6                   # Max restart count → backoffLimit
  success_policy: ChiefWorker        # ChiefWorker | AllWorkers (TFJob only)
```

`clean_pod_policy` controls which pods the Operator deletes after job completion:
- `None` — keep all pods
- `Running` — delete only running pods (failed/succeeded pods remain for log inspection)
- `All` — delete all pods

**lifecycle vs restart field boundary:**
- `lifecycle` wraps the training-operator's `RunPolicy` (Job-level policies): `clean_pod_policy`, `backoff_limit`, `ttl_after_finished`, `active_deadline`, `success_policy`.
- `restart` wraps `ReplicaSpec.RestartPolicy` (Replica-level), which is **not part of** `lifecycle`. The top-level `restart` is injected as the default for all Replicas.
- Platform-level retries (exit-code-triggered, preemption re-queueing, resource retention) are **outside arena-v2's scope** — they require external controller + queue scheduler integration beyond a client-side tool. `restart` + `backoff_limit` already covers training-operator's pod-level restart.

### run and shell (Command Execution Model)

`run` is the training command string, executed via shell. `shell` specifies the interpreter path:

| `shell` value | K8s Pod Spec Generation | Use Case |
|---|---|---|
| Omitted (default `/bin/sh`) | `command: ["/bin/sh", "-c"]`, `args: ["<run>"]` | Most scenarios, consistent with v1 behavior |
| `/bin/bash` | `command: ["/bin/bash", "-c"]`, `args: ["<run>"]` | Requires bash features (`source`, `[[ ]]`, arrays, etc.) |

**Invalid value handling**: `null`, empty string `""`, and non-existent paths all fall back to the default `/bin/sh`. Exec form (without shell wrapping) is not supported — to call an executable directly, use its absolute path in `run`.

**Why not `sh -c "bash -c ..."` workaround:**
- **Signal delivery lost**: K8s sends SIGTERM to PID 1 (sh), which by default does not forward to the child bash process, breaking graceful shutdown
- **Unreliable exit codes**: sh without `set -e` may swallow non-zero exit codes from the bash child, causing K8s to mark the Pod as Succeeded instead of Failed
- **Nested escaping**: Quotes, `$`, and backslashes in user commands require double-escaping, which is error-prone

**Implementation**: `Shell` is a plain `string` type. The Provider validates during `BuildCRD`: only non-empty absolute paths starting with `/` are used; otherwise it falls back to `/bin/sh`.

### Resource Lifecycle

A single `arena job run` invocation may create multiple K8s resources:

```
arena job run -f train.yaml
  ├─ PyTorchJob CRD              ← Primary resource (Operator manages Pod lifecycle)
  ├─ TensorBoard Deployment+Service ← If logging.tensorboard is enabled
  └─ ConfigMap (metadata)        ← Stores original YAML + generated manifest
```

**Design choice: Each task type uses its corresponding top-level resource as the anchor, with all other resources setting ownerReference to it.**

```
PyTorchJob CRD
  ├── owns → ConfigMap "my-job" (stores original YAML; a new CRD could replace the native ConfigMap)
  ├── owns → TensorBoard Deployment
  └── owns → TensorBoard Service
```

**Implementation approach:**

1. Create the primary CRD (e.g. PyTorchJob) first to obtain its UID.
2. Create the ConfigMap with the original YAML and generated manifest, setting `ownerReference` to the CRD.
3. Create any additional resources (TensorBoard Deployment, Service) with `ownerReference` pointing to the CRD.
4. On `arena job delete`, delete the CRD — K8s garbage collection cascades to all resources with matching `ownerReference`.

**`arena job delete` implementation (single step):** Delete the primary CRD. K8s GC cascades deletion to all resources that reference it.

**`arena job get` implementation:** Read the ConfigMap by job name to retrieve the original YAML for display.

**Comparison with v1:**

| | v1 (Helm) | v2 (CRD anchor) |
|---|---|---|
| Metadata storage | ConfigMap `data.app` | ConfigMap `data.arena-v2.yaml` |
| Resource tracking | Resource list in ConfigMap | K8s ownerReference (automatic) |
| `arena job delete` steps | 3 steps: read CM → delete resources → delete CM | **1 step: delete CRD** |
| Resource leak risk | High (list may be incomplete) | **None (K8s GC handles it)** |

**OwnerReference semantics**: `kubectl get pytorchjob -o yaml` will show ownerReference relationships. This is semantically unconventional but:
- **Training Operator is unaffected**: the reconciler reads CRD spec/status to manage Pods, not ownerReferences
- **Transparent to arena CLI users**: users interact via `arena job get/delete`, not direct CRD manipulation
- **Helm validates this pattern**: Helm release Secrets serve as anchors managing all chart resources — a mature K8s ecosystem pattern

### stop / resume (Operational Pause)

Pause/resume are **operations**, not configuration — they do not belong in the YAML schema. When integrating with queue systems like Kueue, applying a label/annotation on the CRD is sufficient.

```bash
arena job stop <name>      # Pause task (via label/annotation)
arena job resume <name>    # Resume task
```

### worker Field Semantics

`worker` represents the **GPU-using nodes** in a training job. The meaning and mapping of `worker` varies by framework:

| Framework | worker mapping | Uses GPU | Notes |
|---|---|---|---|
| pytorch | Master + Worker | Yes | Master also uses GPU, same config as Worker |
| tensorflow | Worker | Yes | Chief/PS/Evaluator do not use worker's GPU config |
| mpi | Worker | Yes | Launcher does not use GPU, does not inherit worker's resource config |
| deepspeed | Worker | Yes | Launcher does not use GPU |
| horovod | Worker | Yes | Launcher does not use GPU |

**MPI/DeepSpeed/Horovod Launchers do not inherit worker's GPU resources by default** — the Provider automatically sets independent resource config for the Launcher (typically no GPU) during CRD generation, avoiding GPU waste on Launcher pods.

**PyTorch Master vs Worker**: Master uses the same `worker.resources` config (including GPU) as Worker. Master replicas is fixed at 1; Worker replicas = `worker.replicas - 1`.

### sync & init (Code/Data Injection)

**`sync` — high-level data/code injection (sugar for initContainer):**

`sync` supports three source types:

| Type | Source | Init Container |
|---|---|---|
| `git` | Git repository URL | `git-sync` clone |
| `rsync` | Remote storage source | `rsync` |
| `hdfs` | HDFS path | `hadoop/hadoop` init container with `hdfs get` |

```yaml
sync:
  - git: https://github.com/org/training-code.git
    branch: main                     # Optional. Default: default branch.
    mounts:
    - name: code
      mount_path: /workspace
      sub_path: /

  - rsync: 10.88.29.56::backup/data/logoRecoTrain.zip    # Rsync from job's remote source
    mounts:
    - name: dataset
      mount_path: /dataset

  - hdfs: hdfs://namenode:8020/models/resnet
    mounts:
    - name: checkpoints
      mount_path: /models
```

Each entry creates an initContainer + emptyDir volume. The emptyDir is mounted at `mount_path` in both the initContainer and the main container across all pods.

**`init` — generic initContainer injection:**

```yaml
init:
  - name: download-model
    image: busybox
    run: wget -O /data/model.bin https://example.com/model.bin

  - name: prepare-dataset
    image: custom-etl:latest
    run: /prepare.sh
    shell: /bin/bash                     # Optional, independent of top-level shell
```

`init[].run` has the same semantics as the top-level `run`: a command string executed via shell. Each init container can independently specify its own `shell` interpreter (defaults to inheriting the top-level `shell` value; falls back to `/bin/sh` if unset).

Init containers generated by `sync` use exec form (CLI auto-generates the command without shell wrapping) — no `run`/`shell` fields needed.

Init containers do not support `envs`, `secrets`, or `resources` configuration.

`sync` and `init` can coexist. `sync` entries are processed before `init` entries.

### worker.replicas → Provider Mapping

```
User writes: worker.replicas: 4
  pytorch    → PyTorchJob  { Master(1) + Worker(3) }   = 4 pods
  tensorflow → TFJob       { Chief(1)  + Worker(4) }   = 5 pods
  mpi        → MPIJob      { Launcher(1) + Worker(4) }  = 5 pods

User writes: worker.replicas: 1
  pytorch    → PyTorchJob  { Master(1) }   = 1 pods
  tensorflow → TFJob       { Chief(1)  + Worker(1) }   = 2 pods
  mpi        → MPIJob      { Launcher(1) + Worker(1) }  = 2 pods

User writes: worker.replicas: 0 (or omitted), invalid value
```

### CLI Override Mapping

All YAML schema fields that are expressible as CLI flags are listed below. Flag names strictly follow schema field names. Complex nested types (`sync`, `init`, `storages`, `scheduling.affinity`) are YAML-only.

The `arena job run` command supports three input modes:
1. **Pure YAML**: `arena job run -f train.yaml`
2. **YAML + overrides**: `arena job run -f train.yaml --gpus 8 --workers 4`
3. **Pure flags**: `arena submit pytorch --name job --image pytorch:2.1 --gpus 2 "python train.py"`, `arena submit pytorchjob --name job --image pytorch:2.1 --gpus 2 "python train.py"`

#### Identity

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--name` | string | `name` | Required |
| `-n, --namespace` | string | `namespace` | Default: kubeconfig context |
| `-l, --label` | stringSlice | `labels` | `KEY=VALUE` repeated |
| `-a, --annotation` | stringSlice | `annotations` | `KEY=VALUE` repeated |

#### Task

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--framework` | string | `framework` | Also first positional arg |
| `--image` | string | `image` | Required |
| `--working-dir` | string | `working_dir` | Container working directory |
| `--shell` | string | `shell` | Shell interpreter path (default: /bin/sh, invalid values fall back to default) |
| (trailing args after `--`) | string | `run` | The command to execute |

#### Scale

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--workers` | int | `worker.replicas` | Worker replica count |

#### Resources

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--gpus` | int | `worker.resources.nvidia.com/gpu` | GPU count |
| `--gpu-type` | string | (node selector hint) | Maps to `nvidia.com/gpu.product` label |
| `--cpus` | string | `worker.resources.cpus` | e.g. `"4"` |
| `--mem` | string | `worker.resources.memory` | e.g. `"16Gi"` |
| `--shm` | string | `storages` (shm type) | e.g. `"8Gi"`, creates emptyDir at /dev/shm |
| `--device` | stringSlice | `worker.resources.<name>=<count>` | Extended resources, e.g. `--device hugepages-2Mi=32Gi` |

#### Environment

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `-e, --env` | stringSlice | `envs` | `KEY=VALUE` repeated |

#### Data (storages)

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `-d, --data` | stringSlice | `storages` | `name:path:pvc` repeated |

#### Scheduling

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--priority` | int | `scheduling.priority` | Pod spec priority field |
| `--priority-class-name` | string | `scheduling.priority_class_name` | Pod spec priorityClassName field |
| `--gang` | bool | `scheduling.gang` | Gang scheduling |
| `--scheduler-name` | string | `scheduling.scheduler_name` | Custom scheduler |
| `--affinity-policy` | string | `scheduling.affinity.policy` | `none` / `spread` / `binpack` |
| `--affinity-constraint` | string | `scheduling.affinity.constraint` | `preferred` / `required` |
| `--selector` | stringSlice | `scheduling.node_selector` | `key=value` repeated |
| `--toleration` | stringSlice | `scheduling.tolerations` | `key:operator:effect` repeated |

#### Lifecycle

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--clean-pod-policy` | string | `lifecycle.clean_pod_policy` | `None` / `Running` / `All` |
| `--active-deadline` | string | `lifecycle.active_deadline` | Duration, e.g. `"2h"` |
| `--ttl-after-finished` | string | `lifecycle.ttl_after_finished` | Duration, e.g. `"7d"` |
| `--backoff-limit` | int | `lifecycle.backoff_limit` | Max restart count |
| `--success-policy` | string | `lifecycle.success_policy` | `ChiefWorker` / `AllWorkers` (TFJob only) |

#### Runtime

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--image-pull-policy` | string | `image_pull_policy` | `Always` / `IfNotPresent` / `Never` |
| `--image-pull-secret` | stringSlice | `image_pull_secrets` | Image pull secret names |
| `--service-account` | string | `service_account` | ServiceAccount name |
| `--restart` | string | `restart` | `Always` / `OnFailure` / `Never` |
| `--host-network` | bool | `host_network` | |
| `--host-ipc` | bool | `host_ipc` | |
| `--host-pid` | bool | `host_pid` | |

#### Logging

| Flag | Type | YAML Field | Notes |
|------|------|-----------|-------|
| `--tensorboard` | bool | `logging.tensorboard.enabled` | Enable TensorBoard sidecar |
| `--tensorboard-logdir` | string | `logging.tensorboard.logdir` | `--logdir` argument |
| `--tensorboard-image` | string | `logging.tensorboard.image` | Override TB container image |

#### Framework Config

Framework-specific flags are validated at runtime — only valid for the matching `--framework`.

| Flag | Type | YAML Field | Framework |
|------|------|-----------|-----------|
| `--nproc-per-node` | string | `framework.options.nproc_per_node` | pytorch |
| `--ps-count` | int | `framework.options.ps_count` | tensorflow |
| `--chief` | bool | `framework.options.chief` | tensorflow |
| `--evaluator` | bool | `framework.options.evaluator` | tensorflow |
| `--slots-per-worker` | int | `framework.options.slots_per_worker` | mpi |
| `--gpu-topology` | bool | `framework.options.gpu_topology` | mpi |
| `--mounts-on-launcher` | bool | `framework.options.mounts_on_launcher` | mpi |

#### Meta

| Flag | Type | Notes |
|------|------|-------|
| `-f, --file` | string | Path to task YAML file |
| `--dry-run` | bool | Print generated CRD without submitting |

#### YAML-Only Fields (not expressible as CLI flags)

These schema sections require YAML input due to nested complexity:

- `sync` — git/rsync/hdfs code injection
- `init` — generic initContainer injection
- `storages` — PVC/shm/tmp/hostpath storage declarations
- `scheduling.affinity` — full affinity rules with policy/constraint/target
- `framework.options` — when multiple provider config fields are needed simultaneously

### Affinity Design Principles

`scheduling.affinity` uses an orthogonal `policy × constraint × target` design:

- **policy**: `none` | `spread` | `binpack` — expresses scheduling intent
- **constraint**: `preferred` | `required` — maps to K8s preferredDuringScheduling / requiredDuringScheduling
- **target**: `pod` | `node` — determines whether to generate podAffinity or nodeAffinity

`rules[]` supports full K8s fields: `topology_key`, `match_labels`, `match_expressions`, `match_fields` (node only), `namespaces`, `namespace_selector` (pod only), `weight`.

**Design principles:**
- No additional fields introduced (e.g. `direction`, `match`, `scope`) — `policy`/`constraint`/`target` already orthogonally cover all semantics
- Users only need to understand three dimensions: policy (intent) + constraint (strength) + target (object)
- `rules[]` is a direct mapping of K8s native fields with no extra abstraction
