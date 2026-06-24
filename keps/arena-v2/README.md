# Arena v2: AI Workload CLI Redesign

## Table of Contents

- [Arena v2: AI Workload CLI Redesign](#arena-v2-ai-workload-cli-redesign)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
    - [Architecture Overview](#architecture-overview)
    - [User Stories](#user-stories)
    - [Notes/Constraints/Caveats](#notesconstraintscaveats)
  - [Design Details](#design-details)
    - [Core Design Principles](#core-design-principles)
      - [1. Simpler maintenance](#1-simpler-maintenance)
      - [2. Faster iteration](#2-faster-iteration)
      - [3. Better user experience](#3-better-user-experience)
      - [4. Fewer failure modes](#4-fewer-failure-modes)
      - [5. Reduced attack surface](#5-reduced-attack-surface)
      - [6. Resource lifecycle](#6-resource-lifecycle)
    - [Detailed Component Comparison](#detailed-component-comparison)
    - [Migration Notes](#migration-notes)
      - [PyTorch: `worker.replicas` (v2) vs `--workers` (v1)](#pytorch-workerreplicas-v2-vs---workers-v1)
    - [Test Plan](#test-plan)
    - [Graduation Criteria](#graduation-criteria)
    - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Version Skew Strategy](#version-skew-strategy)
  - [Production Readiness Review Questionnaire](#production-readiness-review-questionnaire)
    - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
    - [Rollout, Upgrade and Rollback Planning](#rollout-upgrade-and-rollback-planning)
    - [Monitoring Requirements](#monitoring-requirements)
    - [Dependencies](#dependencies)
    - [Scalability](#scalability)
    - [Troubleshooting](#troubleshooting)
  - [References](#references)

## Summary

Arena v2 is a complete architectural redesign of Arena, addressing fundamental limitations in v1 that make the tool difficult to maintain, extend, and use at scale. The redesign eliminates Helm dependencies, reduces code complexity, introduces YAML-first configuration, and replaces shell-outs with direct Kubernetes API calls.

## Motivation

Arena v1, while functional, suffers from architectural decisions that create significant maintenance burden, slow iteration cycles, and poor user experience at scale.

### Goals

Arena v2 addresses fundamental architectural limitations in Arena v1:

1. **Eliminate Helm dependency** — Generate K8s resources directly in Go, not through Helm chart rendering. This removes 23 Helm charts and the entire Helm Go SDK dependency tree.

2. **YAML-first configuration** — v1 is flag-only (40+ CLI flags for a TFJob) and saves **final** Helm chart values into a ConfigMap. v2 introduces a structured YAML schema where users can version, review, and reuse job configurations as code. CLI flags remain available for quick submissions.

3. **Reduce code complexity** — v1's training submission path spans thousands of lines of Go code across `pkg/commands/`, `pkg/training/`, `pkg/argsbuilder/`, `pkg/apis/`, and `pkg/workflow/`.

4. **Direct K8s API usage** — v1 submits jobs through a 6-step pipeline: serialize args → render Helm chart → save to temp files → create ConfigMap → shell out to `kubectl apply` → patch owner references. v2 adds owner references to created resources and calls the K8s dynamic client API directly: `kc.CreateUnstructured(ctx, gvr, crd)`. No temp files, no shell-outs.

5. **Type-agnostic CRD generation** — Use `unstructured.Unstructured` instead of operator-specific Go types. This decouples Arena from specific Training Operator API versions and makes the code resilient to operator upgrades.

### Non-Goals

- **Breaking existing v1 workflows** — v2 provides legacy compatibility through the `arena submit` command that maps v1 flags to v2's internal model.
- **Supporting all v1 features immediately** — v2 implements features incrementally across three phases, prioritizing core training workloads first.
- **Replacing all Arena binaries immediately** — v2 coexists with v1 during the transition period, using separate binary names (`arena-v2` during development, `arena` as the final build target).

## Proposal

### Architecture Overview

Arena v2 adopts a modular architecture that separates concerns and reduces coupling:

```
cmd/arena-v2/main.go          → Entry point
pkg/cli/                       → Cobra commands (root, status, top, check, job, submit, version, completion)
pkg/task/                      → Task data model, YAML loader, flag overrides
pkg/provider/                  → Provider interface + implementations (PyTorchJob, MPIJob)
pkg/client/                    → K8s dynamic + typed client wrapper
pkg/output/                    → Table formatting for CLI output
```

**Provider Interface**: Each framework (PyTorchJob, MPIJob) implements `provider.Provider`:

- `BuildCRD(task) → *unstructured.Unstructured` — Task spec to Operator CRD
- `GetJobStatus()` / `GetPods()` / `GetLogPod()` — Read back status
- `DefaultValues(task)` — Framework-specific defaults

Adding a new operator = implementing this interface. No Helm charts needed.

### User Stories

**Story 1: Data scientist submitting a distributed training job**

As a data scientist, I want to define my training job configuration in a YAML file so that I can version it in Git, review changes with my team, and reuse the same configuration across multiple experiments without retyping 40+ CLI flags.

**Story 2: Platform engineer adding a new framework**

As a platform engineer, I want to add support for a new training framework (e.g., DeepSpeed) by implementing a single Go file rather than creating a Helm chart, argsbuilder, command file, and type definitions across 5+ files.

**Story 3: Operator maintaining backward compatibility**

As a cluster operator maintaining existing CI/CD pipelines, I want to continue using `arena submit pytorchjob --name my-job --workers 4 --gpus 2 "python train.py"` without rewriting my automation scripts, while the platform team gradually adopts v2's YAML-based workflows.

### Notes/Constraints/Caveats

- v2 is a separate project that lives in the `v2` branch and is built from scratch. It does not import any v1 packages.
- The `arena-v2` binary name is temporary for development. The final build target is `arena` — the `-v2` suffix is removed once v2 reaches feature parity.
- v2's legacy `arena submit` command provides v1-style flag-based submission by mapping old flags to v2's internal `Task` model. Users can continue using v1 syntax without rewriting scripts.

## Design Details

### Core Design Principles

#### 1. Simpler maintenance

The tfjob Helm chart template in v1 is 1,323 lines of nested conditionals for PS/Worker/Chief/Evaluator roles. Every field must be duplicated across different blocks (node selectors, tolerations, volumes, init containers, security contexts). In v2, the same logic is a few hundred lines of Go that programmatically constructs corresponding CR field maps.

#### 2. Faster iteration

Adding a new field in v1 requires changes in five places: types struct, argsbuilder, Helm chart `values.yaml`, Helm chart template, and command file. In v2, adding a field requires updating the YAML schema, the Go struct definition, and the provider's relevant methods.

#### 3. Better user experience

v1's flag-only interface forces users to retype 20-50 flags for every submission. There is no way to save, review, or version job configurations. v2's YAML schema lets users define jobs once in a file, commit to Git, and share with teammates. CLI flags remain available for quick one-off submissions.

#### 4. Fewer failure modes

v1's 6-step submission pipeline has 5 points where temp files or ConfigMaps can leak if the process crashes mid-submission. v2's direct API call is atomic — either the CRD is created or it isn't.

#### 5. Reduced attack surface

v1 shells out to `helm` and `kubectl`, which requires these binaries in `PATH` and exposes command-line injection risks if user input is not properly escaped. v2 uses Go libraries only — no shell execution.

#### 6. Resource lifecycle

A single `arena job run` submission creates multiple K8s resources: the Operator CRD (PyTorchJob/MPIJob), optionally a TensorBoard Deployment+Service, and a ConfigMap storing the original YAML and generated manifest. In v1, these resources are tracked via a resource list in the ConfigMap and deleted through a 3-step pipeline (read ConfigMap → delete resources → delete ConfigMap), which risks resource leaks if the list is incomplete or the process crashes mid-delete.

v2 adopts a different pattern: deletion is simplified to a single operation. Deleting the top-level training job CR triggers K8s garbage collection to cascade-delete all owned resources. This eliminates resource leaks and simplifies `arena job delete` to a single API call. Likewise, all owned resources are cleaned up alongside the training job CR when its `ttlSecondsAfterFinished` expires.

```
PyTorchJob CRD
  ├── owns → ConfigMap "my-job" (stores effective training job YAML; a new CRD could replace the native ConfigMap)
  ├── owns → TensorBoard Deployment
  └── owns → TensorBoard Service
```

### Detailed Component Comparison

| Aspect | Arena v1 | Arena v2 |
|--------|----------|----------|
| **Job submission** | Helm chart rendering + `kubectl apply` shell-out | Direct K8s dynamic client API |
| **Configuration** | CLI flags only (no YAML) | YAML schema + CLI flags |
| **Helm dependency** | Yes — `helm.sh/helm/v3` + 23 charts | No |
| **kubectl dependency** | Yes — shells out to `kubectl apply/delete` | No — uses `client-go` directly |
| **Training Go code** | thousands of lines | a few hundred lines |
| **Helm charts** | 23 charts maintained alongside Go code | 0 |
| **Temp files per submission** | 3 (values.yaml, rendered manifest, app-info) | 0 |
| **Resource lifecycle** | 3-step delete (read ConfigMap → delete resources → delete ConfigMap) | 1-step delete (delete anchor resource → K8s GC cascades) |
| **Per-role config** | 40+ CLI flags (e.g., `--worker-image`, `--ps-cpu`, `--chief-memory`) | `chief/ps/evaluator/launcher/master/worker` block in YAML with independent per-role declaration |
| **Adding a new operator** | Implement Go types + argsbuilder + command + Helm chart + chart template | Implement `provider.Provider` interface |
| **Operator version coupling** | Tightly coupled — must update Go types when operator CRD API changes | Loosely coupled — `unstructured.Unstructured` is version-agnostic |
| **Binary size** | Larger (Helm SDK + transitive deps) | Smaller (only `client-go` + `cobra`) |

### Migration Notes

#### PyTorch: `worker.replicas` (v2) vs `--workers` (v1)

> **PyTorch-specific:** Only v1's PyTorch submit path subtracts the master from `--workers`.

**Key difference:** v1's `--workers=N` includes the master (internally decremented by 1), while v2's `worker.replicas=N` excludes it — the master is always an additional Pod.

**Migration formula:** `worker.replicas = --workers - 1`. Since `--workers` is always ≥ 1, when `--workers` is 1, `worker.replicas` would be 0, which violates the > 0 constraint. In this case, omit the worker block and specify the master block instead.

In v2, the PyTorch master is fixed at 1 replica, inherits worker config by default if `master` block is omitted, and can be independently configured via the `master` block. **Total GPU-using replicas** = `worker.replicas + 1`.

For framework-specific mappings across all providers, see the [worker.replicas → Provider Mapping](yaml-schema.md#workerreplicas--provider-mapping) section in the YAML schema specification.

### Test Plan

The test plan for Arena v2 includes:

**Unit tests:**

- YAML schema parsing and validation
- Flag-to-YAML override merging logic
- Provider CRD generation (PyTorchJob, MPIJob)
- K8s client wrapper error handling

**Integration tests:**

- End-to-end job submission with `--dry-run` flag
- Legacy `arena submit` command compatibility
- Multi-operator scenarios (submit PyTorchJob, then query status, then delete)

**E2E tests:**

- Submit real training jobs to a test cluster
- Verify CRD structure matches operator expectations
- Test job lifecycle (submit → running → completed/failed)
- Validate per-role configuration (worker vs master vs PS)

### Graduation Criteria

**Alpha/MVP (Phase 1 - current):**

- v1 remains the default for production use
- Core training submission (PyTorchJob, MPIJob) with YAML schema
  - without `scheduling` semantics support
- Basic commands (`arena version`, `arena help`)
- Display cluster resource info (`arena top`)
- Supports basic job manipulations and status queries (e.g., `arena job delete`, `arena job get`, `arena job list`, `arena job status`, `arena job logs`). The `arena job` prefix provides equivalent functionality to v1 `arena delete/get/list/status/logs` commands.
- Add new sub-commands
  - `job`
    - `run`, launch a training job from a YAML file
    - `suspend`, suspend a running training job (operation, not YAML config)
    - `resume`, resume a suspended training job (operation, not YAML config)
  - `check`, check if the cluster has the required CRD installed
- Dry-run mode for CRD inspection

**Beta (Phase 2 - next):**

- Legacy `arena submit` command for backward compatibility
- Support `scheduling` semantics
- Users begin migrating simple jobs to v2
- Comprehensive CLI flag coverage matching v1 functionality
- Output formats: JSON, YAML, wide
- InferenceTask kind (planned to use [RBG](https://github.com/sgl-project/rbg) for orchestration)
  - Add new sub-commands
    - `serve`
- Add DeepSpeed, TensorFlow, Ray support and examples
- Support overriding stable YAML fields via CLI flags (such as `name`, `worker.replicas`, etc.)

**Stable (Phase 3 - future):**

- Support overriding the entire YAML schema via CLI flags
- Platform-agnostic configuration via `arena configure`
- v1 is deprecated but still maintained for bug fixes
- Agent, evaluation, and benchmark task kinds
- `arena migrate` tool for converting v1 flag-based invocations to v2 YAML files
- Add Horovod support and examples

### Upgrade / Downgrade Strategy

**Upgrade path:**

- v2 and v1 coexist during the transition period
- Users can gradually migrate workloads from v1 to v2
- No breaking changes to existing v1 workflows
- Legacy `arena submit` command ensures backward compatibility

**Downgrade path:**

- If v2 introduces regressions, users can continue using v1 binary
- v2 does not modify v1 code or Helm charts
- Separate binary names (`arena-v2` during development) prevent conflicts

**Migration tooling:**

- Future `arena migrate` command will convert v1 flag-based invocations to v2 YAML files
- Helps users transition existing automation scripts without manual rewriting

### Version Skew Strategy

**Operator version compatibility:**

- v2 uses `unstructured.Unstructured` for CRD generation, making it version-agnostic
- No dependency on specific Training Operator API versions
- Works with Training Operator and MPI Operator

**Kubernetes version compatibility:**

- v2 depends only on `client-go` and `cobra`
- Supports any Kubernetes version that client-go supports
- No dependency on Helm or kubectl binaries

## Production Readiness Review Questionnaire

### Feature Enablement and Rollback

**How can this feature be enabled/disabled?**

- v2 is a separate binary (`arena-v2`) during development, allowing users to opt-in
- Once v2 reaches feature parity, it replaces the `arena` v1 binary
- Users can continue using v1 by not installing v2

**Can the feature be disabled once it has been enabled?**

- Yes, users can simply not use the v2 binary
- v2 does not modify v1 code or configuration
- No cluster-level changes are required

### Rollout, Upgrade and Rollback Planning

**How will this be rolled out?**

- v2 is being developed in the `v2` branch, in parallel with v1
- Users can test v2 in non-production environments first
- Gradual adoption as v2 reaches feature parity with v1

**What is the rollout plan?**

- Phase 1 (current): Core training submission, v1 remains default
- Phase 2 (next): Advanced features, users begin migrating simple jobs
- Phase 3 (future): Full feature parity, v1 deprecated

**How will rollback work if something goes wrong?**

- Users can revert to the v1 binary at any time
- v2 does not modify cluster state or v1 configuration
- No data migration or rollback procedures required

### Monitoring Requirements

**How can an operator determine if the feature is working correctly?**

- v2 provides `--dry-run` flag for inspecting generated CRDs before submission
- Job submission logs show API calls and responses
- `arena job get` and `arena job list` commands show job status

**How can this feature be monitored?**

- Kubernetes metrics for created CRDs (PyTorchJob, MPIJob)
- Job completion/failure rates
- Submission latency (time from CLI invocation to CRD creation)

### Dependencies

**Does this feature depend on any specific services or components?**

- v2 depends only on `client-go` for Kubernetes API access
- No dependency on Helm, kubectl, or external binaries
- Training Operator and MPI Operator are mandatory. KubeRay will become a requirement when Ray support is added.

**Are there any new dependencies introduced?**

- No new dependencies beyond `client-go` and `cobra`
- v2 removes the Helm dependency entirely

### Scalability

**How does this feature scale?**

- v2 generates CRDs locally and submits them to the Kubernetes API
- No local resource consumption beyond the CLI process
- Scalability is limited by Kubernetes API server capacity, not v2

**What are the scalability limits?**

- Same as v1: limited by Kubernetes cluster capacity and operator performance
- v2 may be slightly faster due to elimination of Helm rendering and shell-outs

### Troubleshooting

**How can issues be diagnosed?**

- `--dry-run` flag shows generated CRD before submission
- Verbose logging mode (if implemented) shows API calls and responses
- `arena job get <job-name>` shows job status and pod information
- `arena job logs <job-name>` shows training logs

**What are common failure modes?**

- CRD validation errors (malformed task spec)
- Kubernetes API errors (permission denied, resource quota exceeded)
- Operator errors (job submission failed, pod scheduling failed)
- Network errors (cluster unreachable)

## References

- **Arena v1 codebase**: https://github.com/kubeflow/arena/tree/v0.15.4
- **YAML schema specification**: `keps/arena-v2/yaml-schema.md` (YAML Schema section)
- **CLI flag mapping**: `keps/arena-v2/cli-flag-mapping.md`
- **Kubeflow Trainer**: https://github.com/kubeflow/trainer/tree/release-1.9
- **Kubeflow MPI Operator**: https://github.com/kubeflow/mpi-operator/tree/master
- **client-go**: https://github.com/kubernetes/client-go
