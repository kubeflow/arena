# Monitoring & Status

How to observe and troubleshoot your Arena v2 training jobs.

## Job Status Lifecycle

Arena derives job status from the Kubeflow training-operator CRD `status.conditions` array. Conditions are cumulative (a completed job has both `Created=True` and `Succeeded=True`) and appended chronologically, so Arena scans the list in reverse and returns the **last condition with `status: "True"`** as the current state.

| Status | Meaning |
|--------|---------|
| `Pending` | Job object created, but the operator has not set any conditions yet |
| `Created` | Job accepted by the operator |
| `Running` | Pods are executing |
| `Succeeded` | Job completed successfully |
| `Failed` | Job failed |
| `Suspended` | Job suspended via `spec.runPolicy.suspend` (set by `arena job suspend`) |
| `Unknown` | Conditions exist but none are `True` |

The fallback chain in `extractJobPhase` is: last `True` condition type, then `Suspended` (if `spec.runPolicy.suspend == true`), then `Pending` (no conditions at all), then `Unknown`.

## Is My Job Running?

List all jobs in the current namespace:

```shell
$ arena job list
```

```text
NAME              STATUS    REPLICAS  AGE
pytorch-example   Running   4/4       5m
tf-mnist          Failed    0/2       12m
mpi-resnet        Running   3/3       1h
```

The **REPLICAS** column shows `ready/total`, where `ready` is the sum of active and succeeded replicas across all roles and `total` is the desired replica count from the spec.

For more columns, use the wide format:

```shell
$ arena job list -o wide
```

```text
NAME              NAMESPACE   STATUS    APIVERSION       FRAMEWORK    GPU  REPLICAS  AGE
pytorch-example   default     Running   kubeflow.org/v1  pytorch      4    4/4       5m
mpi-resnet        default     Running   kubeflow.org/v1  mpi          2    3/3       1h
```

Wide adds `NAMESPACE`, `APIVERSION`, `FRAMEWORK`, and `GPU` columns. For scripting and automation, use machine-readable output:

```shell
$ arena job list -o json
$ arena job list -o yaml
```

See [cli-reference.md](cli-reference.md) for the full set of flags.

## What's Happening with My Job?

Get detailed information about a single job, including a pod table:

```shell
$ arena job get pytorch-example
```

```text
Name:      pytorch-example
Namespace: default
Status:    Running
Replicas:  4/4
Age:       5m

Pods:
  NAME                       STATUS     IP            NODE
  pytorch-example-master-0   Running    10.244.1.12   node-gpu-01
  pytorch-example-worker-0   Running    10.244.2.8    node-gpu-02
  pytorch-example-worker-1   Running    10.244.2.9    node-gpu-02
  pytorch-example-worker-2   Running    10.244.1.13   node-gpu-01
```

The pod **STATUS** column mirrors `kubectl`'s logic: it shows container waiting reasons (e.g. `ImagePullBackOff`, `CrashLoopBackOff`, `ContainerCreating`) rather than just the coarse pod phase when a container is not yet running. Terminal phases (`Succeeded`, `Failed`) are returned as-is. This makes failures immediately visible without a separate `kubectl describe`.

To see the exact YAML configuration that was submitted at job creation time:

```shell
$ arena job get pytorch-example --details
```

The `--details` flag reads the job's ConfigMap anchor (key `arena-v2.yaml`) and renders the original `Task` spec under a `Configuration:` section. This requires the ConfigMap to exist, which is created automatically by `arena job run`.

`arena job status <name>` is an alias for `arena job get <name>` and accepts the same `--details` flag.

## Viewing Logs

View logs from the primary pod of a job:

```shell
$ arena job logs pytorch-example
```

Arena selects a default log pod per framework using the provider's label selector:

| Framework | Default log pod |
|-----------|-----------------|
| PyTorch | `master` replica |
| TensorFlow | `chief` replica (falls back to `worker-0` when no chief exists) |
| MPI / Horovod / DeepSpeed | `launcher` replica |

Follow the log stream in real time:

```shell
$ arena job logs pytorch-example --follow
```

Show only the last N lines:

```shell
$ arena job logs pytorch-example --tail 100
```

Target a specific pod or container:

```shell
$ arena job logs pytorch-example --pod pytorch-example-worker-0
$ arena job logs pytorch-example --pod pytorch-example-worker-0 --container tensorboard
```

When `--container` is omitted, the first container in the pod is used. If you specify a container that does not exist, Arena lists the available container names in the error message.

> **Note:** Log streaming targets a single pod at a time. You cannot tail multiple pods simultaneously with one command.

## Why Did It Fail?

The pod STATUS column in `arena job get` is your first diagnostic. Common patterns and their causes:

