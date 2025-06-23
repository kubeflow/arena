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
DIST_DIR ?= $(CURRENT_DIR)/bin
ARENA_CLI_NAME ?= arena
JOB_MONITOR ?= jobmon
ARENA_UNINSTALL ?= arena-uninstall
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

VERSION ?= $(shell cat VERSION)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
DOCKER_BUILD_DATE := $(shell date -u +'%Y%m%d%H%M%S')
GIT_TAG := $(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE := $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
PACKR_CMD := $(shell if [ "`which packr`" ]; then echo "packr"; else echo "go run vendor/github.com/gobuffalo/packr/packr/main.go"; fi)

# Location to install binaries
LOCALBIN ?= $(CURRENT_DIR)/bin
# Location to put temp files
TEMPDIR ?= $(CURRENT_DIR)/tmp
# ARENA_ARTIFACTS
ARENA_ARTIFACTS_CHART_PATH ?= $(CURRENT_DIR)/arena-artifacts

# Versions
GOLANG_VERSION=$(shell grep -e '^go ' go.mod | cut -d ' ' -f 2)
KUBECTL_VERSION ?= v1.28.4
HELM_VERSION ?= v3.13.3
HELM_UNITTEST_VERSION ?= 0.5.1
KIND_VERSION ?= v0.23.0
KIND_K8S_VERSION ?= v1.29.3
ENVTEST_VERSION ?= release-0.18
ENVTEST_K8S_VERSION ?= 1.29.3
GOLANGCI_LINT_VERSION ?= v1.57.2

# Binaries
ARENA ?= arena-v$(VERSION)-$(OS)-$(ARCH)
KUBECTL ?= kubectl-$(KUBECTL_VERSION)-$(OS)-$(ARCH)
HELM ?= helm-$(HELM_VERSION)-$(OS)-$(ARCH)
KIND ?= $(LOCALBIN)/kind-$(KIND_VERSION)
ENVTEST ?= $(LOCALBIN)/setup-envtest-$(ENVTEST_VERSION)
GOLANGCI_LINT ?= golangci-lint-$(GOLANGCI_LINT_VERSION)

# Tarballs
ARENA_INSTALLER ?= arena-installer-$(VERSION)-$(OS)-$(ARCH)
ARENA_INSTALLER_TARBALL ?= $(ARENA_INSTALLER).tar.gz

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
IMAGE ?= $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY):$(IMAGE_TAG)
DOCKER_PUSH=false
BASE_IMAGE ?= debian:12-slim

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

.PHONY: all
all: go-fmt go-vet go-lint unit-test e2e-test

##@ Development

go-fmt: ## Run go fmt against code.
	@echo "Running go fmt..."
	go fmt ./...

go-vet: ## Run go vet against code.
	@echo "Running go vet..."
	go vet ./...

.PHONY: go-lint
go-lint: golangci-lint ## Run golangci-lint linter.
	@echo "Running golangci-lint run..."
	$(LOCALBIN)/$(GOLANGCI_LINT) run --timeout 5m ./...

.PHONY: go-lint-fix
go-lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	@echo "Running golangci-lint run --fix..."
	$(LOCALBIN)/$(GOLANGCI_LINT) run --fix --timeout 5m ./...

.PHONY: unit-test
unit-test: ## Run go unit tests.
	@echo "Running go test..."
	go test $(shell go list ./... | grep -v /e2e) -coverprofile cover.out

.PHONY: e2e-test
e2e-test: envtest ## Run the e2e tests against a Kind k8s instance that is spun up.
	@echo "Running e2e tests..."
	go test ./test/e2e/ -v -ginkgo.v -timeout 30m

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

$(TEMPDIR):
	mkdir -p $(TEMPDIR)

clean: ## Clean up all downloaded and generated files.
	rm -rf $(LOCALBIN) $(TEMPDIR)

.PHONY: arena
arena: $(LOCALBIN) ## Build arena CLI for current platform.
	@echo "Building arena CLI..."
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -tags netgo -ldflags '${LDFLAGS}' -o $(LOCALBIN)/$(ARENA) cmd/arena/main.go

.PHONY: java-sdk
java-sdk: ## Build Java SDK.
	echo "Building arena Java SDK..."
	mvn package -Dmaven.test.skip=true -Dgpg.skip -f sdk/arena-java-sdk

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--tag $(IMAGE) \
		-f Dockerfile \
		.

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push $(IMAGE)

.PHONY: docker-buildx
PLATFORMS ?= linux/amd64,linux/arm64
docker-buildx: ## Build and push docker images for multiple platforms.
	- $(CONTAINER_TOOL) buildx create --name arena-builder
	$(CONTAINER_TOOL) buildx use arena-builder
	- $(CONTAINER_TOOL) buildx build --push \
		--platform=$(PLATFORMS) \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--tag $(IMAGE) \
		-f Dockerfile \
		.
	- $(CONTAINER_TOOL) buildx rm arena-builder

.PHONY: notebook-image-kubeflow
notebook-image-kubeflow:
	docker build --build-arg "BASE_IMAGE=${BASE_IMAGE}" -t cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-gpu -f Dockerfile.notebook.gpu .
	docker tag cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-gpu cheyang/arena-notebook:kubeflow

.PHONY: notebook-image
notebook-image:
	docker build --build-arg "BASE_IMAGE=tensorflow/tensorflow:1.12.0-devel-py3" -t cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-cpu -f Dockerfile.notebook.cpu .
	docker tag cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-cpu cheyang/arena-notebook:cpu

