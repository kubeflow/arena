# CLI Reference

Arena v2 is a lightweight CLI for submitting AI training jobs to Kubernetes.

## Global Flags

These flags are available on all commands.

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--kubeconfig` | | string | `""` | Path to kubeconfig file |
| `--context` | | string | `""` | kubeconfig context to use |
| `--namespace` | `-n` | string | `""` | Kubernetes namespace (priority: flag > YAML > kubeconfig context > default) |
| `--debug` | | bool | `false` | Enable debug mode with detailed error output |
| `--verbose` | `-v` | int | `0` | Verbosity level (higher = more detailed logs) |

---

## arena job

Parent command for managing training jobs. All subcommands inherit the following persistent flag.

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--output` | `-o` | string | `table` | Output format: `table`, `wide`, `json`, `yaml` |

Supported frameworks: `pytorch`, `tensorflow`, `mpi`, `horovod`, `deepspeed`.

---

## arena job run

Submit a training job to Kubernetes from a YAML specification file.

### Syntax

```
arena job run -f <file> [flags]
```

### Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--file` | `-f` | string | `""` | Path to YAML file (required) |
| `--dry-run` | | bool | `false` | Print CRD as JSON or YAML without submitting |
| `--set` | | stringArray | `nil` | Override YAML field (Helm-style: `key=value`, repeatable) |

Also inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
# Submit a PyTorch job from a YAML file
$ arena job run -f examples/v2/quickstart/pytorch-simple.yaml

# Override a field with --set
$ arena job run -f examples/v2/quickstart/pytorch-simple.yaml --set worker.replicas=2

# Dry-run: print the generated CRD without submitting
$ arena job run -f examples/v2/quickstart/pytorch-simple.yaml --dry-run
```

### Output

```text
Job pytorch-simple submitted successfully
```

With `--dry-run`, the generated CRD is printed as indented JSON (default) or YAML (`-o yaml`). If TensorBoard is enabled, the Deployment and Service resources are also printed, separated by `---`.

---

## arena job list

List all training jobs across PyTorchJob, TFJob, and MPIJob CRD kinds.

### Syntax

```
arena job list [flags]
```

### Flags

This command has no flags of its own. It inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
# List all jobs in the default namespace
$ arena job list

# List jobs in a specific namespace
$ arena job list -n kubeflow

# Wide output with additional columns
$ arena job list -o wide

# JSON output
$ arena job list -o json
```

### Output

Default table format:

```text
NAME              STATUS     REPLICAS   AGE
pytorch-simple    Running    1/1        5m
tf-distributed    Succeeded  3/3        1h
mpi-test          Pending    0/2        10s
```

Wide format (`-o wide`) adds `NAMESPACE`, `APIVERSION`, `FRAMEWORK`, and `GPU` columns:

```text
NAME              NAMESPACE   STATUS     APIVERSION          FRAMEWORK   GPU   REPLICAS   AGE
pytorch-simple    default     Running    kubeflow.org/v1     pytorch     1     1/1        5m
tf-distributed    default     Succeeded  kubeflow.org/v1     tensorflow  0     3/3        1h
mpi-test          default     Pending    kubeflow.org/v1     mpi         2     0/2        10s
```

When no jobs exist:

```text
No jobs found
```

---

## arena job get

Retrieve and display detailed information about a training job, including status and pod details.

### Syntax

```
arena job get <name> [flags]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the training job |

### Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--details` | | bool | `false` | Show job configuration details (from the stored YAML) |

Also inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
# Get job details
$ arena job get pytorch-simple

# Include the original YAML configuration
$ arena job get pytorch-simple --details

# JSON output
$ arena job get pytorch-simple -o json
```

### Output

```text
Name:      pytorch-simple
Namespace: default
Status:    Running
Replicas:  1/1
Age:       5m

Pods:
  NAME                     STATUS    IP            NODE
  pytorch-simple-master-0  Running   10.244.0.15   node-1
```

With `--details`, a `Configuration` section is appended showing the original task YAML:

```text
Configuration:
  name: pytorch-simple
  framework:
    name: pytorch
  image: pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime
  ...
```

---

## arena job logs

Stream logs from a training job pod. By default, streams from the primary pod (master/chief/launcher).

### Syntax

```
arena job logs <name> [flags]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the training job |

### Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--follow` | `-f` | bool | `false` | Follow log output |
| `--tail` | | int | `-1` | Number of lines to show from end (`-1` = all) |
| `--pod` | | string | `""` | Pod name (skip label selector) |
| `--container` | | string | `""` | Container name (default: first container) |

Also inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
# View master pod logs
$ arena job logs my-job

# Follow logs with tail
$ arena job logs my-job -f --tail 100

# View a specific worker pod's logs
$ arena job logs my-job --pod my-job-worker-0

# View a specific container's logs (e.g., TensorBoard sidecar)
$ arena job logs my-job --pod my-job-worker-0 --container tensorboard
```

### Output

Streamed log lines from the selected pod and container. No table output is produced.

---

## arena job suspend

Suspend a running training job by setting `spec.runPolicy.suspend` to `true`.

### Syntax

```
arena job suspend <name>
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the training job |

### Flags

This command has no flags of its own. It inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
$ arena job suspend pytorch-simple
```

### Output

```text
pytorchjob/pytorch-simple suspended
```

---

## arena job resume

Resume a suspended training job by setting `spec.runPolicy.suspend` to `false`.

### Syntax

```
arena job resume <name>
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the training job |

### Flags

This command has no flags of its own. It inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
$ arena job resume pytorch-simple
```

### Output

```text
pytorchjob/pytorch-simple resumed
```

---

## arena job delete

Delete a training job by name or YAML file (similar to `kubectl delete -f`).

### Syntax

```
arena job delete [name] [flags]
arena job delete -f <file> [flags]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | One of `name` or `-f` | Name of the training job |

### Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--file` | `-f` | string | `""` | Path to YAML file (extracts job name from file) |

Also inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
# Delete by name
$ arena job delete pytorch-simple

# Delete by YAML file
$ arena job delete -f examples/v2/quickstart/pytorch-simple.yaml

# Delete in a specific namespace
$ arena job delete pytorch-simple -n kubeflow
```

### Output

```text
pytorchjob/pytorch-simple deleted
```

---

## arena job status

Alias for `arena job get`. Show job status and pod details.

### Syntax

```
arena job status <name> [flags]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the training job |

### Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--details` | | bool | `false` | Show job configuration details |

