#!/usr/bin/env bash

# Copyright 2026 The Kubeflow Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script creates a Kind cluster and installs the Kubeflow Training Operator
# for Arena v2 e2e tests.

set -o errexit
set -o nounset
set -o pipefail

# Variables (overridable via environment)
TRAINER_REPO="${TRAINER_REPO:-https://github.com/kubeflow/training-operator.git}"
TRAINER_REF="${TRAINER_REF:-v1.9.3}"
INSTALL_METHOD="${INSTALL_METHOD:-kustomize}"
K8S_VERSION="${K8S_VERSION:-1.32.3}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-arena-v2}"
NAMESPACE="${NAMESPACE:-kubeflow}"
TIMEOUT="5m"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Clone trainer repo to temp directory
TRAINER_DIR="$(mktemp -d /tmp/trainer-XXXXXX)"
trap 'rm -rf "${TRAINER_DIR}"' EXIT

echo "Cloning ${TRAINER_REPO} (ref: ${TRAINER_REF})..."
git clone --depth 1 --branch "${TRAINER_REF}" "${TRAINER_REPO}" "${TRAINER_DIR}"

# 1. Create kind cluster
echo "Creating Kind cluster '${KIND_CLUSTER_NAME}' with Kubernetes v${K8S_VERSION}..."
kind create cluster \
  --name "${KIND_CLUSTER_NAME}" \
  --image "kindest/node:v${K8S_VERSION}" \
  --config "${SCRIPT_DIR}/kind-config.yaml"

# 2. Install training-operator
case "${INSTALL_METHOD}" in
  kustomize)
    echo "Installing training-operator via kustomize..."
    kubectl apply --server-side -k "${TRAINER_DIR}/manifests/overlays/standalone"
    ;;
  helm)
    echo "Error: INSTALL_METHOD=helm is not currently supported."
    echo "The training-operator v1 chart is not available at ref ${TRAINER_REF}."
    echo "Use INSTALL_METHOD=kustomize (the default) instead."
    exit 1
    ;;
  *)
    echo "Error: INSTALL_METHOD must be 'kustomize' or 'helm', got '${INSTALL_METHOD}'"
    exit 1
    ;;
esac

# 3. Wait for training-operator deployment to be ready
echo "Waiting for training-operator deployment to be ready..."
kubectl wait deploy/training-operator \
  -n "${NAMESPACE}" \
  --for=condition=available \
  --timeout "${TIMEOUT}"

# 4. Pre-load busybox image to avoid docker.io rate limits
BUSYBOX_IMAGE="${E2E_BUSYBOX_IMAGE:-docker.io/library/busybox:1.35}"
echo "Pre-loading ${BUSYBOX_IMAGE} into kind cluster..."
docker pull "${BUSYBOX_IMAGE}"
kind load docker-image "${BUSYBOX_IMAGE}" --name "${KIND_CLUSTER_NAME}"

# 5. Print cluster info
echo ""
echo "=== Cluster Info ==="
kubectl get nodes
echo ""
kubectl get pods -n "${NAMESPACE}"
echo ""
echo "Cluster '${KIND_CLUSTER_NAME}' is ready for e2e tests."
