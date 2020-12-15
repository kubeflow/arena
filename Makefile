PACKAGE=github.com/kubeflow/arena
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/bin
ARENA_CLI_NAME=arena
JOB_MONITOR=jobmon
OS_ARCH?=linux-amd64

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_SHORT_COMMIT=$(shell git rev-parse --short HEAD)
DOCKER_BUILD_DATE=$(shell date -u +'%Y%m%d%H%M%S')
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
PACKR_CMD=$(shell if [ "`which packr`" ]; then echo "packr"; else echo "go run vendor/github.com/gobuffalo/packr/packr/main.go"; fi)

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
DOCKER_PUSH=false
IMAGE_TAG=latest

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

# Build the project
.PHONY: default
default:
ifeq ($(OS),Windows_NT)
default: cli-windows
else
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
$(info "Building on Linux")
default: cli-linux-amd64
else ifeq ($(UNAME_S),Darwin)
$(info "Building on Darwin")
default: cli-darwin-amd64
else
$(error "The OS is not supported")
endif
endif

.PHONY: cli-linux-amd64
cli-linux-amd64:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -tags 'netgo' -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${ARENA_CLI_NAME} cmd/arena/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${JOB_MONITOR} cmd/job-monitor/*.go

.PHONY: cli-darwin-amd64
cli-darwin-amd64:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin GO111MODULE=off go build -tags 'netgo' -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${ARENA_CLI_NAME} ./cmd/arena/*.go

.PHONY: cli-windows
cli-windows:
	mkdir -p bin
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows GO111MODULE=off go build -tags 'netgo' -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${ARENA_CLI_NAME} ./cmd/arena/*.go


.PHONY: install-image
install-image:
	docker build -t cheyang/arena:${VERSION}-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT} -f Dockerfile.install .

.PHONY: notebook-image-kubeflow
notebook-image-kubeflow:
	docker build --build-arg "BASE_IMAGE=${BASE_IMAGE}" -t cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-gpu -f Dockerfile.notebook.gpu .
	docker tag cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-gpu cheyang/arena-notebook:kubeflow

.PHONY: notebook-image
notebook-image:
	docker build --build-arg "BASE_IMAGE=tensorflow/tensorflow:1.12.0-devel-py3" -t cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-cpu -f Dockerfile.notebook.cpu .
	docker tag cheyang/arena:${VERSION}-notebook-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT}-cpu cheyang/arena-notebook:cpu

# make OS_ARCH=darwin-amd64 build-pkg for mac
.PHONY: build-pkg
build-pkg:
	docker rm -f arena-pkg || true
	docker build --build-arg "KUBE_VERSION=v1.11.2" \
				 --build-arg "HELM_VERSION=v2.14.1" \
				 --build-arg "COMMIT=${GIT_SHORT_COMMIT}" \
				 --build-arg "VERSION=${VERSION}" \
				 --build-arg "OS_ARCH=${OS_ARCH}" \
				 --build-arg "GOLANG_VERSION=1.14" \
				 --build-arg "TARGET=cli-${OS_ARCH}" \
	-t arena-build:${VERSION}-${GIT_SHORT_COMMIT}-${OS_ARCH} -f Dockerfile.build .
	docker run -itd --name=arena-pkg arena-build:${VERSION}-${GIT_SHORT_COMMIT}-${OS_ARCH} /bin/bash
	docker cp arena-pkg:/arena-installer-${VERSION}-${GIT_SHORT_COMMIT}-${OS_ARCH}.tar.gz .
	docker rm -f arena-pkg
