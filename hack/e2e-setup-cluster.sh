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
TRAINING_OPERATOR_VERSION="${TRAINING_OPERATOR_VERSION:-v1.9.3}"
K8S_VERSION="${K8S_VERSION:-1.32.3}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-arena-v2}"
NAMESPACE="kubeflow"
TIMEOUT="5m"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. Create kind cluster
echo "Creating Kind cluster '${KIND_CLUSTER_NAME}' with Kubernetes v${K8S_VERSION}..."
kind create cluster \
  --name "${KIND_CLUSTER_NAME}" \
  --image "kindest/node:v${K8S_VERSION}" \
  --config "${SCRIPT_DIR}/kind-config.yaml"

# 2. Install training-operator
echo "Installing Kubeflow Training Operator ${TRAINING_OPERATOR_VERSION}..."
kubectl apply --server-side -k \
  "github.com/kubeflow/training-operator.git/manifests/overlays/standalone?ref=${TRAINING_OPERATOR_VERSION}"

# 3. Wait for training-operator deployment to be ready
echo "Waiting for training-operator deployment to be ready..."
kubectl wait deploy/training-operator \
  -n "${NAMESPACE}" \
  --for=condition=available \
  --timeout "${TIMEOUT}"

# 4. Print cluster info
echo ""
echo "=== Cluster Info ==="
kubectl get nodes
echo ""
kubectl get pods -n "${NAMESPACE}"
echo ""
echo "Cluster '${KIND_CLUSTER_NAME}' is ready for e2e tests."
