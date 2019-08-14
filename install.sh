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

set -x -e

SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"

function log() {
    echo $(date +"[%Y%m%d %H:%M:%S]: ") $1
}

if ! [ -f $KUBECONFIG ]; then
    log "Failed to find $KUBECONFIG. Please mount kubeconfig file into the pod and make sure it's $KUBECONFIG"
    exit 1
fi

if ! which kuebctl; then
	cp $SCRIPT_DIR/bin/kubectl /usr/local/bin/kubectl
fi

set +e

if [[ ! -z "${registry}" ]]; then
    find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i "s/registry.cn-zhangjiakou.aliyuncs.com/${registry}/g"
    find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i "s/registry.cn-hangzhou.aliyuncs.com/${registry}/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/registry.cn-zhangjiakou.aliyuncs.com/${registry}/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/registry.cn-hangzhou.aliyuncs.com/${registry}/g"
fi

if [[ ! -z "${namespace}" ]]; then
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/arena-system/${namespace}/g"
fi

if [[ ! -z "${repo_namespace}" ]]; then
    find /charts/ -name *.yaml | xargs sed -i "s/tensorflow-samples/${repo_namespace}/g"
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/tensorflow-samples/${repo_namespace}/g"
fi

if [ "$useLoadBlancer" == "true" ]; then
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

if [ "$usePrometheus" == "true" ]; then
    if [ "$platform" == "ack" ]; then
        sed -i 's|accelerator/nvidia_gpu|aliyun.accelerator/nvidia_count|g' $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
    fi
    if ! kubectl get serviceaccount --all-namespaces | grep prometheus; then
     kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
     kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/prometheus.yaml
     kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/prometheus/grafana.yaml
    fi
fi
set -e

if [ "$useHostNetwork" == "true" ]; then
    find /charts/ -name values.yaml | xargs sed -i "/useHostNetwork/s/false/true/g"
fi


now=$(date "+%Y%m%d%H%M%S")
if [ -f "/usr/local/bin/arena" ]; then
    mv $SCRIPT_DIR/bin/arena /usr/local/bin/arena-$now
fi
cp $SCRIPT_DIR/bin/arena /usr/local/bin/arena

if ! which helm; then
	cp $SCRIPT_DIR/bin/helm /usr/local/bin/helm
fi

if [ -d "/charts" ]; then
    mv /charts /charts-$now
fi
cp -r /charts /
