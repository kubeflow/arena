# Advanced Configuration

This guide covers the advanced Arena v2 configuration areas: scheduling, storage,
environment management, lifecycle, sync/init containers, and security. For the
full field reference, see [yaml-schema.md](yaml-schema.md); for operational
guidance, see [best-practices.md](best-practices.md).

## Scheduling

> **NOTE:** Scheduling fields exist in the YAML schema but are NOT YET IMPLEMENTED
> in the Alpha release. They will be available in Beta. YAML with scheduling
> config will parse but not apply to pods.

The `scheduling` block controls pod placement. It is optional and, when applied,
is injected into all pods created for the job.

| Field | Type | Description |
|---|---|---|
| `node_selector` | map[string]string | Simple key/value node selector applied to the pod spec. |
| `tolerations` | array | List of tolerations (see below). |
| `priority` | int | Integer written to the pod spec `priority` field. |
| `priority_class_name` | string | Name of an existing PriorityClass; written to `priorityClassName`. |
| `gang` | object | Gang scheduling with Volcano/coscheduling. `gang.enabled` (bool) turns it on. |
| `scheduler_name` | string | Custom scheduler name; written to `schedulerName`. |
| `queue` | string | Queue name (used with gang scheduling / queue systems). |
| `affinity` | object | Orthogonal affinity policy (see below). |

### tolerations

Each entry mirrors the Kubernetes toleration spec.

| Field | Type | Description |
|---|---|---|
| `key` | string | Taint key to tolerate. |
| `operator` | string | `Equal` or `Exists`. `Exists` must not have a `value`. |
| `value` | string | Taint value (only with `operator: Equal`). |
| `effect` | string | `NoSchedule`, `PreferNoSchedule`, or `NoExecute`. |
| `toleration_seconds` | int | Seconds to tolerate `NoExecute` before evicting. |

```yaml
scheduling:
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
    - key: dedicated
      operator: Equal
      value: training
      effect: NoSchedule
      toleration_seconds: 300
```

### gang

Gang scheduling requires a gang-aware scheduler (Volcano or the coscheduling
plugin). When `gang.enabled: true`, the full replica count across all roles is
advertised as `minAvailable` so the scheduler waits for the entire gang before
admitting pods.

```yaml
scheduling:
  gang:
    enabled: true
  queue: ml-team
  priority_class_name: high-priority
```

### affinity

`affinity` uses an orthogonal `policy x constraint x target` design (see
[yaml-schema.md](yaml-schema.md) for the design rationale):

- **policy** (`none` | `spread` | `binpack`): scheduling intent. Defaults to
  `none`. When not `none`, `rules[]` is required.
- **constraint** (`preferred` | `required`): strength. Maps to Kubernetes
  preferred/required scheduling. Defaults to `preferred`. `preferred` rules must
  set `weight` between 1 and 100.
- **target** (`pod` | `node`): generates podAffinity or nodeAffinity. Required
  when `rules[]` is present.
- **rules[]**: direct mapping of native Kubernetes affinity fields
  (`topology_key`, `weight`, `match_expressions`, `match_fields`,
  `match_labels`, `namespaces`, `namespace_selector`).

### Complete scheduling example

```yaml
scheduling:
  priority: 100
  priority_class_name: high-priority
  gang:
    enabled: false
  scheduler_name: default
  queue: low-priority
  node_selector:
    disktype: ssd
    zone: us-west-2a
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
  affinity:
    policy: spread
    constraint: preferred
    target: node
    rules:
      - weight: 100
        match_labels:
          accelerator: nvidia
        match_fields:
          - key: metadata.name
            operator: In
            values:
              - node-gpu-01
```

## Storage

The `storages` array declares volumes that are injected into all pods of the
job. Each entry must have a `name`, a `mount_path` (except `shm`, which defaults
to `/dev/shm`), and exactly one storage type. Storage names must be unique
within a task.

| Type | Field | Description |
|---|---|---|
| PVC | `pvc` | Reference an existing PersistentVolumeClaim by name. |
| Shared memory | `shm` | `emptyDir` with `medium: Memory`. Default size `2Gi`, default mount `/dev/shm`. |
| Temporary | `tmp` | `emptyDir` with the given `sizeLimit`. |
| Host path | `hostpath` | Mount a path from the host node. |
| ConfigMap | `configmap` | Mount a ConfigMap. Use `key` to mount a single file. |
| Secret | `secret` | Mount a Secret. Use `key` to mount a single file. |

