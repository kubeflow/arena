# YAML Configuration

## Overview

Arena v2 uses a single YAML file to define a training job. The file specifies the framework, container image, compute resources, storage, scheduling, and optional features like TensorBoard and code sync. Pass the file to the CLI with `arena job run -f <file.yaml>`.

## Minimal Example

The simplest job needs a version, name, framework, image, run command, and at least one replica:

```yaml
version: 0.1.0                          # Arena schema version
name: pytorch-example                   # Job name (becomes the K8s resource name)
image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime  # Container image
framework:
  name: pytorch                         # Training framework
  options:
    nproc_per_node: auto                # One process per GPU
worker:                                 # Worker replica configuration
  replicas: 2                           # Number of worker pods
  resources:
    # nvidia.com/gpu: "1"               # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
envs:
  NCCL_DEBUG: INFO                      # Environment variable for all pods
run: |                                  # Command executed in the container
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
      print('Distributed training complete.')
  dist.destroy_process_group()
  "
```

## Fields Reference

### Identity

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `version` | string | yes | `0.1.0` | Schema version in `MAJOR.MINOR.PATCH` format. If omitted, defaults to `0.1.0`. |
| `name` | string | yes | — | Job name. Used as the Kubernetes resource name. |
| `namespace` | string | no | kubeconfig context | Kubernetes namespace for the job. |
| `description` | string | no | — | Free-form description of the job. |
| `labels` | map | no | — | Labels applied to all created resources. |
| `annotations` | map | no | — | Annotations applied to all created resources. |
| `image` | string | yes | — | Container image to run. |

```yaml
version: 0.1.0
name: llm-finetune
namespace: ml-team
description: "Fine-tune Llama on custom data"
labels:
  team: platform
annotations:
  note: "experiment run"
image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
framework:
  name: pytorch
run: python -c "print('Hello from Arena v2')"
master:
  resources:
    cpu: "2"
    memory: "8Gi"
```

### Framework

