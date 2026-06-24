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
shell: /bin/sh                       # Optional. Shell interpreter path. Default: /bin/sh. Empty and null values fall back to the default.
working_dir: /root                   # Optional. Container working directory (default: image default)

# ─── Task ───
framework:                           # Required. Object with provider config.
  name: pytorch                      # Required. (pytorch | mpi | tensorflow | horovod | ray | deepspeed)
  options:                            # Optional
    nproc_per_node: auto             # auto | gpu | cpu | int → torchrun --nproc-per-node

envs:                                # Optional. Common envs.
  NCCL_DEBUG: INFO
  NCCL_IB_DISABLE: "0"

# ─── Scale ───
worker:                             # Required for non-PyTorch frameworks
  replicas: 4                       # Required if the worker block is present. Must be > 0 if specified.
  envs:                             # Optional
    NCCL_DEBUG: DEBUG               # Merges with top-level envs (worker value overrides)
  resources:                        # Recommended, but not required. Pod resources will be unset if this block is omitted.
    nvidia.com/gpu: 8
    cpu: 32
    memory: 128Gi

# Other fields, such as launcher, chief, ps, or evaluator, may be available depending on the framework. 
# Settings for replicas, resources, and environment variables may vary accordingly. 
# Please refer to the framework-specific configuration documentation for more details.
master:                             # Only valid for PyTorch. Required if the worker block is omitted.
  replicas: 1                       # Optional. Master replicas must be 1.
  envs:                             # Optional
    NCCL_DEBUG: DEBUG               # Merges with top-level envs (master value overrides)
  resources:                        # Optional
    nvidia.com/gpu: 8               # Optional

# ─── Sync & Init ───
sync:                                # Optional
  - git: https://github.com/org/training-code.git   # git clone
    branch: main                     # Optional. Default: main.
    local_path: /code                # Required
    image: git-sync:v3.3.5           # Optional
    mounts:                          # Optional. Override storages by name
    - name: code                     # A new storage entry will be appended if no matching entry is found in the storages block
      tmp: 1Gi
      mount_path: /code

  - rsync: 10.88.29.56::backup/data/logoRecoTrain.zip   # rsync from a remote source
    local_path: /dataset            # Required
    image: rsync:v3.1.0-aliyun       # Optional
    mounts:                          # Optional. Override storages by name
    - name: dataset                  # Override the 'dataset' storage defined in storages
      mount_path: /dataset           # Mounted at /dataset instead of its default /data

  - hdfs: hdfs://namenode:8020/models/resnet # download from a remote HDFS path
    local_path: /models            # Required
    image: apache/hadoop:3.5.0       # Optional
    mounts:                          # Optional. Override storages by name
    - name: checkpoints
      mount_path: /models

# ─── storages ───
# inject into all pods
storages:                 # Optional
  - name: dataset         # Required
    mount_path: /data     # Required if the storage type is not shm.
    sub_path: /data       # Optional
    pvc: dataset-pvc      # Exactly one storage type is required. (pvc | shm | tmp | hostpath | configmap | secret)
  - name: checkpoints
    mount_path: /ckpts
    pvc: ckpt-pvc
  - name: shm
    shm: 64Gi             # Optional. Default: 2Gi
    # mount_path: /dev/shm # Optional. Default: /dev/shm
  - name: tmp
    tmp: 128Gi
    mount_path: /tmp
  - name: host
    hostpath: /runtime-mnt
    mount_path: /runtime
  - name: conf
    configmap: foo
    key: conf.yaml       # Optional. Omitting key mounts the entire ConfigMap as a folder.
    mount_path: /app/conf.yaml  # If key is specified, it represents the exact target file name; otherwise, it represents the directory name.
  - name: credentials
    secret: bar
    key: id_rsa          # Optional. Omitting key mounts the entire Secret as a folder.
    mount_path: /root/.ssh/id_rsa  # If key is specified, it represents the exact target file name; otherwise, it represents the directory name.

