.SILENT:

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

PACKAGE ?= github.com/kubeflow/arena
CURRENT_DIR ?= $(shell pwd)
DIST_DIR ?= ${CURRENT_DIR}/bin
ARENA_CLI_NAME ?= arena
JOB_MONITOR ?= jobmon
ARENA_UNINSTALL ?= arena-uninstall
OS ?= linux
ARCH ?= amd64

VERSION ?= $(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
DOCKER_BUILD_DATE := $(shell date -u +'%Y%m%d%H%M%S')
GIT_TAG := $(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE := $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
PACKR_CMD := $(shell if [ "`which packr`" ]; then echo "packr"; else echo "go run vendor/github.com/gobuffalo/packr/packr/main.go"; fi)

## Location to install binaries
LOCALBIN ?= $(shell pwd)/bin

# Versions
GOLANG_VERSION=$(shell grep -e '^go ' go.mod | cut -d ' ' -f 2)
KUBECTL_VERSION ?= 1.28.4
HELM_VERSION ?= 3.13.3
GOLANGCI_LINT_VERSION ?= 1.57.2

# Binaries
ARENA ?= $(LOCALBIN)/arena
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

BUILDER_IMAGE=arena-builder
BASE_IMAGE=registry.aliyuncs.com/kubeflow-images-public/tensorflow-1.12.0-notebook-gpu:v0.4.0
# NOTE: the volume mount of ${DIST_DIR}/pkg below is optional and serves only
# to speed up subsequent builds by caching ${GOPATH}/pkg between builds.
BUILDER_CMD=docker run --rm \
  -v ${CURRENT_DIR}:/root/go/src/${PACKAGE} \
  -v ${DIST_DIR}/pkg:/root/go/pkg \
  -w /root/go/src/${PACKAGE} ${BUILDER_IMAGE}

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE} \
  -extldflags "-static"

# docker image publishing options
IMAGE_REGISTRY ?= docker.io
IMAGE_REPOSITORY ?= kubeflow/arena
IMAGE_TAG ?= $(VERSION)
IMAGE ?= $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY):$(IMAGE_TAG:-latest)
DOCKER_PUSH=false

ifneq (${GIT_TAG},)
IMAGE_TAG=${GIT_TAG}
override LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif
ifneq (${IMAGE_NAMESPACE},)
override LDFLAGS += -X ${PACKAGE}/cmd/arena/commands.imageNamespace=${IMAGE_NAMESPACE}
endif
ifneq (${IMAGE_TAG},)
override LDFLAGS += -X ${PACKAGE}/cmd/arena/commands.imageTag=${IMAGE_TAG}
endif

ifeq (${DOCKER_PUSH},true)
ifndef IMAGE_NAMESPACE
$(error IMAGE_NAMESPACE must be set to push images (e.g. IMAGE_NAMESPACE=arena))
endif
endif

ifdef IMAGE_NAMESPACE
IMAGE_PREFIX=${IMAGE_NAMESPACE}/
endif

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: go-fmt
go-fmt: ## Run go fmt against code.
	@echo "Running go fmt..."
	go fmt ./...

.PHONY: go-vet
go-vet: ## Run go vet against code.
	@echo "Running go vet..."
	go vet ./...

.PHONY: go-lint
go-lint: golangci-lint ## Run golangci-lint linter.
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run --timeout 5m --go 1.21 ./...

.PHONY: unit-test
unit-test: ## Run go unit tests.
	@echo "Running go test..."
	go test ./... -coverprofile cover.out

# Build the project
.PHONY: default
default:
ifeq ($(OS),Windows_NT)
default: arena-windows
else
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
default: arena-linux-amd64
else ifeq ($(UNAME_S),Darwin)
default: arena-darwin-amd64
else
$(error "The OS is not supported")
endif
endif

##@ Build

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: arena
arena: ## Build arena CLI for current platform.
	@echo "Building arena CLI..."
	CGO_ENABLED=0 go build -tags netgo -ldflags '${LDFLAGS}' -o $(LOCALBIN)/arena cmd/arena/main.go

.PHONY: arena-linux-amd64
arena-linux-amd64: ## Build arena CLI for linux/amd64.
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags '${LDFLAGS}' -o $(LOCALBIN)/arena-linux-amd64-v$(VERSION) cmd/arena/main.go

.PHONY: arena-linux-arm64
arena-linux-arm64: ## Build arena CLI for linux/arm64.
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags netgo -ldflags '${LDFLAGS}' -o $(LOCALBIN)/arena-linux-arm64-v$(VERSION) cmd/arena/main.go

.PHONY: arena-darwin-amd64
arena-darwin-amd64: ## Build arena CLI for darwin/amd64.
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags netgo -ldflags '${LDFLAGS}' -o $(LOCALBIN)/arena-darwin-amd64-v$(VERSION) cmd/arena/main.go

.PHONY: arena-darwin-arm64
arena-darwin-arm64: ## Build arena CLI for darwin/arm64.
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -tags netgo -ldflags '${LDFLAGS}' -o $(LOCALBIN)/arena-darwin-arm64-v$(VERSION) cmd/arena/main.go

.PHONY: build-java-sdk
build-java-sdk: ## Build Java SDK.
	echo "Building arena Java SDK..."
	mvn package -Dmaven.test.skip=true -Dgpg.skip -f sdk/arena-java-sdk

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build \
		-t $(IMAGE) \
		-f Dockerfile.install .

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push $(IMAGE)

.PHONY: notebook-image-kubeflow
notebook-image-kubeflow:
	docker build --build-arg "BASE_IMAGE=${BASE_IMAGE}" -t cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-gpu -f Dockerfile.notebook.gpu .
	docker tag cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-gpu cheyang/arena-notebook:kubeflow

.PHONY: notebook-image
notebook-image:
	docker build --build-arg "BASE_IMAGE=tensorflow/tensorflow:1.12.0-devel-py3" -t cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-cpu -f Dockerfile.notebook.cpu .
	docker tag cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-cpu cheyang/arena-notebook:cpu

.PHONY: build-pkg
build-pkg:
	docker build \
		--build-arg OS=${OS} \
		--build-arg ARCH=${ARCH} \
		--build-arg COMMIT=${GIT_SHORT_COMMIT} \
		--build-arg VERSION=${VERSION} \
		--build-arg GOLANG_VERSION=${GOLANG_VERSION} \
		--build-arg KUBECTL_VERSION=${KUBECTL_VERSION} \
		--build-arg HELM_VERSION=${HELM_VERSION} \
		-t arena-build:${VERSION}-${GIT_SHORT_COMMIT}-${OS}-${ARCH} \
		-f Dockerfile.build \
		.
	docker run -itd --name=arena-pkg arena-build:${VERSION}-${GIT_SHORT_COMMIT}-${OS}-${ARCH} /bin/bash
	docker cp arena-pkg:/workspace/arena-installer-v${VERSION}-${GIT_SHORT_COMMIT}-${OS}-${ARCH}.tar.gz .
	docker rm -f arena-pkg

build-dependabot:
	python3 hack/create_dependabot.py

##@ Dependencies

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@v$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
