# PR #1469 Review Fixes Design

## Goal

Fix all actionable issues identified in the PR #1469 review comments on the `pr/v2-5-cli-core` branch, backporting relevant test infrastructure from `pr/v2-7-examples-tests`.

## Background

PR #1469 (feat: add CLI core with submission, query, and output formatting) received 14 line-level review comments identifying 7 distinct issues. Two issues (go.mod deps, e2e tests) are partially addressed in v7; five issues (typosquat deps, partial-failure rollback, extractPods, getRealPods error swallowing, :latest tags) are open as of v7.

## Issues and Fixes

### 1. go.yaml.in typosquat dependencies in go.mod

**Problem:** `go.mod` lines 37-38 declare `go.yaml.in/yaml/v2 v2.4.4` and `go.yaml.in/yaml/v3 v3.0.4` as indirect dependencies. These are non-canonical modules that look like typosquats of `gopkg.in/yaml.v2` and `gopkg.in/yaml.v3`.

**Fix:** Remove both `go.yaml.in` entries from go.mod, then run `go mod tidy` to regenerate go.sum with only canonical dependencies.

**Files:** `go.mod`, `go.sum`

### 2. createJobResources partial-failure leaves orphaned CRD

**Problem:** `createJobResources` (lifecycle.go:20) creates the CRD first (in submit.go/run.go), then creates ConfigMap, RBAC resources, and TensorBoard resources. If any auxiliary creation fails, the CRD remains in the cluster with no cleanup. This produces a partially-submitted job that `get`/`list` cannot display correctly (they depend on the ConfigMap anchor).

**Fix:** Add a `defer` cleanup in `createJobResources` that deletes the CRD if the function returns an error. The cleanup uses `k8sClient.Delete` with the CRD's kind, namespace, and name. This ensures atomicity: either all resources are created, or the CRD is rolled back.

**Files:** `pkg/cli/lifecycle.go`, `pkg/cli/lifecycle_test.go` (new test), `pkg/cli/lifecycle_rbac_test.go` (new test)

### 3. extractPods reads wrong CRD status path

**Problem:** `extractPods` (list.go:191-227) reads `status.replicaStatuses.<role>.labelSelector` as a slice of pod maps. In Kubeflow training operators, `labelSelector` is a string (a label selector), not a list of pod entries. `replicaStatuses` only contains aggregate counts (`active`, `succeeded`, `failed`). The function is dead code that always returns nil.

**Fix:** Remove `extractPods` entirely. It is a broken fallback that never produces useful output. The primary pod source is `getRealPods`, which queries the K8s API directly. Callers of `extractPods` should rely solely on `getRealPods` and surface its errors (see fix #4).

**Files:** `pkg/cli/list.go`, `pkg/cli/list_test.go` (if tests reference extractPods)

### 4. getRealPods swallows all errors

**Problem:** `getRealPods` (get.go:129-157) returns `nil` for config load errors, clientset creation errors, and pod list errors — all without logging. Callers cannot distinguish "0 pods in cluster" from "API server unreachable" or "RBAC denied." When `getRealPods` returns nil, the CLI silently shows no pods.

**Fix:** Change `getRealPods` to return `([]client.PodInfo, error)`. Log each error via `klog.Warning` and return it to the caller. The caller (`getStatus` in get.go) should log the error and continue with an empty pod list (best-effort), but the user will see a warning explaining why pods are missing.

**Files:** `pkg/cli/get.go`, `pkg/cli/get_test.go` (if exists)

### 5. DefaultRsyncImage uses :latest tag

**Problem:** `DefaultRsyncImage = "instrumentisto/rsync-ssh:latest"` (constants.go:13) makes submissions non-reproducible.

**Fix:** Pin to a specific stable tag: `instrumentisto/rsync-ssh:3.4.0`. This matches the approach already used by `DefaultHDFSImage = "apache/hadoop:3.5.0"` and `DefaultTensorBoardImage = "tensorflow/tensorflow:2.21.0"`.

**Files:** `pkg/constants/constants.go`, `pkg/constants/constants_test.go` (if tests reference the value)

### 6. Missing e2e test coverage

**Problem:** PR #1469 has no tests under `test/e2e/`. The Makefile defines `v2-e2e-test` but the directory is empty.

**Fix:** Backport e2e test files from `pr/v2-7-examples-tests` that test v5-scope commands only (submit, run, get, list, status, dry-run). Skip tests that depend on v6 commands (delete, check, logs, top, suspend, resume). Also backport the `test/` integration tests that are v5-scope.

**Files from v7 to backport (e2e):**
- `test/e2e/suite_test.go`
- `test/e2e/dryrun_test.go`
- `test/e2e/format_test.go`
- `test/e2e/pytorch_test.go`
- `test/e2e/mpi_test.go`
- `test/e2e/yaml_submit_test.go`
- `test/e2e/cli_overrides_test.go`
- `test/e2e/features_test.go`
- `test/e2e/storage_test.go`
- `test/e2e/tfjob_test.go`

**Files from v7 to backport (integration):**
- `test/helpers_test.go`
- `test/loader_test.go`
- `test/flow_test.go`
- `test/output_test.go`
- `test/override_test.go`
- `test/scheduling_test.go`
- `test/storage_test.go`
- `test/validation_test.go`
- `test/multijob_test.go`
- `test/error_display_test.go`
- `test/testdata/*.yaml`

**Skip (v6-scope):**
- `test/e2e/yaml_delete_test.go` (uses `job delete`)
- `test/e2e/check_test.go` (uses `job check`)
- `test/stop_resume_test.go` (uses `job suspend`/`job resume`)

## Testing Strategy

All code fixes follow TDD: write failing test first, implement fix, verify pass.

- **Fix 1 (go.mod):** Verify with `go build ./cmd/arena-v2/` and `go mod tidy`
- **Fix 2 (rollback):** Unit test with fake client — inject RBAC creation failure, verify CRD is deleted
- **Fix 3 (extractPods):** Remove dead code; verify existing tests still pass
- **Fix 4 (getRealPods):** Unit test error return path; verify caller handles error
- **Fix 5 (:latest):** Update constant; verify existing tests pass
- **Fix 6 (e2e):** Backport tests; verify they compile (e2e tests require live cluster to run)

## Global Constraints

- Do NOT modify v1 code (`cmd/arena/`, `pkg/commands/`, etc.)
- Do NOT add Helm as a dependency
- Do NOT import v1 packages from v2 code
- Use `unstructured.Unstructured` for CRD construction
- Follow TDD: RED → GREEN → COMMIT
- All commits must include `Signed-off-by` trailer (DCO requirement)
