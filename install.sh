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

sudo_prefix=""
if [ `id -u` -ne 0 ]; then
    sudo_prefix="sudo"
fi

rm -rf /usr/local/bin/arena-kubectl
${sudo_prefix} cp $SCRIPT_DIR/bin/kubectl /usr/local/bin/arena-kubectl

if ! arena-kubectl cluster-info >/dev/null 2>&1; then
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


if ! arena-kubectl get serviceaccount --all-namespaces | grep jobmon; then
    arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/jobmon/jobmon-role.yaml
fi

# if KubeDL has installed, will skip install some CRDs of framework operator
if arena-kubectl get serviceaccount --all-namespaces | grep kubedl; then
    log "KubeDL has been detected, will skip install tf-operator"
else
    if ! arena-kubectl get serviceaccount --all-namespaces | grep tf-job-operator; then
        arena-kubectl apply -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-crd.yaml
        arena-kubectl apply -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-operator.yaml
    else
        if arena-kubectl get crd tfjobs.kubeflow.org -oyaml |grep -i 'version: v1alpha2'; then
            arena-kubectl delete -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-operator-v1alpha2.yaml
            arena-kubectl apply -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-crd.yaml
            arena-kubectl apply -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-operator.yaml
        fi
    fi
fi
if ! arena-kubectl get serviceaccount --all-namespaces | grep mpi-operator; then
    arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
fi

if [ "$USE_PROMETHEUS" == "true" ]; then
    if [ "$PLATFORM" == "ack" ]; then
        sed -i 's|accelerator/nvidia_gpu|aliyun.accelerator/nvidia_count|g' $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
    fi
    if ! arena-kubectl get serviceaccount --all-namespaces | grep prometheus; then
     arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
     arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/prometheus.yaml
     arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/grafana.yaml
    fi
fi
# set -e

if [ "$USE_HOSTNETWORK" == "true" ]; then
    find $SCRIPT_DIR/charts/ -name values.yaml | xargs sed -i "/useHostNetwork/s/false/true/g"
fi

now=$(date "+%Y%m%d%H%M%S")
if [ -f "/usr/local/bin/arena" ]; then
    ${sudo_prefix} cp /usr/local/bin/arena /usr/local/bin/arena-$now
fi
${sudo_prefix} cp $SCRIPT_DIR/bin/arena /usr/local/bin/arena

${sudo_prefix} rm -rf /usr/local/bin/arena-helm

${sudo_prefix} cp $SCRIPT_DIR/bin/helm /usr/local/bin/arena-helm

# For non-root user, put the charts dir to the home directory
if [ `id -u` -eq 0 ];then  
    if [ -d "/charts" ]; then
       mv /charts /charts-$now
    fi
    cp -r $SCRIPT_DIR/charts / 
else  
    if [ -d "~/charts" ]; then
      mv ~/charts ~/charts-$now
    fi
    cp -r $SCRIPT_DIR/charts ~/  
fi

log "--------------------------------"
log "Arena has been installed successfully!"
