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
	log "Failed to run 'helm list', please check if tiller is installed appropriately."
	exit 1
fi

set +e

if ! kubectl get serviceaccount --all-namespaces | grep jobmon; then
	kubectl apply -f /root/kubernetes-artifacts/jobmon/jobmon-role.yaml
fi

if ! kubectl get serviceaccount --all-namespaces | grep tf-job-operator; then
	kubectl apply -f /root/kubernetes-artifacts/tf-operator/tf-operator.yaml
fi
if ! kubectl get serviceaccount --all-namespaces | grep mpi-operator; then
	kubectl apply -f /root/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
fi
set -e

if [ $# -eq 0 ]; then
   cp /usr/local/bin/arena /host/usr/local/bin/arena
   cp -r /charts /host
else
   bash -c "$*"
fi
