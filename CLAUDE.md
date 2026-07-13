# Arena v2 — AI Workload CLI for Kubernetes

## Project Overview

Arena v2 is a lightweight CLI for submitting AI training/inference tasks to Kubernetes.
It directly generates Operator CRDs (PyTorchJob, TFJob, MPIJob) and related resources without Helm dependencies.

- **Module**: `github.com/kubeflow/arena`
- **Branch**: `v2`
- **Entry point**: `cmd/arena-v2/main.go`
- **Existing v1 code Branch**: `v0.15.4` (do NOT modify)

## Why Arena v2

Arena v2 eliminates Helm dependencies, reduces code complexity, and adds YAML-first configuration. See [keps/arena-v2/README.md](keps/arena-v2/README.md) for the full design rationale, v1 vs v2 technical comparison, and upgrade path.

## Build & Test

Use bazel to build Arena v2 cli, every single build and test action must add into makefile.

```bash
# Build the v2 binary
go build -o bin/arena-v2 ./cmd/arena-v2/

# Build with version info
go build -ldflags "-X github.com/kubeflow/arena/pkg/cli.versionGitCommit=$(git rev-parse --short HEAD) \
  -X github.com/kubeflow/arena/pkg/cli.versionBuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/arena-v2 ./cmd/arena-v2/

# Vet all v2 packages
go vet ./pkg/cli/ ./pkg/task/ ./pkg/provider/ ./pkg/client/ ./pkg/output/ ./cmd/arena-v2/

# Dry-run test (no cluster required)
./bin/arena-v2 job run -f examples/v2/pytorch-train.yaml --dry-run
./bin/arena-v2 submit pytorch --name test --image pytorch:2.1 --gpus 2 --dry-run
./bin/arena-v2 submit pytorch --name test --image pytorch:2.1 --dry-run "python train.py"

# Run tests (when added)
go test ./pkg/task/ ./pkg/provider/ ./pkg/client/ -v
```

## DO NOT

1. **Do NOT modify v1 code.** Old Arena code lives in `cmd/arena/`, `pkg/commands/`, `pkg/apis/`, `pkg/training/`, `pkg/serving/`, `pkg/argsbuilder/`. Leave it untouched. All v2 work goes in the new packages listed above.

2. **Do NOT add Helm as a dependency.** The entire point of v2 is eliminating Helm. CRDs are built as Go structs → `unstructured.Unstructured`. Never use `helm.sh/helm/v3` in v2 code paths.

3. **Do NOT import v1 packages from v2 code.** v2 packages (`pkg/cli`, `pkg/task`, `pkg/provider`, `pkg/client`, `pkg/output`) must be self-contained. No imports from `pkg/commands`, `pkg/apis`, `pkg/training`, etc.

4. **Do NOT use Kubeflow Training Operator Go types directly.** Use `unstructured.Unstructured` for CRD construction. This avoids a hard dependency on specific Operator API versions and makes the code resilient to Operator upgrades.

5. **Do NOT add cloud provider SDKs.** Arena v2 is K8s-native only. No AWS/GCP/Azure SDK imports.

6. **Do NOT create a server component.** Arena v2 is client-only (like kubectl). All cluster interaction goes through the K8s API.

7. **Do NOT use the `arena-v2` binary name in committed code paths.** The build target is `arena` — the `-v2` suffix is temporary for development alongside v1.

8. **Do NOT commit the compiled binary.** Add `arena-v2` to `.gitignore` if needed.

9. **Do NOT break the `submit` legacy compat.** The `arena submit <type> --flags` interface must remain backward compatible with old Arena usage patterns.

10. **Do NOT add Docker/container runtime dependencies.** Arena v2 creates K8s resources only; it does not build or push container images.