For `configmap` and `secret`, the optional `key` field mounts a single key as a
file: `mount_path` becomes the exact target file name. Without `key`, the entire
ConfigMap/Secret is mounted as a directory and `mount_path` is the directory.

> **NOTE:** Arena v2 does not create Secrets or PVCs — it only references
> existing ones. Create them with `kubectl` (or an external secret system) first.

```yaml
storages:
  - name: data-vol
    pvc: my-pvc
    mount_path: /data
  - name: shm
    shm: {}
    mount_path: /dev/shm
  - name: config
    configmap: my-config
    mount_path: /etc/config
  - name: secret
    secret: my-secret
    key: token
    mount_path: /etc/secret/token
```

A more complete example combining several volume types:

```yaml
storages:
  - name: dataset
    mount_path: /data
    pvc: dataset-pvc
  - name: checkpoints
    mount_path: /ckpts
    pvc: ckpt-pvc
  - name: shm
    shm: 64Gi              # optional, default 2Gi
    # mount_path defaults to /dev/shm
  - name: tmp
    tmp: 128Gi
    mount_path: /tmp
  - name: host
    hostpath: /runtime-mnt
    mount_path: /runtime
  - name: conf
    configmap: app-config
    key: conf.yaml         # mount single file
    mount_path: /app/conf.yaml
  - name: credentials
    secret: ssh-creds
    key: id_rsa            # mount single file
    mount_path: /root/.ssh/id_rsa
```

## Environment Management

The `envs` map accepts three value forms. The same map is available at the
top level and per-role (`worker.envs`, `master.envs`, `launcher.envs`, etc.).
Worker-level (and other role-level) envs merge with top-level envs, with the
role-level value overriding the top-level value on key conflict.

| Form | YAML | Behavior |
|---|---|---|
| Plain | `NCCL_DEBUG: INFO` | Literal value in the pod spec. |
| Secret ref | `HF_TOKEN: {secret: hf-creds, key: token}` | `secretKeyRef` to an existing Secret. |
| ConfigMap ref | `DB_HOST: {configmap: db-config, key: host}` | `configMapKeyRef` to an existing ConfigMap. |

A mapping value must specify exactly one of `secret` or `configmap`, and the
`key` is required in both cases.

```yaml
envs:
  NCCL_DEBUG: INFO
  HF_TOKEN:
    secret: hf-creds
    key: token
  DB_HOST:
    configmap: db-config
    key: host

worker:
  replicas: 4
  resources:
    nvidia.com/gpu: 8
  envs:
    NCCL_DEBUG: DEBUG       # overrides top-level NCCL_DEBUG for workers
```

## Lifecycle Management

The `lifecycle` block wraps the training-operator `RunPolicy` (job-level
policies). It is optional.

| Field | Type | Description |
|---|---|---|
| `clean_pod_policy` | string | `None`, `Running`, or `All` — which pods the operator cleans after completion. |
| `active_deadline` | duration | Max active wall-clock time; auto-stops the job after the timeout. |
| `ttl_after_finished` | duration | Auto-delete the job this long after completion. |
| `backoff_limit` | int | Retry count on failure. |
| `success_policy` | string | `AllWorkers` or `ChiefWorker`. TFJob only. `ChiefWorker` is an alias for the default. |
| `suspend` | bool | Start the job suspended (or suspend via the CLI). |

Durations support standard Go format (`30s`, `5m`, `1h`, `1h30m`) plus a day
suffix (`7d`, `1.5d`).

`clean_pod_policy` controls post-completion pod cleanup:
- `None` — keep all pods.
- `Running` — delete only running pods (failed/succeeded pods remain for logs).
- `All` — delete all pods.

> **NOTE:** `lifecycle` covers job-level policies. The top-level `restart` field
> (Always | OnFailure | Never) is a separate, replica-level restart policy, not
> part of `lifecycle`.

Suspend and resume are operations, not configuration. Use the CLI to toggle them
at runtime:

```bash
arena job suspend <name>
arena job resume <name>
```

```yaml
lifecycle:
  clean_pod_policy: Running
  active_deadline: 7d
  ttl_after_finished: 1h
  backoff_limit: 6
  success_policy: AllWorkers   # TFJob only
```

