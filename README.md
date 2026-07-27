# Arena

## Overview

Arena is a command-line interface for running machine learning training jobs on Kubernetes. It generates Operator CRDs (PyTorchJob, TFJob, MPIJob) directly and manages the full job lifecycle without Helm dependencies.

Arena v2 uses a YAML-first configuration approach — you define your training job in a single YAML file and submit it with one command. No Helm charts, no complex flag parsing.

## Features

- **YAML-first job submission** with `--set` overrides for runtime customization
- **Multi-framework support**: PyTorch, TensorFlow, MPI (DeepSpeed, Horovod)
- **Full job lifecycle**: run, list, get, logs, suspend, resume, delete
- **Scheduling controls**: node selectors, tolerations, affinity, gang scheduling, priority classes
- **Dry-run mode** for inspecting generated CRDs before submission
- **Output formats**: table (default), wide, JSON, YAML
- **Cluster checks**: `arena check` verifies required CRDs; `arena top job` shows GPU usage
- **Shell completions**: bash, zsh, fish, powershell
- **Backward-compatible** `arena submit` command for v1 users

## Documentation

- [Installation](docs/installation.md) — set up K8s operators and arena binary
- [Quick Start](docs/quickstart.md) — submit your first training job
- [YAML Configuration](docs/yaml-schema.md) — all YAML fields and options
- [Framework Guides](docs/frameworks.md) — PyTorch, TensorFlow, MPI usage
- [CLI Reference](docs/cli-reference.md) — all commands and flags
- [Best Practices](docs/best-practices.md) — practical guidance for effective usage
- [Monitoring & Status](docs/monitoring.md) — observe and troubleshoot jobs
- [Advanced Configuration](docs/advanced.md) — scheduling, storage, env, lifecycle, security
- [Migration Guide](docs/migration.md) — migrate from Arena v1 to v2
- [FAQ](docs/faq.md) — common questions and troubleshooting
- [Design Document](keps/arena-v2/README.md) — architecture and migration guide

## Installation

### Build from source

Prerequisites: [Go](https://go.dev/dl/) >= 1.26.5, [make](https://www.gnu.org/software/make/)

```shell
git clone https://github.com/kubeflow/arena.git
cd arena
make arena-v2
```

The binary is at `bin/arena-v2`. Add it to your `$PATH`:

```shell
export PATH=$(pwd)/bin:$PATH
mv bin/arena-v2 bin/arena   # optional: rename to 'arena'
```

Or install directly to `$GOPATH/bin`:

```shell
make v2-install
```

For full installation instructions including K8s operator setup, see the [Installation guide](docs/installation.md).

## Quick Start

```shell
arena check                                          # verify CRDs are installed
arena job run -f examples/v2/quickstart/pytorch-simple.yaml     # submit a training job
arena job list                                        # view running jobs
```

See the [Quick Start guide](docs/quickstart.md) for a full walkthrough.

## Developing

```shell
make arena-v2              # build
make v2-test               # unit tests
make v2-vet                # go vet
make v2-fmt                # format code
make v2-lint               # golangci-lint
make v2-e2e-test           # e2e tests (requires K8s cluster)
```

## Status

Arena v2 is in **Alpha** (MVP phase). It is not production ready. APIs and YAML schema may change between releases. See [keps/arena-v2/README.md](keps/arena-v2/README.md) for the full design document and graduation criteria.
