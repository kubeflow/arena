#!/usr/bin/env bash
set -x -e

function log() {
        echo $(date +"[%Y%m%d %H:%M:%S]: ") $1
}

if ! [ -f $KUBECONFIG ]; then
	log "Failed to find $KUBECONFIG. Please mount kubeconfig file into the pod and make sure it's $KUBECONFIG"
	exit 1
fi

if ! helm list >/dev/null 2>&1; then
	log "Warning: Failed to run 'helm list', please check if tiller is installed appropriately."
fi

set +e

if [[ ! -z "${registry}" ]]; then
  find /charts/ -name *.yaml | xargs sed -i "s/registry.cn-zhangjiakou.aliyuncs.com/${registry}/g"
  find /charts/ -name *.yaml | xargs sed -i "s/registry.cn-hangzhou.aliyuncs.com/${registry}/g"
  find /root/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/registry.cn-zhangjiakou.aliyuncs.com/${registry}/g"
  find /root/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/registry.cn-hangzhou.aliyuncs.com/${registry}/g"
fi

if [[ ! -z "${namespace}" ]]; then
  find /root/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/arena-system/${namespace}/g"
fi

if [[ ! -z "${repo_namespace}" ]]; then
  find /charts/ -name *.yaml | xargs sed -i "s/tensorflow-samples/${repo_namespace}/g"
  find /root/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/tensorflow-samples/${repo_namespace}/g"
fi

if [ "$useLoadBlancer" == "true" ]; then
	find /charts/ -name *.yaml | xargs sed -i "s/NodePort/LoadBalancer/g"
	find /root/kubernetes-artifacts/ -name *.yaml | xargs sed -i "s/NodePort/LoadBalancer/g"
fi


if ! kubectl get serviceaccount --all-namespaces | grep jobmon; then
	kubectl apply -f /root/kubernetes-artifacts/jobmon/jobmon-role.yaml
fi

if ! kubectl get serviceaccount --all-namespaces | grep tf-job-operator; then
	kubectl apply -f /root/kubernetes-artifacts/tf-operator/tf-operator.yaml
fi
if ! kubectl get serviceaccount --all-namespaces | grep mpi-operator; then
	kubectl apply -f /root/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
fi

if [ "$usePrometheus" == "true" ]; then
    if [ "$platform" == "ack" ]; then
        sed -i 's|accelerator/nvidia_gpu|aliyun.accelerator/nvidia_count|g' /root/kubernetes-artifacts/prometheus/gpu-exporter.yaml
    fi
	if ! kubectl get serviceaccount --all-namespaces | grep prometheus; then
	   kubectl apply -f /root/kubernetes-artifacts/prometheus/gpu-exporter.yaml
	   kubectl apply -f /root/kubernetes-artifacts/prometheus/prometheus.yaml
    fi
fi
set -e

if [ "$useHostNetwork" == "true" ]; then
	find /charts/ -name values.yaml | xargs sed -i "/useHostNetwork/s/false/true/g"
fi


if [ -d "/host" ]; then
   cp /usr/local/bin/arena /host/usr/local/bin/arena
   if [ -d "/host/charts" ]; then
      mv /host/charts /host/charts_$(date "+%Y%m%d%H%M%S")
   fi
   cp -r /charts /host
fi

tail -f /dev/null