# ─── Scheduling (K8s-specific) ───
# inject into all pods
scheduling:                            # Optional
  priority: 100                        # Integer → pod spec priority field
  priority_class_name: high-priority   # String → pod spec priorityClassName field
  gang:                                # Gang scheduling (Volcano/coscheduling)
    enabled: false                          
  scheduler_name: default              # Custom scheduler (optional)
  queue: low-priority                  # Queue name (optional)
  node_selector:
    disktype: ssd
    zone: us-west-2a
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
  affinity:                          # Optional
    policy: spread                   # Optional. Default: none. (none | spread | binpack)
    constraint: preferred            # Optional. Default: preferred. (preferred | required)
    target: node                     # Required if policy is not none. (pod | node)
    rules:                           # Required if policy is not none.
    - weight: 1                      # Only for constraint: preferred
      # topology_key: nvidia.com/gpu   # Only for target: pod
      match_expressions:
      - key: foo
        operator: In
        values:
        - bar
      match_fields:                   # Only for target: node
      - key: foo
        operator: In
        values:
        - bar
      match_labels:
        a: b
      # namespaces: ["a", "b"]         # Only for target: pod (filter namespaces by name)
      # namespace_selector:            # Only for target: pod (filter namespaces by label)
      #   match_labels:
      #     team: platform
      #   match_expressions:
      #   - key: foo
      #     operator: In
      #     values:
      #     - bar

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

Fields under `worker.resources` are all **scalar** values, applied identically to both `requests` and `limits` (**Guaranteed QoS**).

```yaml
worker:
  resources:
    nvidia.com/gpu: 8
    cpu: 32
    memory: 128Gi
```

**Implementation**: When building the CRD, user-specified scalar values are written to both `requests` and `limits` in the Pod spec with identical values, ensuring K8s assigns the Guaranteed QoS class.

**Why Guaranteed only**: Training workloads are long-running, GPU-intensive jobs that require stable resource guarantees. Burstable QoS (requests < limits) leads to priority eviction under node contention, which is unsuitable for training scenarios.

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
4. **Security**: arena-v2 never reads, encodes, or transmits sensitive credentials. Secrets are managed by cluster administrators or external systems.

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

### envs Value Forms

`envs` values accept three forms:

| Form | YAML | Behavior |
| --- | --- | --- |
| Plain | `BATCH_SIZE: "32"` | Literal value in Pod spec |
| Secret ref | `DB_PASSWORD: {secret: name, key: k}` | `secretKeyRef` to existing K8s Secret |
| ConfigMap ref | `CONFIG_PATH: {configmap: name, key: k}` | `configMapKeyRef` to existing K8s ConfigMap |

### framework-specific configuration

`framework` accepts one form: an **object** where the `name` key holds the provider name, and the `options` key contains provider-specific settings.

Different training frameworks have different roles, and the default behaviors for replicas, envs, and resources vary accordingly. The corresponding configurations and descriptions are shown below.

```yaml
framework:
  name: pytorch
  options:
    nproc_per_node: auto             # auto | gpu | cpu | int → torchrun --nproc-per-node

worker:                              # Optional if the master block is specified.
  ...
master:                              # Required if the worker block is omitted. If the master block is omitted while the worker block is specified, it inherits the worker configuration (replicas fixed to 1). Otherwise, it uses its own configuration.
  ...
```

```yaml
framework:
  name: mpi
  options:
    slots_per_worker: 4              # MPI slots per worker node
    mounts_on_launcher: true         # Launcher also mounts PVCs (default: false)
    run_launcher_as_worker: false    # Optional: Launcher also runs as worker (default: false)

launcher:   # Optional. Defaults to a CPU-only configuration (replicas fixed to 1) if omitted. Uses its own configuration if specified.
  ...
worker:     # Required
  ...
```

```yaml
# Due to TensorFlow's diverse distributed strategies, roles (chief, worker, ps, evaluator) must be explicitly specified; no defaults are applied.
framework:
  name: tensorflow

worker:     # Required
  ...

chief:      # Optional. If this block is omitted, no pods will be created.
  ...

ps:         # Optional. If this block is omitted, no pods will be created.
  ...

evaluator:  # Optional. If this block is omitted, no pods will be created.
  ...

```