Also inherits [global flags](#global-flags) and `--output` from `arena job`.

### Examples

```shell
$ arena job status pytorch-simple
```

Output is identical to [`arena job get`](#arena-job-get).

---

## arena submit (legacy)

Submit a training job using CLI flags instead of a YAML file.

> **Note:** `arena job run` with a YAML file is the preferred method for v2. The `submit` command keeps the v1 flag interface so existing scripts work unchanged.

> **Note:** Per-framework subcommands are also available: `arena submit pytorchjob`, `tfjob`, `mpijob`, `horovodjob`, and `deepspeedjob`, plus v1 aliases like `pytorch`, `tf`, and `mpi`. Each subcommand's `--help` shows only that framework's flags; the generic `arena submit <type>` form keeps working unchanged.

### Syntax

```
arena submit <type> [flags] -- [command]
```

Supported types (case-insensitive): `pytorch`, `tensorflow`, `tf`, `mpi`, `horovod`, `deepspeed`.

Trailing arguments after `--` are used as the run command.

### Flags

**Required:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | `""` | Job name (required) |
| `--image` | string | `""` | Container image (required) |

**Worker configuration:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--workers` | int | `1` | Number of worker replicas (PyTorch: N total = 1 master + N-1 workers) |

**Resources:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--gpus` | int | `0` | Number of GPUs per worker |
| `--cpu` | string | `""` | CPU request (e.g. `500m`, `2`) |
| `--memory` | string | `""` | Memory request (e.g. `1Gi`, `512Mi`) |

**Environment and data:**

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--env` | `-e` | stringArray | `nil` | Environment variable (`key=value`, repeatable) |
| `--data` | `-d` | stringArray | `nil` | Data volume (`name:path:pvc`, repeatable); v1 form `<pvc>:<path>` also accepted (see migration guide) |
| `--data-dir` | | stringArray | `nil` | Host path volume (`name:path:hostpath`, repeatable); v1 forms `<host_path>` / `<host_path>:<container_path>` also accepted (see migration guide) |
| `--config-file` | | stringArray | `nil` | ConfigMap volume (`name:path:configmap`, repeatable); the v1 host-file form is rejected with guidance (see migration guide) |
| `--label` | `-l` | stringArray | `nil` | Label (`key=value`, repeatable) |
| `--annotation` | `-a` | stringArray | `nil` | Annotation (`key=value`, repeatable) |

**Scheduling:**

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--selector` | | stringArray | `nil` | Node selector (`key=value`, repeatable) |
| `--toleration` | | stringArray | `nil` | Toleration (`key=value:effect`, repeatable) |
| `--priority` | `-p` | string | `""` | Priority class name (v1 semantics) |
| `--gang` | | bool | `false` | Enable gang scheduling |
| `--scheduler` | | string | `""` | Custom scheduler name |
| `--affinity-policy` | | string | `""` | Affinity policy |
| `--affinity-constraint` | | string | `""` | Affinity constraint |
| `--affinity-target` | | string | `""` | Affinity target (pod or node) |
| `--queue` | | bool | `false` | Suspend the job so an external queue manager (e.g. [kube-queue](https://github.com/kube-queue/kube-queue)) can schedule it |

**Lifecycle:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--clean-task-policy` | string | `Running` | Clean pod policy (`None`, `Running`, `All`); pass an empty value to use the operator default |
| `--running-timeout` | string | `""` | Running timeout (e.g. `2h`, `7d`) |
| `--ttl-after-finished` | string | `""` | TTL after finished (e.g. `7d`) |
| `--job-backoff-limit` | int | `0` | Max restart count for the job |
| `--retry` | int | `0` | Times to retry the job (same as `--job-backoff-limit`) |
| `--success-policy` | string | `""` | Success policy (`ChiefWorker`, `AllWorkers`, TF only) |

**Runtime:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--image-pull-policy` | string | `""` | Image pull policy (`Always`, `IfNotPresent`, `Never`) |
| `--image-pull-secret` | stringArray | `nil` | Image pull secret name (repeatable) |
| `--service-account` | string | `""` | Service account name |
| `--job-restart-policy` | string | `""` | Restart policy (`Always`, `OnFailure`, `Never`) |
| `--hostNetwork` | bool | `false` | Use host network |
| `--hostIPC` | bool | `false` | Use host IPC namespace |
| `--hostPID` | bool | `false` | Use host PID namespace |

**Task:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--working-dir` | string | `""` | Working directory in container |
| `--shell` | string | `""` | Shell to use (default `/bin/sh`) |
| `--share-memory` | string | `2Gi` | Shared memory size (e.g. `8Gi`); pass an empty value to skip the `/dev/shm` volume |
| `--device` | stringArray | `nil` | Extended resource (`name=count`, repeatable) |
| `--gpu-type` | string | `""` | GPU type (sets node selector `nvidia.com/gpu.product`) |

**Logging / TensorBoard:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tensorboard` | bool | `false` | Enable TensorBoard sidecar |
| `--logdir` | string | `/training_logs` | TensorBoard log directory; pass an empty value to opt out of the default |
| `--tensorboard-image` | string | `""` | TensorBoard container image |

**PyTorch-specific:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--nproc-per-node` | string | `""` | Processes per node (`auto`, `gpu`, `cpu`, or integer) |

**TensorFlow-specific:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ps` | int | `0` | Number of parameter servers |
| `--chief` | bool | `false` | Enable Chief worker |
| `--evaluator` | bool | `false` | Enable Evaluator worker |

**MPI-specific:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--slots-per-worker` | int | `0` | Slots per worker |
| `--gputopology` | bool | `false` | Enable GPU topology (sets host networking, `gpu-topology` labels, and the MPI annotation) |
| `--mounts-on-launcher` | bool | `false` | MPI: when false the launcher main container gets no volumeMounts; volumes stay declared in the pod spec so init containers keep their mounts. |

**Code sync (v1 compatibility):**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sync-mode` | string | `""` | Code sync mode: `git`, `rsync`, or `hdfs` |
| `--sync-source` | string | `""` | Sync source URL/path (required when `--sync-mode` is set) |
| `--sync-image` | string | `""` | Image used by the sync init container (defaults per mode: git-sync, rsync, or HDFS image) |

Synced code lands in `<working-dir>/code` (default `/root/code`) as an init container. The code is exchanged through a shared `code-sync` emptyDir volume mounted by both the sync init container and the main container (size limit 10Gi; v1 used an unlimited emptyDir). The `--sync-mode` path always uses this per-pod ephemeral `code-sync` emptyDir and never targets persistent storage, so it is unaffected by the sync write-target validation (see the YAML schema docs).

**Per-role resources (TensorFlow):**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ps-cpu` | string | `""` | CPU for PS pods |
| `--ps-memory` | string | `""` | Memory for PS pods |
| `--ps-gpus` | int | `0` | GPUs for PS pods |
| `--chief-cpu` | string | `""` | CPU for Chief pods |
| `--chief-memory` | string | `""` | Memory for Chief pods |
| `--evaluator-cpu` | string | `""` | CPU for Evaluator pods |
| `--evaluator-memory` | string | `""` | Memory for Evaluator pods |
| `--worker-cpu` | string | `""` | CPU for worker pods (overrides `--cpu`) |
| `--worker-memory` | string | `""` | Memory for worker pods (overrides `--memory`) |

Resource values are applied as both requests and limits (Guaranteed QoS). Role-specific flags win over the generic `--cpu`/`--memory`.

The v1 `-limit` variants of the per-role flags (`--ps-cpu-limit`, `--worker-memory-limit`, etc.) are also accepted. A `-limit` flag overrides only the limit — requests stay at the request flag — so `--worker-cpu 1 --worker-cpu-limit 2` yields requests cpu=1 / limits cpu=2 (Burstable QoS). Without `-limit` flags, requests equal limits (Guaranteed QoS).

**Deprecated v1 flags:**

The remaining v1-only flags are accepted as deprecated compatibility flags so existing scripts do not break; most have no effect — see the table for exceptions. Using them prints a deprecation warning.

| Flag | Behavior |
|------|----------|
| `--config <path>` | works (bound to `--kubeconfig`); deprecated |
| `--loglevel <level>` | mapped to verbosity (`--verbose`): debug=2, info=1, warn/error=0 |
| `--rdma` | ignored; request RDMA hardware with `--device <resource>=<count>` |
| `--{ps,worker,chief,evaluator,launcher}-selector`, `--{worker,launcher}-annotation`, `--{ps,worker}-image`, `--{ps,worker,chief,ssh}-port`, `--{ps,worker}-affinity-{policy,constraint}` | accepted with a warning, ignored (needs per-role RoleConfig extension) |
| `--pprof`, `--trace`, `--helm-binary`, `--arena-namespace`, `--model-name`, `--model-source`, `--starting-timeout`, `--role-sequence`, `--ssh-secret` | accepted with a warning, no effect |

**Dry-run:**

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--dry-run` | | bool | `false` | Print CRD as JSON or YAML without submitting |
| `--output` | `-o` | string | `json` (when `--dry-run`) | Dry-run output format: `json` or `yaml` |

Also inherits [global flags](#global-flags).

### Examples

```shell
# Submit a PyTorch job with 2 workers and 1 GPU each
$ arena submit pytorch --name my-job --image pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime \
    --workers 2 --gpus 1 -- python -c "import torch; print('CUDA:', torch.cuda.is_available())"

# Submit a TensorFlow job with chief and parameter server
$ arena submit tensorflow --name tf-job --image tensorflow/tensorflow:2.16.1-gpu \
    --workers 3 --gpus 1 --chief --ps 1 -- python -c "import tensorflow as tf; print('TF:', tf.__version__)"

# Submit an MPI job with deepspeed
$ arena submit deepspeed --name ds-job --image deepspeed/deepspeed \
    --workers 4 --gpus 8 --slots-per-worker 8 -- python -c "print('DeepSpeed ready')"

# Dry-run to inspect the generated CRD
$ arena submit pytorch --name my-job --image pytorch/pytorch --dry-run

# Git code sync with a priority class and per-worker resources
$ arena submit pytorch --name sync-job --image pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime \
    --workers 3 --gpus 1 --cpu 4 --memory 16Gi -p high \
    --sync-mode git --sync-source https://github.com/example/repo.git \
    -- python train.py --epochs 10
```

### Output

```text
Job my-job submitted successfully
```

---

## arena check

Verify that the required Kubeflow training operator CRDs (PyTorchJob, TFJob, MPIJob) are installed and accessible in the cluster.

### Syntax

```
arena check
```

### Flags

This command has no flags of its own. It inherits [global flags](#global-flags).

### Examples

```shell
$ arena check
```

### Output

When all CRDs are installed and compatible:

```text
✓ PyTorchJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ TFJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ MPIJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
  compatible: ✓ (storage version v1 supported by arena)
```

When a CRD is missing:

```text
✗ MPIJob: not installed
```

When the MPIJob storage version is incompatible:

```text
✓ MPIJob: installed (expected: kubeflow.org/v2)
  versions: v2 (served, storage)
  compatible: ✗ (storage version v2, arena supports: v1, v2beta1)
```

The command exits with a non-zero status if any CRD is missing or incompatible.

---

## arena version

Print Arena version information.

### Syntax

```
arena version
```

### Flags

This command has no flags of its own. It inherits [global flags](#global-flags).

### Examples

```shell
$ arena version
```

### Output

```text
Arena v2
  Version:    dev
  Git Commit: unknown
  Build Date: unknown
```

Values are injected at build time. In release builds, `Version`, `Git Commit`, and `Build Date` reflect the actual build.

---

## arena completion

Generate shell completion scripts for Arena.

### Syntax

```
arena completion <shell>
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `shell` | Yes | One of: `bash`, `zsh`, `fish`, `powershell` |

### Flags

This command has no flags of its own. It inherits [global flags](#global-flags).

### Examples

```shell
# Install bash completion
$ arena completion bash > /etc/bash_completion.d/arena

# Install zsh completion
$ arena completion zsh > "${fpath[1]}/_arena"

# Install fish completion
$ arena completion fish > ~/.config/fish/completions/arena.fish

# Generate PowerShell completion
$ arena completion powershell > arena.ps1
```

### Output

The shell completion script is printed to stdout. Redirect to the appropriate completion file for your shell.

---

## arena top

Display resource (GPU) usage. This is a parent command; use `arena top job` to view job-level GPU usage.

### Subcommands

| Command | Description |
|---------|-------------|
| [`job`](#arena-top-job) | Display resource (GPU) usage of jobs |

---

## arena top job

Display GPU resource usage of training jobs across PyTorchJob, TFJob, and MPIJob CRD kinds.

### Syntax

```
arena top job [flags]
```

### Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--output` | `-o` | string | `table` | Output format: `table`, `wide`, `json`, `yaml` |

Also inherits [global flags](#global-flags).

### Examples

```shell
# Show GPU usage of all jobs
$ arena top job

# Wide output with additional columns
$ arena top job -o wide

# JSON output
$ arena top job -o json
```

### Output

Default table format:

```text
NAME              STATUS     GPU_REQUESTED   REPLICAS   AGE
pytorch-simple    Running    1               1/1        5m
mpi-test          Pending    2               0/2        10s
```

Wide format (`-o wide`) adds `NAMESPACE`, `APIVERSION`, and `FRAMEWORK` columns:

```text
NAME              NAMESPACE   STATUS     APIVERSION            FRAMEWORK   GPU_REQUESTED   REPLICAS   AGE
pytorch-simple    default     Running    kubeflow.org/v1       pytorch     1               1/1        5m
mpi-test          default     Pending    kubeflow.org/v1       mpi         2               0/2        10s
```

When no jobs exist:

```text
No jobs found
```

---

## Job Status Values

The `STATUS` column in `list`, `get`, and `top job` output reflects the CRD's condition state. Common values include:

| Status | Description |
|--------|-------------|
| `Pending` | Job has been created but no conditions are set yet |
| `Created` | Job has been acknowledged by the training operator |
| `Running` | Job pods are running |
| `Restarting` | Job is restarting after a failure |
| `Succeeded` | Job completed successfully |
| `Failed` | Job has failed |
| `Suspended` | Job has been suspended via `arena job suspend` |
| `Unknown` | Conditions exist but none are True |

---

## Output Formats

The `-o/--output` flag controls the output format for `arena job list`, `arena job get`, and `arena top job`.

| Format | Description |
|--------|-------------|
| `table` | Aligned text table (default) |
| `wide` | Wider table with additional columns |
| `json` | JSON with 2-space indentation |
| `yaml` | YAML output |

For `json` and `yaml` formats, the full structured data (including pod details and configuration) is serialized. For `table` and `wide`, a human-readable table is rendered with dynamic column widths.

---

## Namespace Resolution

Arena resolves the target Kubernetes namespace using a 4-level priority chain:

1. `-n/--namespace` CLI flag (highest priority)
2. `namespace` field in the YAML task file
3. Namespace from the kubeconfig context
4. `default` (lowest priority)

All commands that interact with the cluster use this resolution. For `arena job run`, the YAML namespace is overridable with `-n`. For other commands (`list`, `get`, `logs`, `suspend`, `resume`, `delete`), the namespace comes from the flag or kubeconfig context.
