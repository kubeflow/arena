#!/usr/bin/env bash

# Copyright 2018 The Kubeflow Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"

function log() {
    echo $(date +"[%Y%m%d %H:%M:%S]: ") $1
}

if ! which kubectl >/dev/null 2>&1; then
	cp $SCRIPT_DIR/bin/kubectl /usr/local/bin/kubectl
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
    log "Please setup kubeconfig correctly before installing arena"
    exit 1
fi

# set +e

if [[ ! -z "${DOCKER_REGISTRY}" ]]; then
    find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i "s/registry.cn-zhangjiakou.aliyuncs.com/${DOCKER_REGISTRY}/g"
    find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i "s/registry.cn-hangzhou.aliyuncs.com/${DOCKER_REGISTRY}/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/registry.cn-zhangjiakou.aliyuncs.com/${DOCKER_REGISTRY}/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/registry.cn-hangzhou.aliyuncs.com/${DOCKER_REGISTRY}/g"
fi

if [[ ! -z "${NAMESPACE}" ]]; then
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/arena-system/${NAMESPACE}/g"
fi

if [[ ! -z "${REGISTRY_REPO_NAMESPACE}" ]]; then
    find /charts/ -name *.yaml | xargs sed -i "s/tensorflow-samples/${REGISTRY_REPO_NAMESPACE}/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/tensorflow-samples/${REGISTRY_REPO_NAMESPACE}/g"
fi

if [ "$USE_LOADBALANCER" == "true" ]; then
    find /charts/ -name *.yaml | xargs sed -i "s/NodePort/LoadBalancer/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/NodePort/LoadBalancer/g"
fi


if ! kubectl get serviceaccount --all-namespaces | grep jobmon; then
    kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/jobmon/jobmon-role.yaml
fi

if ! kubectl get serviceaccount --all-namespaces | grep tf-job-operator; then
    kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/tf-operator/tf-crd.yaml
    kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/tf-operator/tf-operator.yaml
else
    if kubectl get crd tfjobs.kubeflow.org -oyaml |grep -i 'version: v1alpha2'; then
        kubectl delete -f $SCRIPT_DIR/kubernetes-artifacts/tf-operator/tf-operator-v1alpha2.yaml
        kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/tf-operator/tf-crd.yaml
        kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/tf-operator/tf-operator.yaml
    fi
fi
if ! kubectl get serviceaccount --all-namespaces | grep mpi-operator; then
    kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
fi

if [ "$USE_PROMETHEUS" == "true" ]; then
    if [ "$PLATFORM" == "ack" ]; then
        sed -i 's|accelerator/nvidia_gpu|aliyun.accelerator/nvidia_count|g' $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
    fi
    if ! kubectl get serviceaccount --all-namespaces | grep prometheus; then
     kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
     kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/prometheus.yaml
     kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/grafana.yaml
    fi
fi
# set -e

if [ "$USE_HOSTNETWORK" == "true" ]; then
    find /charts/ -name values.yaml | xargs sed -i "/useHostNetwork/s/false/true/g"
fi

now=$(date "+%Y%m%d%H%M%S")
if [ -f "/usr/local/bin/arena" ]; then
    cp /usr/local/bin/arena /usr/local/bin/arena-$now
fi
cp $SCRIPT_DIR/bin/arena /usr/local/bin/arena

if ! which arena-helm; then
	cp $SCRIPT_DIR/bin/helm /usr/local/bin/arena-helm
fi

if [ -d "/charts" ]; then
    mv /charts /charts-$now
fi
cp -r $SCRIPT_DIR/charts /

log "--------------------------------"
log "Arena has been installed successfully!"
