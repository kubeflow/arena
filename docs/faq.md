# FAQ

## Installation

### Q: `arena check` shows "✗ PyTorchJob: not installed"
Cause: The Kubeflow Training Operator CRD for PyTorchJob is not installed on your cluster.
Solution: Install the Kubeflow Training Operator. See [installation.md](installation.md) for instructions.

### Q: `arena check` shows "✗ TFJob: not installed" or "✗ MPIJob: not installed"
Cause: The corresponding Training Operator CRD (TFJob or MPIJob) is missing from the cluster.
Solution: Install the Kubeflow Training Operator, which includes all three CRDs. If only MPIJob is missing, ensure your operator version includes MPIJob support. See [installation.md](installation.md).

### Q: `arena check` shows "compatible: ✗" for MPIJob
Cause: The MPIJob CRD storage version on your cluster does not match what Arena supports. Arena supports `v1` and `v2beta1`.
Solution: Check your MPIJob CRD storage version with `kubectl get crd mpijobs.kubeflow.org -o jsonpath='{.spec.versions[?(@.storage)].name}'`. Install a compatible MPI Operator version that serves `v1` or `v2beta1` as the storage version. See [installation.md](installation.md).

### Q: `arena check` shows "? PyTorchJob: version not resolved"
Cause: Arena could not resolve the API version for the CRD kind. This may indicate a partially installed or corrupted CRD.
Solution: Reinstall the Kubeflow Training Operator and run `arena check` again.

### Q: Build fails with a Go version error
Cause: Arena v2 requires Go >= 1.26.5 (per `go.mod`).
Solution: Upgrade your Go toolchain to 1.26.5 or later. Run `go version` to check your current version.

## Job Submission

### Q: `job "myjob" already exists (type: PyTorchJob)`
Cause: A job with the same name already exists in the target namespace.
Solution: Delete the existing job first with `arena job delete myjob`, or submit with a different `--name`.

### Q: `job was created by arena v1`
Cause: The job was created by Arena v1 (Helm-based) and carries no Arena v2 framework label. The v2 CLI cannot manage v1 jobs.
Solution: Use the v1 CLI to manage the job, or delete it and resubmit with `arena-v2`. See [best-practices.md](best-practices.md) for migration guidance.

### Q: `job "myjob" not found in namespace "default"`
Cause: No job with the given name exists in the specified namespace. The CLI checked both the ConfigMap anchor (v2 fast path) and direct CRD lookups.
Solution: Verify the job name and namespace. Use `arena job list` to see all jobs in the current namespace.

### Q: `ray provider is not yet implemented`
Cause: The Ray framework is listed as a valid framework type but its provider is a placeholder and not yet implemented in the Alpha release.
Solution: Use `pytorch`, `tensorflow`, `mpi`, `horovod`, or `deepspeed` frameworks instead.

### Q: `unsupported framework type: "foo" (must be pytorch, tensorflow, mpi, horovod, deepspeed, or ray)`
Cause: The framework name passed to `arena submit` is not recognized. Framework names are case-insensitive but must match one of the supported types.
Solution: Use a supported framework name. For example: `arena submit pytorch --name myjob --image pytorch:2.1`.

### Q: `unsupported framework: "foo" (must be pytorch, tensorflow, mpi, horovod, deepspeed, or ray)`
Cause: The `framework.name` field in your YAML file does not match any supported framework.
Solution: Set `framework.name` to one of: `pytorch`, `tensorflow`, `mpi`, `horovod`, `deepspeed`, or `ray` (note: `ray` is not yet implemented).

### Q: `--file is required`
Cause: You ran `arena job run` without the `--file` (`-f`) flag.
Solution: Provide a YAML file path: `arena job run -f examples/v2/pytorch-train.yaml`. Use `--dry-run` to preview without submitting.

### Q: `name is required`
Cause: The `name` field is missing from your YAML configuration (or the `--name` flag was not provided for `submit`).
Solution: Add a `name` field to your YAML, or pass `--name <job-name>` on the command line. Names must be valid DNS labels (lowercase alphanumeric and hyphens, max 63 chars).

### Q: `invalid name "My Job": ...`
Cause: The job name does not meet Kubernetes DNS label requirements (RFC 1123). Names must contain only lowercase alphanumeric characters or hyphens, start with an alphanumeric character, and be at most 63 characters long.
Solution: Use a valid DNS label name, e.g. `my-job` or `training-run-01`.

