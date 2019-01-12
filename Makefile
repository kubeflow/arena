PACKAGE=github.com/kubeflow/arena
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/bin
ARENA_CLI_NAME=arena
JOB_MONITOR=jobmon

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_SHORT_COMMIT=$(shell git rev-parse --short HEAD)
DOCKER_BUILD_DATE=$(shell date -u +'%Y%m%d%H%M%S')
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
PACKR_CMD=$(shell if [ "`which packr`" ]; then echo "packr"; else echo "go run vendor/github.com/gobuffalo/packr/packr/main.go"; fi)

BUILDER_IMAGE=arena-builder
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
.PHONY: all
all: cli 

.PHONY: cli
cli:
	mkdir -p bin
	CGO_ENABLED=1 go build -tags 'netgo'  -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${ARENA_CLI_NAME} cmd/arena/*.go
	CGO_ENABLED=1 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${JOB_MONITOR} cmd/job-monitor/*.go

.PHONY: install-image
install-image:
	docker build -t cheyang/arena:${VERSION}-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT} -f Dockerfile.install .

.PHONY: notebook-image
notebook-image:
  docker build -t cheyang/arena:${VERSION}-${DOCKER_BUILD_DATE}-${GIT_SHORT_COMMIT} -f Dockerfile.notebook .