## Init Containers & Sync

Arena v2 injects code and data through two mechanisms: high-level `sync` sources
and custom `init` containers. Both produce init containers that run before the
main training container.

### Sync Containers

`sync` is sugar for common data/code injection sources. Each entry requires
`local_path` (the target path inside the container) and exactly one of `git`,
`rsync`, or `hdfs`. An optional `image` overrides the default sync image, and an
optional `mounts` array overrides storage mount points by name.

| Source | Field | Init container |
|---|---|---|
| Git | `git` | `git-sync` clone. `branch` defaults to `main`. |
| rsync | `rsync` | rsync from a remote source. |
| HDFS | `hdfs` | `hdfs dfs -get` download. |

```yaml
sync:
  - git: https://github.com/org/training-code.git
    branch: main              # optional, default: main
    local_path: /workspace    # required
    image: git-sync:v3.3.5    # optional
    mounts:
      - name: code
        mount_path: /workspace
```

### Init Containers

`init` declares custom init containers. Each requires `name`, `image`, and
`run`; `run` is a command string executed via shell, with the same semantics as
the top-level `run`. The optional `shell` field specifies the interpreter path
(it inherits the top-level `shell`, which defaults to `/bin/sh`). The optional
`mounts` array selects storages to mount.

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

### Execution Order

1. System sync containers, named `arena-sync-0`, `arena-sync-1`, ... in
   declaration order.
2. User `init` containers, in declaration order.

`sync` and `init` can coexist in the same task. Sync containers generated from
`sync` entries use exec form (the CLI auto-generates the command without shell
wrapping); user `init` containers use the `run`/`shell` form.

### Mount Resolution

Volume mounts for sync and init containers follow a three-case contract based on
the task's `storages` and the container's optional `mounts` array:

| Scenario | Result |
|---|---|
| No `storages` | No volumes are mounted. |
| `storages` present, no `mounts` field | All storages are mounted (at their declared `mount_path`). |
| `storages` present, with `mounts` field | Only the listed storages are mounted; `mount_path`/`sub_path` in the mount override the storage defaults. |

When a `mounts` entry references a storage by `name`, the mount's `mount_path`
takes precedence and falls back to the storage's `mount_path` if empty. A mount
referencing a name not present in `storages` causes a validation error.

### Example: git sync + custom init container

```yaml
storages:
  - name: code
    tmp: 5Gi
    mount_path: /workspace
  - name: data
    pvc: training-data-pvc
    mount_path: /data

sync:
  - git: https://github.com/org/training-code.git
    branch: main
    local_path: /workspace
    mounts:
      - name: code
        mount_path: /workspace

init:
  - name: download-model
    image: busybox
    run: wget -O /data/model.bin https://example.com/model.bin
    mounts:
      - name: data
        mount_path: /data
```

## Security Considerations

### TensorBoard Has No Authentication

When TensorBoard is enabled (`logging.tensorboard.enabled: true`), Arena prints
the following warning at submit time:

> TensorBoard has no built-in authentication; the UI will be accessible to any
> pod with network access to this namespace.

The TensorBoard Deployment and Service are created with owner references to the
training job CRD (so they are garbage-collected on `arena job delete`). The
Service is only exposed within the namespace. Do not expose TensorBoard
externally without additional protection such as an authenticating Ingress or a
NetworkPolicy.

### Plaintext Secret Warning

Arena inspects env var names against common secret-like patterns
(`token`, `key`, `secret`, `password`, `credential`, case-insensitive). When a
matching env var is set to a plaintext value instead of a `secretKeyRef`, Arena
warns:

> env var looks like a secret but is stored in plaintext in ConfigMap — use
> secretKeyRef instead.

To avoid the warning and the security risk, reference an existing Secret:

```yaml
envs:
  HF_TOKEN:
    secret: hf-creds
    key: token
```

### System Namespace Warning

Arena warns when creating resources in Kubernetes system namespaces
(`kube-system`, `kube-public`, `kube-node-lease`):

> creating resources in system namespace — ensure this is intentional

Ensure that targeting a system namespace is intentional; user workloads normally
belong in a dedicated namespace.

### Service Account

Use the top-level `service_account` field to specify a custom service account
for the pods. Arena v2 does not create service accounts — reference an existing
one. The value is applied to the pod spec as `serviceAccountName`.

```yaml
service_account: training-sa
```