### Q: `image is required`
Cause: No container image was specified in the YAML `image` field or via `--image`.
Solution: Add the `image` field to your YAML (e.g. `image: pytorch/pytorch:2.1.0-cuda11.8-cudnn8-runtime`) or pass `--image <image>`.

### Q: `run is required`
Cause: No run command was specified. The `run` field tells the container what to execute.
Solution: Add a `run` field to your YAML (e.g. `run: python train.py`), or pass the command after `--` in `arena submit`: `arena submit pytorch --name myjob --image pytorch:2.1 -- python train.py`.

### Q: How do I use `--set` with resource names containing dots (e.g. `nvidia.com/gpu`)?
Use single quotes around the dotted segment:
```bash
arena job run -f job.yaml --set worker.resources.'nvidia.com/gpu'=2
```
Single-quoted segments are treated as literal keys and not split on dots. See [yaml-schema.md](yaml-schema.md) for the full `--set` syntax.

### Q: `failed to parse --set "key": no '=' found in expression`
Cause: The `--set` expression is missing the `=` separator between key and value.
Solution: Use the format `--set key=value`, e.g. `--set worker.replicas=4`.

### Q: `failed to parse --set "'foo": mismatched single quote in expression`
Cause: A single-quoted segment in the `--set` expression is missing its closing quote.
Solution: Ensure every opening single quote has a matching closing quote, e.g. `--set worker.resources.'nvidia.com/gpu'=2`.

### Q: What should the `version` field in YAML be?
The current schema version is `0.1.0` (format: `MAJOR.MINOR.PATCH`). If omitted, it defaults to `0.1.0` automatically. Using a version newer than `0.1.x` produces: `version "0.2.0" is newer than supported (current: 0.1.x)`.

### Q: `pytorch requires worker or master (at least one must be specified)`
Cause: For PyTorch jobs, you must define either a `worker` block or a `master` block (or both). Neither was found.
Solution: Add at least one role. For single-node training, use `master` only. For multi-node, use `worker` with `replicas >= 1`.

### Q: `worker.replicas must be > 0, got 0`
Cause: The `worker.replicas` field is set to zero or a negative number.
Solution: Set `worker.replicas` to at least 1.

### Q: `master role is only valid for pytorch framework`
Cause: The `master` role block was specified for a non-PyTorch framework. The `master` role is only valid for PyTorch.
Solution: Remove the `master` block or switch to the `pytorch` framework. Similarly, `chief`, `ps`, and `evaluator` are only valid for `tensorflow`, and `launcher` is only valid for `mpi`, `horovod`, and `deepspeed`.

### Q: `launcher role is constrained to replicas=1, got 3`
Cause: Constrained roles (`master`, `chief`, `launcher`, `evaluator`) can only have 1 replica. You specified more.
Solution: Set replicas to 1 (or omit the `replicas` field, which defaults to 1 for constrained roles).

### Q: `success_policy is only valid for tensorflow framework`
Cause: The `success_policy` lifecycle field was set for a non-TensorFlow framework.
Solution: Remove `success_policy` from the lifecycle section, or switch to the `tensorflow` framework. Valid values are `ChiefWorker` (alias for the default) and `AllWorkers`.

### Q: `invalid clean_pod_policy: "Foo" (must be None, Running, or All)`
Cause: The `clean_pod_policy` value is not one of the allowed values.
Solution: Use `None`, `Running`, or `All`.

### Q: `invalid mpi_implementation: "Foo" (must be OpenMPI, Intel, or MPICH)`
Cause: The `mpi_implementation` framework option for MPI jobs has an invalid value.
Solution: Set `mpi_implementation` to `OpenMPI`, `Intel`, or `MPICH`.

### Q: Warning: `env var looks like a secret but is stored in plaintext in ConfigMap`
Cause: An environment variable name matches a secret-like pattern (contains `token`, `key`, `secret`, `password`, or `credential`) but the value is stored as a plaintext string in the ConfigMap.
Solution: Use a `secretKeyRef` instead. In YAML:
```yaml
envs:
  HF_TOKEN:
    secret: my-hf-creds
    key: token
```
See [best-practices.md](best-practices.md) for secrets management.

### Q: Warning: `creating resources in system namespace — ensure this is intentional`
Cause: You are creating resources in a Kubernetes system namespace (`kube-system`, `kube-public`, or `kube-node-lease`).
Solution: Use a dedicated namespace for training jobs. Specify it with `-n <namespace>` or the `namespace` field in YAML.

## Job Status

### Q: REPLICAS shows "0/4" — what does this mean?
0 pods are ready out of 4 requested. The job is still starting, or pods are failing to schedule or launch. Check pod status with `arena job get <name>` or `kubectl get pods`.