| Field | Type | Required | Description |
|---|---|---|---|
| `framework.name` | string | yes | Training framework: `pytorch`, `tensorflow`, `mpi`, `deepspeed`, `horovod`, or `ray`. |
| `framework.options` | map | no | Framework-specific options. See [Framework-specific Options](#framework-specific-options). |

```yaml
framework:
  name: pytorch
  options:
    nproc_per_node: auto
```

### Replicas

Replica blocks define the pods that run your training job. Which blocks are valid depends on the framework (see [Framework-specific Options](#framework-specific-options)).

Each replica block supports the same sub-fields:

| Field | Type | Required | Description |
|---|---|---|---|
| `replicas` | int | yes (for `worker` and `ps`) | Number of pods. Must be > 0 for `worker`. Fixed to 1 for `master`, `chief`, `launcher`, and `evaluator`. |
| `resources` | map | no | Resource requests/limits (applied identically to both). Omit to leave pod resources unset. |
| `envs` | map | no | Per-role environment variables. Merged with top-level `envs`; role-level values override top-level. |

**Role blocks:**

| Block | Frameworks | Description |
|---|---|---|
| `worker` | all (required for non-PyTorch) | GPU-using worker pods. |
| `master` | pytorch | Master pod. If omitted while `worker` is present, inherits worker config with replicas fixed to 1. |
| `launcher` | mpi, deepspeed, horovod | Launcher pod. Defaults to a CPU-only config (replicas fixed to 1) if omitted. |
| `chief` | tensorflow | Chief pod. No defaults applied. |
| `ps` | tensorflow | Parameter server pods. Can have multiple replicas. |
| `evaluator` | tensorflow | Evaluator pod. No defaults applied. |

```yaml
worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: 8  # Uncomment for GPU clusters
    cpu: "32"
    memory: 128Gi
  envs:
    NCCL_DEBUG: DEBUG      # Overrides top-level envs for workers only

master:                     # PyTorch only
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters
  envs:
    ROLE: master
```

**Resources** values are scalars applied to both `requests` and `limits`. Specifying both `cpu` and `memory` gives the pod Guaranteed QoS.

```yaml
worker:
  resources:
    # nvidia.com/gpu: 8   # Uncomment for GPU clusters
    cpu: "32"               # CPU cores
    memory: 128Gi           # Memory
```

**Replica-to-pod mapping** (for `worker.replicas: 4`):

| Framework | Pods created |
|---|---|
| pytorch | Master(1) + Worker(4) = 5 pods |
| mpi | Launcher(1) + Worker(4) = 5 pods |
| tensorflow | Worker(4) = 4 pods |

### Run

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `run` | string | yes | — | Training command, executed via shell. |
| `shell` | string | no | `/bin/sh` | Shell interpreter path. Empty/null falls back to default. |
| `working_dir` | string | no | image default | Container working directory. |

```yaml
run: torchrun --nproc_per_node=1 python -c "import torch; print('torchrun OK, CUDA:', torch.cuda.is_available())"
shell: /bin/bash          # Use /bin/bash if you need bash features
working_dir: /workspace
```

Multi-line commands are supported:

```yaml
run: |
  pip install -e .
  python -c "import torch; print('Setup complete, torch:', torch.__version__)"
```

### Scheduling

All scheduling fields are injected into every pod.

| Field | Type | Default | Description |
|---|---|---|---|
| `scheduling.priority` | int | — | Pod priority value. |
| `scheduling.priority_class_name` | string | — | Kubernetes PriorityClass name. |
| `scheduling.gang.enabled` | bool | `false` | Enable gang scheduling (Volcano/coscheduling). |
| `scheduling.scheduler_name` | string | — | Custom scheduler name. |
| `scheduling.queue` | string | — | Queue name for batch schedulers. |
| `scheduling.node_selector` | map | — | Node label selector (same as K8s `nodeSelector`). |
| `scheduling.tolerations` | list | — | Pod tolerations (same as K8s `Toleration`). |
| `scheduling.affinity` | object | — | Affinity rules (see below). |

```yaml
scheduling:
  priority_class_name: high-priority
  gang:
    enabled: true
  node_selector:
    disktype: ssd
    zone: us-west-2a
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
```

**Affinity** uses three orthogonal dimensions:

| Dimension | Values | Description |
|---|---|---|
| `policy` | `none`, `spread`, `binpack` | Scheduling intent. Default: `none`. |
| `constraint` | `preferred`, `required` | Strength (preferred vs required). Default: `preferred`. |
| `target` | `pod`, `node` | Generate podAffinity or nodeAffinity. Required when `policy` is not `none`. |

`rules[]` is required when `policy` is not `none` and maps directly to K8s affinity rule fields:

| Rule field | Applies to | Description |
|---|---|---|
| `weight` | preferred constraint | Weight (1-100). |
| `topology_key` | target: pod | Topology domain key. |
| `match_labels` | both | Label match. |
| `match_expressions` | both | Label selector requirements (key, operator, values). |
| `match_fields` | target: node | Node field selector. |
| `namespaces` | target: pod | Filter namespaces by name. |
| `namespace_selector` | target: pod | Filter namespaces by label selector. |

```yaml
scheduling:
  affinity:
    policy: spread                   # Spread pods across topology domains
    constraint: preferred            # Best-effort (not hard requirement)
    target: node                     # Use nodeAffinity
    rules:
      - weight: 1
        match_fields:
          - key: metadata.name
            operator: In
            values:
              - node-1
```

### Storage

`storages` defines volumes mounted into all pods. Each entry requires a `name`, a `mount_path` (except for `shm`), and exactly one storage type.

| Storage type | Field | Key field | Description |
|---|---|---|---|
| PVC | `pvc` | — | Mount an existing PersistentVolumeClaim. |
| Shared memory | `shm` | — | `emptyDir` with medium `Memory`. Size is required (e.g. `64Gi`). Mount path defaults to `/dev/shm` if omitted. |
| Temporary | `tmp` | — | `emptyDir` with a size limit. |
| Host path | `hostpath` | — | Mount a host directory. |
| ConfigMap | `configmap` | `key` (optional) | Mount a ConfigMap. Omitting `key` mounts the entire ConfigMap as a directory. |
| Secret | `secret` | `key` (optional) | Mount a Secret. Omitting `key` mounts the entire Secret as a directory. |

When `key` is specified for ConfigMap/Secret, `mount_path` is the exact target file path. When `key` is omitted, `mount_path` is the target directory.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Volume name. |
| `mount_path` | string | yes (except `shm`) | Path in the container. |
| `sub_path` | string | no | Sub-path within the volume. |

```yaml
storages:
  - name: dataset                # PVC
    mount_path: /data
    pvc: dataset-pvc

  - name: checkpoints            # PVC for model checkpoints
    mount_path: /ckpts
    pvc: ckpt-pvc

  - name: shm                    # Shared memory (size required, mounts at /dev/shm by default)
    shm: 64Gi

  - name: tmp                    # Temporary emptyDir
    tmp: 128Gi
    mount_path: /tmp

  - name: host                   # Host path mount
    hostpath: /runtime-mnt
    mount_path: /runtime

  - name: conf                   # ConfigMap as a single file
    configmap: app-config
    key: conf.yaml
    mount_path: /app/conf.yaml

  - name: credentials            # Secret as a single file
    secret: ssh-creds
    key: id_rsa
    mount_path: /root/.ssh/id_rsa
```

### Sync

`sync` injects code or data into pods using init containers. Each entry creates an init container that runs before the main container. Three source types are supported:

| Type | Field | Init container | Description |
|---|---|---|---|
| Git | `git` | `git-sync` | Clone a Git repository. |
| Rsync | `rsync` | `rsync` | Sync from a remote rsync source. |
| HDFS | `hdfs` | `apache/hadoop` | Download from HDFS via `hdfs dfs -get`. |

| Field | Type | Required | Description |
|---|---|---|---|
| `git` / `rsync` / `hdfs` | string | exactly one | Source URL/path. |
| `branch` | string | no (git only) | Git branch. Default: `main`. |
| `local_path` | string | yes | Target path inside the container. |
| `image` | string | no | Override the init container image. |
| `mounts` | list | no | Override storage mount points by name. |

Each entry in `mounts` references a storage by `name` and can override its `mount_path` and `sub_path`. The referenced storage must be defined in `storages`.

```yaml
storages:
  - name: dataset
    mount_path: /data              # Default mount path
    pvc: dataset-pvc
  - name: code
    mount_path: /workspace
    tmp: 1Gi                       # emptyDir for synced code
  - name: checkpoints
    mount_path: /models
    pvc: ckpt-pvc

sync:
  - git: https://github.com/org/training-code.git
    branch: main
    local_path: /workspace
    mounts:
      - name: code                 # References storages entry
        mount_path: /workspace

  - rsync: 10.88.29.56::backup/data/dataset.zip
    local_path: /dataset
    mounts:
      - name: dataset              # Overrides storages mount_path
        mount_path: /dataset

  - hdfs: hdfs://namenode:8020/models/resnet
    local_path: /models
    mounts:
      - name: checkpoints
        mount_path: /models
```

### Init

`init` defines generic init containers for custom setup tasks. Each entry runs before the main container.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Container name. |
| `image` | string | yes | Container image. |
| `run` | string | yes | Command to execute (same shell semantics as top-level `run`). |
| `shell` | string | no | Shell interpreter. Inherits top-level `shell` if omitted (default `/bin/sh`). |
| `mounts` | list | no | Override storage mount points by name (same as `sync`). |

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

`sync` entries run before `init` entries. Both can coexist in the same job.

### TensorBoard

Enable a TensorBoard sidecar pod (Deployment + Service) to visualize training logs.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `logging.tensorboard.enabled` | bool | yes | `false` | Enable TensorBoard. |
| `logging.tensorboard.logdir` | string | yes (if enabled) | — | Directory containing event files (`--logdir` argument). |
| `logging.tensorboard.image` | string | no | default image | Override the TensorBoard container image. |
| `logging.tensorboard.mounts` | list | no | — | Selectively mount storages into the TensorBoard pod by name. |

```yaml
storages:
  - name: logs
    pvc: tensorboard-logs-pvc
    mount_path: /logs

logging:
  tensorboard:
    enabled: true
    logdir: /tb/logs
    mounts:                        # Only mount the "logs" storage in TensorBoard
      - name: logs
        mount_path: /tb/logs       # Override the mount path
```

### Environment

Environment variables can be set at the top level (applied to all pods) or per-role (merged with top-level, role values take precedence).

**Value forms:**

| Form | YAML | Behavior |
|---|---|---|
| Plain string | `BATCH_SIZE: "32"` | Literal value in the pod spec. |
| Secret ref | `HF_TOKEN: {secret: my-creds, key: token}` | `secretKeyRef` to an existing K8s Secret. |
| ConfigMap ref | `DB_HOST: {configmap: db-config, key: host}` | `configMapKeyRef` to an existing K8s ConfigMap. |

```yaml
envs:
  NCCL_DEBUG: INFO                           # Plain string
  HF_TOKEN:                                  # Secret reference
    secret: my-hf-creds
    key: token
  DB_HOST:                                   # ConfigMap reference
    configmap: db-config
    key: host
```

Arena does not create Secrets or ConfigMaps — it only references existing ones. Create them beforehand with `kubectl`:

```bash
kubectl create secret docker-registry nvcr-registry \
  --docker-server=nvcr.io \
  --docker-username=$NVCR_USER \
  --docker-password=$NVCR_TOKEN
```

### Lifecycle

Job-level lifecycle policies (injected into all pods).

| Field | Type | Default | Description |
|---|---|---|---|
| `lifecycle.clean_pod_policy` | string | — | Which pods to clean after completion: `None`, `Running`, or `All`. |
| `lifecycle.active_deadline` | string | — | Max wall-clock duration (e.g. `2h`). Maps to `activeDeadlineSeconds`. |
| `lifecycle.ttl_after_finished` | string | — | Auto-delete job after this duration (e.g. `7d`). Maps to `ttlSecondsAfterFinished`. |
| `lifecycle.backoff_limit` | int | — | Max restart count. Maps to `backoffLimit`. |
| `lifecycle.success_policy` | string | — | `ChiefWorker` or `AllWorkers` (TensorFlow only). |

```yaml
lifecycle:
  clean_pod_policy: Running        # Delete only running pods after completion
  active_deadline: 2h              # Kill job after 2 hours
  ttl_after_finished: 7d           # Auto-delete job 7 days after completion
  backoff_limit: 6                 # Max 6 restarts
```

### Runtime

Container runtime settings (injected into all pods).

| Field | Type | Default | Description |
|---|---|---|---|
| `image_pull_policy` | string | — | `Always`, `IfNotPresent`, or `Never`. |
| `image_pull_secrets` | list | — | Names of existing imagePullSecrets. |
| `service_account` | string | — | Service account name. |
| `restart` | string | — | `Always`, `OnFailure`, or `Never`. Replica-level restart policy. |
| `host_network` | bool | `false` | Use host network. |
| `host_ipc` | bool | `false` | Use host IPC namespace. |
| `host_pid` | bool | `false` | Use host PID namespace. |

```yaml
image_pull_policy: Always
image_pull_secrets:
  - nvcr-registry                  # Must already exist in the namespace
  - gcr-creds
service_account: training-sa
restart: OnFailure
host_network: false
host_ipc: false
host_pid: false
```

## Framework-specific Options

### PyTorch

| Option | Type | Default | Description |
|---|---|---|---|
| `nproc_per_node` | string | — | Processes per node: `auto`, `gpu`, `cpu`, or a positive integer. Maps to `torchrun --nproc-per-node`. |

**Roles:** `worker` (required if `master` is omitted) and `master` (optional).

- If `master` is omitted while `worker` is present, the master inherits the worker's image, resources, and envs (replicas fixed to 1).
- If `master` is specified, it uses its own configuration.
- Total GPU-using pods = `worker.replicas + 1` (the master).

```yaml
framework:
  name: pytorch
  options:
    nproc_per_node: auto

worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters

master:                             # Optional — inherits worker config if omitted
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters
  envs:
    ROLE: master
```

See: [examples/v2/quickstart/pytorch-simple.yaml](../examples/v2/quickstart/pytorch-simple.yaml), [examples/v2/quickstart/pytorch-standalone.yaml](../examples/v2/quickstart/pytorch-standalone.yaml), [examples/v2/reference/pytorch-mnist.yaml](../examples/v2/reference/pytorch-mnist.yaml)

### TensorFlow

TensorFlow uses explicit role blocks. No defaults are applied — only the roles you specify create pods.

**Roles:** `worker`, `chief`, `ps`, `evaluator`. At least one role block must be present.

| Option | Type | Description |
|---|---|---|
| `framework.options.ps_count` | int | Number of parameter servers (alternative to specifying `ps.replicas`). |

```yaml
framework:
  name: tensorflow

worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters

chief:                              # Runs the main training loop
  resources:
    # nvidia.com/gpu: 2  # Uncomment for GPU clusters

ps:                                 # Parameter servers (can have multiple replicas)
  replicas: 2
  resources:
    cpu: "4"
    memory: 16Gi

evaluator:                          # Runs evaluation
  resources:
    cpu: "2"
    memory: 4Gi
```

See: [examples/v2/quickstart/tensorflow-simple.yaml](../examples/v2/quickstart/tensorflow-simple.yaml), [examples/v2/reference/tf-with-roles.yaml](../examples/v2/reference/tf-with-roles.yaml)

### MPI

| Option | Type | Default | Description |
|---|---|---|---|
| `slots_per_worker` | int | — | MPI slots per worker node. |
| `mounts_on_launcher` | bool | `false` | Launcher also mounts PVCs. |
| `run_launcher_as_worker` | bool | `false` | Launcher also runs as a worker. |
| `gpu_topology` | bool | `false` | Enable GPU topology awareness. |
| `mpi_implementation` | string | — | `OpenMPI`, `Intel`, or `MPICH`. |
| `launcher_creation_policy` | string | — | `AtStartup` or `WaitForWorkersReady`. |
| `ssh_auth_mount_path` | string | — | SSH auth mount path. |

**Roles:** `worker` (required) and `launcher` (optional).

- The launcher does not use GPUs and does not inherit worker config.
- If `launcher` is omitted, it defaults to a CPU-only configuration (replicas fixed to 1).

```yaml
framework:
  name: mpi
  options:
    slots_per_worker: 4
    mpi_implementation: OpenMPI
    launcher_creation_policy: AtStartup

worker:
  replicas: 4
  resources:
    # nvidia.com/gpu: 4  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"

launcher:                           # Optional — defaults to CPU-only if omitted
  resources:
    cpu: "1"
    memory: "2Gi"
```

See: [examples/v2/quickstart/mpi-simple.yaml](../examples/v2/quickstart/mpi-simple.yaml), [examples/v2/reference/mpi-horovod-mnist.yaml](../examples/v2/reference/mpi-horovod-mnist.yaml)

### DeepSpeed

Same structure as MPI: `worker` (required) and `launcher` (optional, defaults to CPU-only). DeepSpeed requires GPUs at runtime.

```yaml
framework:
  name: deepspeed

worker:
  replicas: 2
  resources:
    nvidia.com/gpu: 1  # GPU required for DeepSpeed
```

See: [examples/v2/pretrain/deepspeed-bert.yaml](../examples/v2/pretrain/deepspeed-bert.yaml) — verified end-to-end on 2x Tesla T4.

### Horovod

Same structure as MPI: `worker` (required) and `launcher` (optional, defaults to CPU-only).

```yaml
framework:
  name: horovod

worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: 1  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
```

### Ray

Placeholder — no detailed configuration yet.

```yaml
framework:
  name: ray
```

### GPU-using Roles Summary

| Framework | GPU-using roles | Launcher/Master behavior |
|---|---|---|
| pytorch | master, worker | Master inherits worker config; replicas fixed to 1 |
| mpi | worker | Launcher is CPU-only; does not inherit worker config |
| tensorflow | worker, chief, evaluator | No inheritance between roles |
| deepspeed | worker | Launcher is CPU-only; does not inherit worker config |
| horovod | worker | Launcher is CPU-only; does not inherit worker config |

## Full Example

```yaml
version: 0.1.0
name: llm-finetune
namespace: ml-team
description: "Fine-tune Llama on custom data"
labels:
  team: platform

image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime
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
shell: /bin/bash
working_dir: /workspace

framework:
  name: pytorch
  options:
    nproc_per_node: auto

envs:
  NCCL_DEBUG: INFO
  HF_TOKEN:
    secret: my-hf-creds
    key: token

worker:
  replicas: 2
  resources:
    # nvidia.com/gpu: 8  # Uncomment for GPU clusters
    cpu: "2"
    memory: "8Gi"
  envs:
    NCCL_DEBUG: DEBUG             # Overrides top-level for workers

master:                            # Inherits worker config, overrides envs
  envs:
    ROLE: master

storages:
  - name: dataset
    mount_path: /data
    pvc: dataset-pvc
  - name: checkpoints
    mount_path: /ckpts
    pvc: ckpt-pvc
  - name: shm
    mount_path: /dev/shm
    shm: 64Gi
  - name: code
    mount_path: /workspace
    tmp: 1Gi

sync:
  - git: https://github.com/org/training-code.git
    branch: main
    local_path: /workspace
    mounts:
      - name: code
        mount_path: /workspace

scheduling:
  priority_class_name: high-priority
  gang:
    enabled: true
  node_selector:
    disktype: ssd
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule

lifecycle:
  clean_pod_policy: Running
  active_deadline: 24h
  ttl_after_finished: 7d
  backoff_limit: 3

image_pull_policy: Always
image_pull_secrets:
  - nvcr-registry
restart: OnFailure

logging:
  tensorboard:
    enabled: true
    logdir: /training_logs
    mounts:
      - name: checkpoints
        mount_path: /training_logs
```

## Examples

Example YAML files are available in the Arena repository under `examples/v2/`:

| File | Description |
|---|---|
| [pytorch-simple.yaml](../examples/v2/quickstart/pytorch-simple.yaml) | Minimal distributed PyTorch job with 2 workers (CPU-runnable). |
| [pytorch-standalone.yaml](../examples/v2/quickstart/pytorch-standalone.yaml) | PyTorch job using only a master role (no worker block). |
| [pytorch-mnist.yaml](../examples/v2/reference/pytorch-mnist.yaml) | PyTorch MNIST training with Master+Worker topology. |
| [pytorch-tensorboard-mounts.yaml](../examples/v2/reference/pytorch-tensorboard-mounts.yaml) | PyTorch job with TensorBoard and selective storage mounts. |
| [tensorflow-simple.yaml](../examples/v2/quickstart/tensorflow-simple.yaml) | Minimal TensorFlow job with 2 workers (CPU-runnable). |
| [tf-with-roles.yaml](../examples/v2/reference/tf-with-roles.yaml) | TensorFlow job with chief, PS, and evaluator roles. |
| [mpi-simple.yaml](../examples/v2/quickstart/mpi-simple.yaml) | Minimal MPI job with 2 workers (CPU-runnable). |
| [mpi-horovod-mnist.yaml](../examples/v2/reference/mpi-horovod-mnist.yaml) | Horovod distributed TensorFlow MNIST with custom launcher. |
| [deepspeed-bert.yaml](../examples/v2/pretrain/deepspeed-bert.yaml) | DeepSpeed BERT pre-training (GPU required, verified on 2x Tesla T4). |
