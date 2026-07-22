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
GIT_TAG := $(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE := $(shell if [ -z "`git status --porcelain`" ]; then echo "clean"; else echo "dirty"; fi)

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

# Pinned golangci-lint version for v2-lint target
GOLANGCI_LINT_VERSION ?= v2.12.2

# Version info injected via ldflags at build time
# Keep in sync with .goreleaser.yaml ldflags (Makefile uses shell vars, GoReleaser uses template vars)
V2_LDFLAGS := -X ${PACKAGE}/pkg/cli.version=${VERSION} \
  -X ${PACKAGE}/pkg/cli.gitCommit=${GIT_SHORT_COMMIT} \
  -X ${PACKAGE}/pkg/cli.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}/pkg/cli.gitTag=${GIT_TAG} \
  -X ${PACKAGE}/pkg/cli.gitTreeState=${GIT_TREE_STATE}

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
	go vet -tags v2e2e ./test/e2e/

.PHONY: v2-fmt
v2-fmt: ## Run gofmt on arena v2 packages.
	@echo "Running gofmt on arena v2 packages..."
	gofmt -w $(V2_ALL_PACKAGES)

.PHONY: v2-lint
v2-lint: ## Run golangci-lint on arena v2 packages.
	@echo "Running golangci-lint on arena v2 packages..."
	@which golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	golangci-lint run $(V2_ALL_PACKAGES)

.PHONY: v2-install
v2-install: ## Install arena v2 CLI to GOBIN.
	@echo "Installing arena v2 CLI to $(GOBIN)..."
	go install -ldflags '$(V2_LDFLAGS)' ./cmd/arena-v2/

# E2E cluster configuration (overridable via environment or command line)
TRAINER_REPO ?= https://github.com/kubeflow/training-operator.git
TRAINER_REF ?= v1.9.3
INSTALL_METHOD ?= kustomize
K8S_VERSION ?= 1.32.3
KIND_CLUSTER_NAME ?= arena-v2
NAMESPACE ?= kubeflow

.PHONY: v2-e2e-setup-cluster
v2-e2e-setup-cluster: ## Create kind cluster and install training-operator for v2 e2e tests.
	TRAINER_REPO=$(TRAINER_REPO) \
	TRAINER_REF=$(TRAINER_REF) \
	INSTALL_METHOD=$(INSTALL_METHOD) \
	K8S_VERSION=$(K8S_VERSION) \
	KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) \
	NAMESPACE=$(NAMESPACE) \
	./hack/e2e-setup-cluster.sh

.PHONY: v2-e2e-test
v2-e2e-test: arena-v2 ## Run arena v2 e2e tests (requires cluster from v2-e2e-setup-cluster).
	@echo "Running arena v2 e2e tests..."
	go test -tags v2e2e ./test/e2e/ -v -ginkgo.v -timeout 30m

.PHONY: v2-e2e-teardown
v2-e2e-teardown: ## Delete the kind cluster used for v2 e2e tests.
	kind delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY: v2-release-snapshot
v2-release-snapshot: ## Build arena v2 release snapshot locally (no tag required).
	@goreleaser release --snapshot --clean