### Q: Pod status shows "ImagePullBackOff"
Cause: The container image cannot be pulled from the registry.
Solution: Verify the image name and tag. Ensure image pull secrets are configured (`image_pull_secrets` in YAML or `--image-pull-secret` on CLI). Check registry access from your cluster nodes. See [monitoring.md](monitoring.md).

### Q: Pod status shows "CrashLoopBackOff"
Cause: The container starts but crashes repeatedly, usually due to an application error or missing dependency.
Solution: Check logs with `arena job logs <name>` to identify the application error. Common causes: missing Python packages, incorrect entrypoint, or insufficient memory.

### Q: Pods stuck in "Pending"
Cause: The Kubernetes scheduler cannot find suitable nodes for the pods. This is often due to GPU shortage, node selector mismatch, or insufficient resources.
Solution: Run `kubectl describe pod <pod-name>` and check the Events section for scheduling failures. Verify GPU availability with `kubectl get nodes -o custom-columns=NAME:.metadata.name,GPUS:.status.capacity['nvidia\.com/gpu']`.

### Q: Job status shows "Suspended"
Cause: The job has the `suspend` lifecycle field set to `true`. Suspended jobs do not launch any pods.
Solution: Set `suspend: false` (or remove the field) and resubmit the job.

### Q: `job "myjob" is an MPIJob but MPIJob CRD is not installed`
Cause: You are trying to access an MPIJob that exists in the cluster, but the MPIJob CRD is not installed or not accessible.
Solution: Install the Kubeflow Training Operator with MPIJob support. See [installation.md](installation.md).

## Storage and Volumes

### Q: `storage "data": must specify exactly one of pvc, shm, tmp, hostpath, configmap, or secret`
Cause: A storage entry does not declare any storage type, or the type is not recognized.
Solution: Specify exactly one storage type. Example:
```yaml
storages:
  - name: data
    mount_path: /data
    pvc: my-data-pvc
```

### Q: `storage "data": cannot specify multiple storage types (pvc, hostpath)`
Cause: A single storage entry has more than one storage type declared (e.g. both `pvc` and `hostpath`).
Solution: Use separate storage entries for different volume types. Each entry must declare exactly one type.

### Q: `storages: duplicate storage name "data"`
Cause: Two or more storage entries share the same `name`. Storage names must be unique within a job.
Solution: Give each storage entry a unique name.

### Q: `storage "data": key can only be used with configmap or secret`
Cause: The `key` field was set on a storage entry that is not a ConfigMap or Secret volume.
Solution: The `key` field is only valid for `configmap` and `secret` storage types. Remove it for other types.

### Q: `sync[0].mounts[0].name "data" not found in storages`
Cause: A sync entry references a storage name that does not exist in the `storages` section.
Solution: Ensure every `sync[].mounts[].name` matches a `storages[].name` entry. See [yaml-schema.md](yaml-schema.md) for sync configuration.

## Features

### Q: Does Arena v2 support model registry?
No. Model registry (`--model-name`, `--model-source`) is not planned for v2.

### Q: Does Arena v2 support RDMA?
No. `--rdma` is not planned for v2.

### Q: Does Arena v2 support elastic training (ETJob)?
No. ETJob is not planned for v2.

### Q: I configured scheduling fields but they don't work
Scheduling fields (`node_selector`, `tolerations`, `priority`, `gang`, `affinity`) exist in the YAML schema but are NOT YET IMPLEMENTED in the Alpha release. They will be available in Beta. The fields are validated but not applied to the CRD.

### Q: Does TensorBoard have authentication?
No. TensorBoard UI is accessible to any pod with network access to the job's namespace. Arena logs a warning at submission time: `TensorBoard has no built-in authentication; the UI will be accessible to any pod with network access to this namespace`. Do not expose the TensorBoard Service externally without additional protection (e.g. an authenticating reverse proxy). See [best-practices.md](best-practices.md).

### Q: Which MPIJob API versions does Arena v2 support?
Arena supports `v1` and `v2beta1`. The version is auto-detected from the cluster's MPIJob CRD storage version at submission time. If the cluster's storage version is unsupported, you will see: `cluster MPIJob storage version is v2, arena supports: v1, v2beta1`.

### Q: Can I manage Arena v1 jobs with the v2 CLI?
No. The v2 CLI detects v1 jobs (which lack the Arena v2 framework label) and returns: `job was created by arena v1`. Use the v1 `arena` binary to manage v1 jobs, or delete and resubmit them with `arena-v2`.