```yaml
framework:
  name: deepspeed

worker:     # Required
  ...
launcher:   # Optional. Defaults to a CPU-only configuration (replicas fixed to 1) if omitted. Uses its own configuration if specified.
  ...
```

```yaml
framework:
  name: ray # No detailed design for Ray — currently a placeholder
```

```yaml
framework:
  name: horovod

worker:     # Required
  ...
launcher:   # Optional. Defaults to a CPU-only configuration (replicas fixed to 1) if omitted. Uses its own configuration if specified.
  ...
```

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

**Invalid value handling**: `null` and empty string `""` both fall back to the default `/bin/sh`.

**Why not `sh -c "bash -c ..."` workaround:**
- **Signal delivery lost**: K8s sends SIGTERM to PID 1 (sh), which by default does not forward to the child bash process, breaking graceful shutdown
- **Unreliable exit codes**: sh without `set -e` may swallow non-zero exit codes from the bash child, causing K8s to mark the Pod as Succeeded instead of Failed
- **Nested escaping**: Quotes, `$`, and backslashes in user commands require double-escaping, which is error-prone

### Resource Lifecycle

A single `arena job run` invocation may create multiple K8s resources:

```
arena job run -f train.yaml
  ├─ PyTorchJob CRD              ← Primary resource (Operator manages Pod lifecycle)
  ├─ TensorBoard Deployment+Service ← If logging.tensorboard is enabled
  └─ ConfigMap (metadata)        ← Stores effective training job YAML
```

**Design choice: Each task type uses its corresponding top-level resource as the anchor, with all other resources setting ownerReferences to it.**

```
PyTorchJob CRD
  ├── owns → ConfigMap "my-job" (stores original YAML; a new CRD could replace the native ConfigMap)
  ├── owns → TensorBoard Deployment
  └── owns → TensorBoard Service
```

**Implementation approach:**

1. Create the primary CRD (e.g. PyTorchJob) first to obtain its UID.
2. Create the ConfigMap with the original YAML and generated manifest, setting `ownerReferences` to the CRD.
3. Create any additional resources (TensorBoard Deployment, Service) with `ownerReferences` pointing to the CRD.
4. On `arena job delete`, delete the CRD — K8s garbage collection cascades to all resources with matching `ownerReferences`.

**`arena job delete` implementation (single step):** Delete the primary CRD. K8s GC cascades deletion to all resources that reference it.

**`arena job get` implementation:** Read the ConfigMap by job name to retrieve the original YAML for display.

**Comparison with v1:**

| | v1 (Helm) | v2 (CRD anchor) |
|---|---|---|
| Metadata storage | ConfigMap `data.app` | ConfigMap `data.arena-v2.yaml` |
| Resource tracking | Resource list in ConfigMap | K8s ownerReferences (automatic) |
| `arena job delete` steps | 3 steps: read CM → delete resources → delete CM | **1 step: delete CRD** |
| Resource leak risk | High (list may be incomplete) | **None (K8s GC handles it)** |

**OwnerReferences semantics**: `kubectl get configmap -o yaml` will show ownerReferences relationships.

- **Training Operator is unaffected**: the reconciler reads CRD spec/status to manage Pods, not ownerReferences
- **Transparent to arena CLI users**: users interact via `arena job get/delete`, not direct CRD manipulation
- **Helm uses this pattern**: Helm release Secrets serve as anchors managing all chart resources — a mature K8s ecosystem pattern

### suspend / resume (Operational Pause)

Pause/resume are **operations**, not configuration — they do not belong in the YAML schema. When integrating with queue systems like Kueue, applying a label/annotation on the CRD or patching `runPolicy.suspend` is sufficient.

```bash
arena job suspend <name>      # Suspend task
arena job resume <name>    # Resume task
```

### worker Field Semantics

`worker` represents the **GPU-using worker nodes** in a training job. Its framework-specific mapping is shown below:

| Framework | worker mapping | Uses GPU | Notes |
|---|---|---|---|
| pytorch | Worker | Yes | Master defaults to use the same config as Worker except for replicas |
| tensorflow | Worker | Yes | Chief/PS/Evaluator do not inherit worker's config |
| mpi | Worker | Yes | Launcher does not use GPU, does not inherit worker's config |
| deepspeed | Worker | Yes | Launcher does not use GPU, does not inherit worker's config |
| horovod | Worker | Yes | Launcher does not use GPU, does not inherit worker's config |

**MPI/DeepSpeed/Horovod Launchers do not inherit worker's config by default** — the Provider automatically sets independent resource config for the Launcher (typically no GPU) during CRD generation, avoiding GPU waste on Launcher pods.

**PyTorch Master vs Worker**: Master runs as Worker. Master replica count is fixed at 1; **total GPU-using replicas** = `worker.replicas + 1`.

### sync & init (Code/Data Injection)

**`sync` — high-level data/code injection (sugar for initContainer):**

`sync` supports three source types:

| Type | Source | Init Container |
|---|---|---|
| `git` | Git repository URL | `git-sync` clone |
| `rsync` | Remote storage source | `rsync` |
| `hdfs` | HDFS path | `apache/hadoop` init container with `hdfs dfs -get` |

```yaml
sync:
  - git: https://github.com/org/training-code.git
    branch: main                     # Optional. Default: main.
    local_path: /workspace
    mounts:
    - name: code
      mount_path: /workspace
      sub_path: /

  - rsync: 10.88.29.56::backup/data/logoRecoTrain.zip    # rsync from a remote source
    local_path: /dataset
    mounts:
    - name: dataset
      mount_path: /dataset

  - hdfs: hdfs://namenode:8020/models/resnet
    local_path: /models
    mounts:
    - name: checkpoints
      mount_path: /models
```

Each entry creates an initContainer and volumes inherited from `storages`. The `mounts` field overrides these inherited configurations by name, and the final volumes are mounted at `mount_path` in both the initContainer and main container across all pods.

**`init` — generic initContainer injection:**

```yaml
init:
  - name: download-model
    image: busybox
    run: wget -O /data/model.bin https://example.com/model.bin
    mounts:
    - name: data
      mount_path: /data

  - name: prepare-dataset
    image: custom-etl:latest
    run: /prepare.sh
    shell: /bin/bash
```

`init[].run` has the same semantics as the top-level `run`: a command string executed via shell.
`init[].shell` specifies the interpreter path for init containers. If omitted, it inherits the top-level `shell` value (which itself defaults to `/bin/sh`).

Init containers generated by `sync` use exec form (CLI auto-generates the command without shell wrapping) — no `run`/`shell` fields needed.

`sync` and `init` can coexist. `sync` entries are processed before `init` entries.

### worker.replicas → Provider Mapping

```
User writes: worker.replicas: 4
  pytorch    → PyTorchJob  { Master(1) + Worker(4) }   = 5 pods
  mpi        → MPIJob      { Launcher(1) + Worker(4) }  = 5 pods
  tensorflow → TFJob       { Worker(4) }   = 4 pods

User writes: worker.replicas: 1
  pytorch    → PyTorchJob  { Master(1) + Worker(1) }   = 2 pods
  mpi        → MPIJob      { Launcher(1) + Worker(1) }  = 2 pods
  tensorflow → TFJob       { Worker(1) }   = 1 pods

User writes: worker.replicas: 0. Validation fails.
User omits: worker. Validation passes only if the framework is PyTorch and the master block is specified.
```

### Affinity Design Principles

`scheduling.affinity` uses an orthogonal `policy × constraint × target` design:

- **policy**: `none` | `spread` | `binpack` — expresses scheduling intent
- **constraint**: `preferred` | `required` — maps to K8s preferredDuringScheduling / requiredDuringScheduling
- **target**: `pod` | `node` — determines whether to generate podAffinity or nodeAffinity

`rules[]` is required when `policy` is not none. It supports full K8s `affinity` semantics.

**Design principles:**

- No additional fields introduced (e.g. `direction`, `match`, `scope`) — `policy`/`constraint`/`target` already orthogonally cover all semantics
- Users only need to understand three dimensions: policy (intent) + constraint (strength) + target (object)
- `rules[]` is a direct mapping of K8s native fields with no extra abstraction
