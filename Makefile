.SILENT:

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

PACKAGE ?= github.com/kubeflow/arena
CURRENT_DIR ?= $(shell pwd)
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

VERSION ?= $(shell cat VERSION 2>/dev/null || echo "0.0.0-dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_SHORT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Location to install binaries
LOCALBIN ?= $(CURRENT_DIR)/bin

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Arena v2

# Shared package lists for v2 targets
V2_PACKAGES := ./pkg/constants/ ./pkg/log/ ./pkg/cli/ ./pkg/task/ ./pkg/provider/ ./pkg/client/ ./pkg/output/
V2_ALL_PACKAGES := $(V2_PACKAGES) ./cmd/arena-v2/

# Version info injected via ldflags at build time
V2_LDFLAGS := -X ${PACKAGE}/pkg/cli.version=${VERSION} \
  -X ${PACKAGE}/pkg/cli.gitCommit=${GIT_SHORT_COMMIT} \
  -X ${PACKAGE}/pkg/cli.buildDate=${BUILD_DATE}

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: arena-v2
arena-v2: $(LOCALBIN) ## Build arena v2 CLI for current platform.
	@echo "Building arena v2 CLI..."
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags '$(V2_LDFLAGS)' -o $(LOCALBIN)/arena-v2 ./cmd/arena-v2/

.PHONY: v2-test
v2-test: ## Run arena v2 unit tests.
	@echo "Running arena v2 unit tests..."
	go test $(V2_PACKAGES) -v

.PHONY: v2-vet
v2-vet: ## Run go vet on arena v2 packages.
	@echo "Running go vet on arena v2 packages..."
	go vet $(V2_ALL_PACKAGES)

.PHONY: v2-integration-test
v2-integration-test: ## Run arena v2 integration tests (no cluster required).
	@echo "Running arena v2 integration tests..."
	go test -tags integration $(V2_PACKAGES) -run TestIntegration -v -timeout 5m

.PHONY: v2-install
v2-install: ## Install arena v2 CLI to GOBIN.
	@echo "Installing arena v2 CLI to $(GOBIN)..."
	go install -ldflags '$(V2_LDFLAGS)' ./cmd/arena-v2/

# Portable SHA-256 checksum command.
# macOS ships a BSD sha256sum that can't read checksums from stdin pipes,
# so use shasum on Darwin and sha256sum everywhere else.
ifeq ($(shell uname -s 2>/dev/null),Darwin)
SHA256CMD := shasum -a 256
else
SHA256CMD := sha256sum
endif

# Pinned to immutable commit SHAs to prevent supply-chain risk from tag relocation.
# Override with: make v2-e2e-setup TRAINER_SHA=<sha> MPI_OPERATOR_SHA=<sha>
# trainer v1.9.3
TRAINER_SHA ?= c77ee3ff4d21aa8e5027580d83361c1a426ee789
# mpi-operator v0.8.0
MPI_OPERATOR_SHA ?= 101db144e019e1de65aaded07745e9d805c2852a

# Expected SHA256 checksums for the pinned CRD versions above.
# Update these when bumping TRAINER_SHA or MPI_OPERATOR_SHA.
PYTORCHJOBS_SHA256 ?= eb80e57a984cdc7ceefaaf3fc561e21b673c30cbb58589f2389528d2b2760d22
TFJOBS_SHA256 ?= 1fea6d2d9d3fec4697db322d4490be67b0d616478f42566e4c48ae865130a8c8
MPIJOBS_SHA256 ?= 78b0131f931fcc8d99f6ce0407281f41b62527e443ec691d26d9634b390ad3bb

.PHONY: v2-e2e-setup
v2-e2e-setup: ## Download Kubeflow Training Operator and MPI Operator CRDs for v2 e2e tests.
	@mkdir -p test/e2e/crds
	@echo "Downloading Kubeflow Training Operator CRDs..."
	@curl --fail -sSL https://raw.githubusercontent.com/kubeflow/trainer/$(TRAINER_SHA)/manifests/base/crds/kubeflow.org_pytorchjobs.yaml -o test/e2e/crds/pytorchjobs.yaml
	@echo "$(PYTORCHJOBS_SHA256)  test/e2e/crds/pytorchjobs.yaml" | $(SHA256CMD) --check --quiet || { echo "ERROR: pytorchjobs.yaml checksum mismatch"; exit 1; }
	@curl --fail -sSL https://raw.githubusercontent.com/kubeflow/trainer/$(TRAINER_SHA)/manifests/base/crds/kubeflow.org_tfjobs.yaml -o test/e2e/crds/tfjobs.yaml
	@echo "$(TFJOBS_SHA256)  test/e2e/crds/tfjobs.yaml" | $(SHA256CMD) --check --quiet || { echo "ERROR: tfjobs.yaml checksum mismatch"; exit 1; }
	@echo "Downloading Kubeflow MPI Operator CRDs..."
	@curl --fail -sSL https://raw.githubusercontent.com/kubeflow/mpi-operator/$(MPI_OPERATOR_SHA)/manifests/base/kubeflow.org_mpijobs.yaml -o test/e2e/crds/mpijobs.yaml
	@echo "$(MPIJOBS_SHA256)  test/e2e/crds/mpijobs.yaml" | $(SHA256CMD) --check --quiet || { echo "ERROR: mpijobs.yaml checksum mismatch"; exit 1; }
	@echo "CRDs downloaded and verified in test/e2e/crds/"

.PHONY: v2-e2e-test
v2-e2e-test: v2-e2e-setup arena-v2 ## Run arena v2 e2e tests (requires real K8s cluster).
	@echo "Running arena v2 e2e tests..."
	go test -tags v2e2e ./test/e2e/ -v -ginkgo.v -timeout 30m