| Pod STATUS | Likely cause | What to do |
|------------|--------------|------------|
| `ImagePullBackOff` | Wrong image name, or registry requires authentication | Verify the image tag; add `image_pull_secrets` to your YAML |
| `CrashLoopBackOff` | Application exits immediately (code error, bad args, missing data) | Run `arena job logs <name>` to read the container's stderr |
| `ContainerCreating` | Pulling image or mounting volumes (transient) | Wait; if it persists, check PVC availability and storage classes |
| `OOMKilled` | Container exceeded its memory limit | Increase `memory` in the role's `resources` block |
| `Pending` | Scheduler cannot place the pod (no GPUs, taints, quota) | Check node GPU availability and tolerations |
| `Failed` | Pod reached a terminal failure state | Inspect logs and conditions for details |

Example of a failed job where the image cannot be pulled:

```shell
$ arena job get bad-job
```

```text
Name:      bad-job
Namespace: default
Status:    Failed
Replicas:  0/2
Age:       3m

Pods:
  NAME               STATUS              IP   NODE
  bad-job-master-0   ImagePullBackOff         node-gpu-01
  bad-job-worker-0   ImagePullBackOff         node-gpu-02
```

For application-level errors, stream the logs to see the traceback:

```shell
$ arena job logs bad-job --tail 50
```

See [best-practices.md](best-practices.md) for guidance on preventing common failures.

## GPU Usage

View GPU requests across all jobs:

```shell
$ arena top job
```

```text
NAME              STATUS    GPU_REQUESTED  REPLICAS  AGE
pytorch-example   Running   4              4/4       5m
mpi-resnet        Running   2              3/3       1h
```

The **GPU_REQUESTED** column is the total GPUs requested across all roles, calculated as `replicas × per-pod GPU` (from `resources.requests["nvidia.com/gpu"]` on the first container of each role). For example, a job with 2 workers each requesting 2 GPUs shows `4`.

Use the wide format for namespace, framework, and API version detail:

```shell
$ arena top job -o wide
```

```text
NAME              NAMESPACE   STATUS    APIVERSION       FRAMEWORK    GPU_REQUESTED  REPLICAS  AGE
pytorch-example   default     Running   kubeflow.org/v1  pytorch      4              4/4       5m
```

> **Note:** `arena top job` shows **requested** GPUs from the CRD spec, not real-time utilization. It does not query running GPU metrics.

## Viewing Job Configuration

`arena job get --details` reads the ConfigMap anchor that `arena job run` creates alongside the CRD. It renders the exact `Task` YAML you submitted, including all fields, defaults, and overrides applied via `--set`.

```shell
$ arena job get pytorch-example --details
```

```text
Name:      pytorch-example
Namespace: default
Status:    Running
Replicas:  4/4
Age:       5m

Pods:
  NAME                       STATUS    IP            NODE
  pytorch-example-master-0   Running   10.244.1.12   node-gpu-01
  ...

Configuration:
  name: pytorch-example
  framework:
    name: pytorch
  image: pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime
  run: python train.py --epochs 10
  worker:
    replicas: 3
    resources:
      nvidia.com/gpu: 1
  ...
```

If the ConfigMap is missing (for example, the job was created outside Arena), the `Configuration:` section is silently omitted and only the status and pod table are shown.

## Verifying Cluster Prerequisites

Before submitting jobs, confirm the required CRDs are installed:

```shell
$ arena check
```

```text
✓ PyTorchJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ TFJob: installed (expected: kubeflow.org/v1)
  versions: v1 (served, storage)
✓ MPIJob: installed (expected: kubeflow.org/v2beta1)
  versions: v2beta1 (served, storage)
  compatible: ✓ (storage version v2beta1 supported by arena)
```

A missing or incompatible CRD is marked with `✗` and the command exits with an error. For MPIJob, Arena additionally checks that the CRD's storage version is one it supports (`v2beta1` or `v1`).

## Suspending and Resuming

Suspend a running job by setting `spec.runPolicy.suspend` to `true`:

```shell
$ arena job suspend pytorch-example
pytorchjob/pytorch-example suspended
```

Resume it later:

```shell
$ arena job resume pytorch-example
pytorchjob/pytorch-example resumed
```

While suspended, the job status shows `Suspended` and its pods are removed by the operator. Resuming recreates them.

## Limitations

Arena v2 is focused on job submission and status. Be aware of these gaps when troubleshooting:

- **No `arena top node`** — v1 had node-level GPU reporting; this is not yet ported to v2.
- **No real-time GPU utilization** — `arena top job` shows requested GPUs only, not live usage. Use `kubectl` or a metrics server for that.
- **No Kubernetes events display** — Arena does not surface `kubernetes.io` events. Use `kubectl describe pod` for scheduling and volume errors.
- **No job history** — there is no start/completion timestamp, duration, or exit-code tracking beyond what the CRD conditions expose.
- **Single-pod log streaming** — `arena job logs` targets one pod at a time; it cannot tail multiple pods simultaneously.
- **TensorBoard access requires manual port-forward** — Arena creates the TensorBoard Deployment and Service, but you must run `kubectl port-forward` yourself to reach the UI.
- **No Prometheus/Grafana integration** — there is no built-in metrics pipeline. Wire your own monitoring stack if you need dashboards and alerts.

For command-level details, see [cli-reference.md](cli-reference.md). For recommendations on avoiding common pitfalls, see [best-practices.md](best-practices.md).