.PHONY: build-dependabot
build-dependabot:
	python3 hack/create_dependabot.py

.PHONY: arena-installer
arena-installer: $(ARENA_INSTALLER_TARBALL) ## Build arena installer tarball
$(ARENA_INSTALLER_TARBALL): arena kubectl helm
	echo "Building arena installer tarball..." && \
	rm -rf $(TEMPDIR)/$(ARENA_INSTALLER) && \
	mkdir -p $(TEMPDIR)/$(ARENA_INSTALLER)/bin && \
	cp $(LOCALBIN)/$(ARENA) $(TEMPDIR)/$(ARENA_INSTALLER)/bin/arena && \
	cp $(LOCALBIN)/$(KUBECTL) $(TEMPDIR)/$(ARENA_INSTALLER)/bin/kubectl && \
	cp $(LOCALBIN)/$(HELM) $(TEMPDIR)/$(ARENA_INSTALLER)/bin/helm && \
	cp -R charts $(TEMPDIR)/$(ARENA_INSTALLER) && \
	cp -R arena-artifacts $(TEMPDIR)/$(ARENA_INSTALLER) && \
	cp arena-gen-kubeconfig.sh $(TEMPDIR)/$(ARENA_INSTALLER)/bin && \
	cp install.sh $(TEMPDIR)/$(ARENA_INSTALLER) && \
	cp uninstall.sh $(TEMPDIR)/$(ARENA_INSTALLER)/bin/arena-uninstall && \
	tar -zcf $(ARENA_INSTALLER).tar.gz -C $(TEMPDIR) $(ARENA_INSTALLER) && \
	echo "Successfully saved arena installer to $(ARENA_INSTALLER).tar.gz."
	
##@ Helm

.PHONY: helm-unittest
helm-unittest: helm-unittest-plugin ## Run Helm chart unittests.
	set -x && $(LOCALBIN)/$(HELM) unittest $(ARENA_ARTIFACTS_CHART_PATH) --strict --file "tests/**/*_test.yaml" --chart-tests-path $(CURRENT_DIR)
	
##@ Dependencies

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN)/$(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(LOCALBIN)/$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(LOCALBIN)/$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: kubectl
kubectl: $(LOCALBIN)/$(KUBECTL)
$(LOCALBIN)/$(KUBECTL): $(LOCALBIN) $(TEMPDIR)
	$(eval KUBECTL_URL=https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(OS)/$(ARCH)/kubectl)
	$(eval KUBECTL_SHA_URL=$(KUBECTL_URL).sha256)

	cd $(TEMPDIR) && \
	echo "Download $(KUBECTL) if not present..." && \
	if [ ! -f $(KUBECTL) ]; then \
		curl -sSLo $(KUBECTL) $(KUBECTL_URL); \
	fi && \
	echo "Download $(KUBECTL).sha256 if not present..." && \
	if [ ! -f kubectl.sha256 ]; then \
		curl -sSLo $(KUBECTL).sha256 $(KUBECTL_SHA_URL); \
	fi && \
	echo "Verifying checksum..." && \
	echo -n "$$(cat $(KUBECTL).sha256)  $(KUBECTL)" | shasum -a 256 --check --quiet || (echo "Checksum verification failed, exiting." && false) && \
	echo "Make kubectl executable and move it to bin directory..." && \
	chmod +x $(KUBECTL) && \
	cp $(KUBECTL) $(LOCALBIN) && \
	echo "Successfully installed kubectl to $(LOCALBIN)/$(KUBECTL)."

.PHONY: helm
helm: $(LOCALBIN)/$(HELM)
$(LOCALBIN)/$(HELM): $(LOCALBIN) $(TEMPDIR)
	$(eval HELM_URL=https://get.helm.sh/$(HELM).tar.gz)
	$(eval HELM_SHA_URL=https://get.helm.sh/$(HELM).tar.gz.sha256sum)

	cd $(TEMPDIR) && \
	echo "Download $(HELM).tar.gz if not present..." && \
	if [ ! -f $(HELM).tar.gz ]; then \
		wget -qO $(HELM).tar.gz $(HELM_URL); \
	fi && \
	echo "Download $(HELM).tar.gz.sha256sum if not present..." && \
	if [ ! -f $(HELM).tar.gz.sha256sum ]; then \
		wget -qO $(HELM).tar.gz.sha256sum $(HELM_SHA_URL); \
	fi && \
	echo "Verifying checksum..." && \
	cat $(HELM).tar.gz.sha256sum | shasum -a 256 --check --quiet || (echo "Checksum verification failed, exiting." && false) && \
	echo "Extract helm tarball and move it to bin directory..." && \
	tar -zxf $(HELM).tar.gz && \
	cp ${OS}-${ARCH}/helm $(LOCALBIN)/$(HELM) && \
	echo "Successfully installed helm to $(LOCALBIN)/$(HELM)."

.PHONY: helm-unittest-plugin
helm-unittest-plugin: helm ## Download helm unittest plugin locally if necessary.
	if [ -z "$(shell $(LOCALBIN)/$(HELM) plugin list | grep unittest)" ]; then \
		echo "Installing helm unittest plugin"; \
		$(LOCALBIN)/$(HELM) plugin install https://github.com/helm-unittest/helm-unittest.git --version $(HELM_UNITTEST_VERSION); \
	fi

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
