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
HELM_OPTIONS=""
CRD_DIRS=""

function help() {
    echo -e "
Usage:

    install.sh [OPTION1] [OPTION2] ...

Options:
    --kubeconfig string              Specify the kubeconfig file
    --namespace string               Specify the namespace that operators will be installed in
    --only-binary                    Only install arena binary
    --region-id string               Specify the region id(it is available in Alibaba Cloud)
    --host-network                   Enable host network
    --docker-registry string         Specify the docker registry
    --registry-repo-namespace string Specify the docker registry repo namespace
    --loadbalancer                   Specify k8s service type with loadbalancer
    --prometheus                     Install prometheus
    --platform string                Specify the platform(eg: ack)
    --rdma                           Enable rdma feature
    --deploy-with-helm               Install operator with helm
    --install-binary-on-master       Install binary on master
    --chart-value-file               Specify the chart value file
    --update-existed-artifacts       Update the existed artifacts or not
"

}


function logger() {
    timestr=$(date +"%Y-%m-%d/%H:%M:%S")
    level=$(echo $1 | tr 'a-z' 'A-Z')
    echo ${timestr}"  "${level}"  "$2
}

function set_sed_option() {
    if [[ $(uname -s) == "Darwin" ]];then
        export SED_OPTION=".sed_backup"
    else
        export SED_OPTION=""
    fi
}

# if pull images by aliyun vpc,change the image
function support_image_regionalization(){
    for file in $(ls $1);do
        local path=$1"/"$file
        if [ -d $path ];then
            support_image_regionalization $path
        else
            sed -i $SED_OPTION "s@registry\..*aliyuncs.com@${REGION}@g" $path
        fi
    done
}

# if execute the install.sh needs sudo,add sudo command to all command
function set_sudo() {
    export sudo_prefix=""
    if [ `id -u` -ne 0 ]; then
        export sudo_prefix="sudo"
    fi
}

# install kubectl and rename arena-kubectl
# install helm and rename arena-helm
function install_kubectl_and_helm() {
    logger "debug" "start to install arena-kubectl and arena-helm"
    ${sudo_prefix} rm -rf /usr/local/bin/arena-kubectl
    ${sudo_prefix} rm -rf /usr/local/bin/arena-helm

    ${sudo_prefix} cp $SCRIPT_DIR/bin/kubectl /usr/local/bin/arena-kubectl
    ${sudo_prefix} cp $SCRIPT_DIR/bin/helm /usr/local/bin/arena-helm

    if [[ $ONLY_BINARY != "true" ]];then 
        if ! ${sudo_prefix} arena-kubectl cluster-info >/dev/null 2>&1; then
            logger "error" "failed to execute 'arena-kubectl cluster-info'"
            logger "error" "Please setup kubeconfig correctly before installing arena"
            exit 1
        fi
    fi
    logger "debug" "succeed to install arena-kubectl and arena-helm"
}

function custom_charts() {
    if [[ $REGION != "" ]];then
        logger "debug" "enable image regionalization"
	    support_image_regionalization $SCRIPT_DIR/charts
    fi

    if [ "$USE_HOSTNETWORK" == "true" ]; then
        logger "debug" "enable host network mode"
        find $SCRIPT_DIR/charts/ -name values.yaml | xargs sed -i $SED_OPTION "/useHostNetwork/s/false/true/g"
    fi

    if [[ ${DOCKER_REGISTRY} != "" ]]; then
        logger "debug" "custom the docker registry with ${DOCKER_REGISTRY}"
        find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i $SED_OPTION "s/registry.cn-zhangjiakou.aliyuncs.com/${DOCKER_REGISTRY}/g"
        find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i $SED_OPTION "s/registry.cn-hangzhou.aliyuncs.com/${DOCKER_REGISTRY}/g"
    fi

    if [[ ${REGISTRY_REPO_NAMESPACE} != "" ]]; then
        logger "debug" "custom the docker registry repo namespace with ${REGISTRY_REPO_NAMESPACE}"
        find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i $SED_OPTION "s/tensorflow-samples/${REGISTRY_REPO_NAMESPACE}/g"
    fi

    if [ "$USE_LOADBALANCER" == "true" ]; then
        logger "debug" "specify service with loadbalancer type"
        find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i $SED_OPTION "s/NodePort/LoadBalancer/g"
    fi

    if [[ $USE_RDMA == "true" ]];then
        find $SCRIPT_DIR/charts/ -name *.yaml | xargs sed -i $SED_OPTION "/enableRDMA/s/false/true/g"
    fi
    # this command is used to delete files generated by sed command, please make sure this command is last one in the function
    find $SCRIPT_DIR/charts/ -name *.sed_backup | xargs rm -rf
}

function custom_manifests() {
    if [[ $REGION != "" ]];then
        support_image_regionalization $SCRIPT_DIR/kubernetes-artifacts
    fi

    if [[ ${DOCKER_REGISTRY} != "" ]]; then
        find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i $SED_OPTION "s/registry.cn-zhangjiakou.aliyuncs.com/${DOCKER_REGISTRY}/g"
        find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i $SED_OPTION "s/registry.cn-hangzhou.aliyuncs.com/${DOCKER_REGISTRY}/g"
    fi

    if [[ "${NAMESPACE}" != "" ]]; then
        logger "debug" "custom the namespace(${NAMESPACE}) which operators will be installed in"
        find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i $SED_OPTION "s/arena-system/${NAMESPACE}/g"
    fi

    if [[ ${REGISTRY_REPO_NAMESPACE} != "" ]]; then
        find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i $SED_OPTION "s/tensorflow-samples/${REGISTRY_REPO_NAMESPACE}/g"
    fi

    if [ "$USE_LOADBALANCER" == "true" ]; then
        find $SCRIPT_DIR/kubernetes-artifacts/ -name *.yaml | xargs sed -i $SED_OPTION "s/NodePort/LoadBalancer/g"
    fi

    if [ "$PLATFORM" == "ack" ]; then
        sed -i $SED_OPTION 's|accelerator/nvidia_gpu|aliyun.accelerator/nvidia_count|g' $SCRIPT_DIR/kubernetes-artifacts/prometheus/gpu-exporter.yaml
    fi

    # this command is used to delete files generated by sed command, please make sure this command is last one in the function
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.sed_backup | xargs rm -rf
}

# install arena command tool and charts
function install_arena_and_charts() {
    now=$(date "+%Y%m%d%H%M%S")
    find $SCRIPT_DIR/kubernetes-artifacts/ -name *.sed_backup | xargs rm -rf
    if [ -f "/usr/local/bin/arena" ]; then
        ${sudo_prefix} cp /usr/local/bin/arena /usr/local/bin/arena-$now
    fi
    ${sudo_prefix} cp $SCRIPT_DIR/bin/arena /usr/local/bin/arena
    if [ -f /usr/local/bin/arena-uninstall ];then
        ${sudo_prefix} rm -rf /usr/local/bin/arena-uninstall
    fi
    ${sudo_prefix} cp $SCRIPT_DIR/bin/arena-uninstall /usr/local/bin
    if [ ! -d $SCRIPT_DIR/charts/kubernetes-artifacts ];then
        cp -a $SCRIPT_DIR/kubernetes-artifacts $SCRIPT_DIR/charts
    fi
    if [ ! -d $SCRIPT_DIR/charts/arena-artifacts ];then
        cp -a $SCRIPT_DIR/arena-artifacts $SCRIPT_DIR/charts
    fi
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
}

function install_arena_gen_kubeconfig() {
    ${sudo_prefix} rm -rf /usr/local/bin/arena-gen-kubeconfig.sh
    ${sudo_prefix} cp $SCRIPT_DIR/bin/arena-gen-kubeconfig.sh /usr/local/bin/arena-gen-kubeconfig.sh
}

function check_addons() {
    addon_name=$1
    check_kubedl=$2
    sa=$3
    option=$4
    if $check_kubedl;then
        # if KubeDL has installed, will skip install some CRDs of framework operator
        if arena-kubectl get serviceaccount --all-namespaces | grep kubedl &> /dev/null; then
            logger "warning" "KubeDL has been detected, will skip install $addon_name"
            export HELM_OPTIONS="$HELM_OPTIONS --set $option"
            return
        fi
    fi
    # if service account has been existed and it is managed by helm chart
    if (arena-kubectl get serviceaccount --all-namespaces -l helm.sh/chart=arena-artifacts | grep $sa  2>1) &> /dev/null; then
        logger debug "service account $sa has been existed,and it is managed by helm,skip to reinstall $addon_name"
        return 
    fi    
    # if service account has been existed and it is not managed by helm chart
    if arena-kubectl get serviceaccount --all-namespaces | grep $sa &> /dev/null; then
        logger debug "service account $sa has been existed,and it is not managed by helm,skip to reinstall $addon_name"
        export HELM_OPTIONS="$HELM_OPTIONS --set $option"
        return 
    fi    
}


function apply_tf() {
    check_addons tf-operator true tf-job-operator tf.enabled=false
}

# TODO: the pytorch-operator update
function apply_pytorch() {
     check_addons pytorch-operator true pytorch-operator pytorch.enabled=false
}

function apply_mpi() {
    check_addons mpi-operator true mpi-operator mpi.enabled=false
}

function apply_et() {
    check_addons et-operator true et-operator et.enabled=false
}

function clean_old_env() {
    set +e 
    # update the crd version
    if (arena-kubectl get crd tfjobs.kubeflow.org -oyaml |grep -i 'version: v1alpha2' 2>1) &> /dev/null; then
        arena-kubectl delete crd tfjobs.kubeflow.org    
        arena-kubectl create -f ${SCRIPT_DIR}/arena-artifacts/charts/tf-operator/crds
    fi
    # remove the old kubedl-operator
    if arena-kubectl get deployment,Service,ServiceAccount -n $NAMESPACE | grep kubedl-operator &> /dev/null;then 
        arena-kubectl delete deployment kubedl-operator -n $NAMESPACE
        arena-kubectl delete ServiceAccount kubedl-operator -n $NAMESPACE
        arena-kubectl delete crd crons.apps.kubedl.io
        arena-kubectl delete ClusterRole kubedl-operator-role -n $NAMESPACE
        arena-kubectl delete ClusterRoleBinding kubedl-operator-rolebinding -n $NAMESPACE
        arena-kubectl delete Service kubedl-operator -n $NAMESPACE
        arena-kubectl delete ServiceMonitor kubedl-operator -n $NAMESPACE
    fi
    set -e    
}
function apply_cron() {
    check_addons cron-operator true cron-operator cron.enabled=false
}

function apply_tf_dashboard() {
    check_addons tf-dashboard false tf-job-dashboard tfdashboard.enabled=false
}

function apply_rdma() {
    if [[ $USE_RDMA != "true" ]];then
        return 
    fi  
    # if service account has been existed and it is managed by helm chart
    if (arena-kubectl get ds --all-namespaces -l helm.sh/chart=arena-artifacts | grep rdma-sriov-dp-ds 2>1) &> /dev/null; then
        export HELM_OPTIONS="$HELM_OPTIONS --set rdma.enabled=true" 
        logger debug "daemonset rdma-sriov-dp-ds has been existed,and it is managed by helm,skip to reinstall it"
        return 
    fi   
    if arena-kubectl get ds --all-namespaces | grep rdma-sriov-dp-ds &> /dev/null; then
        logger debug "daemonset rdma-sriov-dp-ds has been existed,and it is not managed by helm,skip to reinstall it"
        export HELM_OPTIONS="$HELM_OPTIONS --set rdma.enabled=false"
        return 
    fi
}


function create_namespace() {
    if [[ "${NAMESPACE}" == "" ]]; then
        export NAMESPACE="arena-system"
    fi
    namespace=${NAMESPACE}
    if arena-kubectl get ns | grep -E "\<$namespace\>" &> /dev/null;then
        logger "debug" "namespace $namespace has been existed,skip to create it"
        return
    fi
    arena-kubectl create ns $namespace
}

function install_binary_on_master() {
    if [[ $INSTALL_BINARY_ON_MASTER != "true" ]];then
        return 
    fi
    master_count=$(arena-kubectl get nodes -l node-role.kubernetes.io/master | grep -v NAME  | wc -l)
    master_count=$(echo $master_count)
    export HELM_OPTIONS="$HELM_OPTIONS --set binary.enabled=true --set binary.masterCount=$master_count"
}

function binary() {
    install_kubectl_and_helm
    custom_charts
    install_arena_and_charts
    install_arena_gen_kubeconfig
}

function operators() {
    if [[ $ONLY_BINARY == "true" ]];then
        logger "debug" "skip to install operators,because --only-binary is enabled"
        return
    fi
    create_namespace
    if [[ $DEPLOY_WITH_HELM == "true" ]];then
        if ! helm version &> /dev/null;then
            logger error "not found helm binary,can not install arena by helm"
            exit 2
        fi    
    fi  
    clean_old_env
    apply_tf
    apply_pytorch
    apply_mpi
    apply_et
    apply_rdma
    apply_cron
    apply_tf_dashboard
    install_binary_on_master
    if [[ $DEPLOY_WITH_HELM != "true" ]];then
        deploy_witout_helm
    else
        deploy_with_helm
    fi 

}

function deploy_witout_helm() {
    # if detect service account with label  helm.sh/chart=arena-artifacts,we see helm chart arena-artifacts has been installed
    sas=$(arena-kubectl get serviceaccount --all-namespaces -l helm.sh/chart=arena-artifacts 2>1)
    if [[ $sas != "" ]];then
        logger debug "helm chart arena-artifacts has been installed,skip to deploy operators retry"
        return 
    fi  
    if ! echo $HELM_OPTIONS | grep "tf.enabled=false" &> /dev/null;then
        arena-kubectl apply -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-crd.yaml
        arena-kubectl apply -f ${SCRIPT_DIR}/kubernetes-artifacts/tf-operator/tf-operator.yaml
    fi
    if ! echo $HELM_OPTIONS | grep "mpi.enabled=false" &> /dev/null;then
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/mpi-operator/mpi-operator.yaml
    fi
    if ! echo $HELM_OPTIONS | grep "pytorch.enabled=false" &> /dev/null;then
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/pytorch-operator/pytorch-operator.yaml
    fi
    if ! echo $HELM_OPTIONS | grep "et.enabled=false" &> /dev/null;then
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/et-operator/et-operator.yaml
    fi
    if ! echo $HELM_OPTIONS | grep "cron.enabled=false" &> /dev/null;then
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/cron-operator/cron-crd.yaml
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/cron-operator/cron-operator.yaml
    fi
    if echo $HELM_OPTIONS | grep "rdma.enabled=true" &> /dev/null;then
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/rdma/rdma-config.yaml
        arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/rdma/device-plugin.yaml
    fi
    arena-kubectl apply -f $SCRIPT_DIR/kubernetes-artifacts/arena-configmap/configmap.yaml             
}

function deploy_with_helm() {
    if [[ $CHART_VALUE_FILE != "" ]];then
        export HELM_OPTIONS="$HELM_OPTIONS -f $CHART_VALUE_FILE"
    fi  
    if helm list -n $NAMESPACE | grep "arena-artifacts" &> /dev/null;then
        if [[ $UPDATE_EXISTED_ARTIFACTS != "true" ]];then 
            echo "user not want to update artifacts,skip"
            return 
        fi 
        logger debug "arena-artifacts has been installed,start to upgrade it"
        helm upgrade arena-artifacts -n $NAMESPACE $HELM_OPTIONS $SCRIPT_DIR/arena-artifacts 
        return 
    fi  
    helm install arena-artifacts -n $NAMESPACE $HELM_OPTIONS $SCRIPT_DIR/arena-artifacts 
}


function parse_args() {
    while [[ $# -gt 0 ]];do
        key="$1"
        case $key in
        --only-binary)
            export ONLY_BINARY="true"
        ;;
        --host-network)
            export USE_HOSTNETWORK="true"
        ;;
        --deploy-with-helm)
            export DEPLOY_WITH_HELM="true"
        ;;
        --install-binary-on-master)
            export INSTALL_BINARY_ON_MASTER="true"
        ;;
        --update-existed-artifacts)
            export UPDATE_EXISTED_ARTIFACTS="true"
        ;;
        --rdma)
            export USE_RDMA="true"
        ;;
        --loadbalancer)
            export USE_LOADBALANCER="true"
        ;;
        --prometheus)
            export USE_PROMETHEUS="true"
        ;;
        --kubeconfig)
            check_option_value "--kubeconfig" $2
            export KUBECONFIG=$2
            shift
        ;;
        --platform)
            check_option_value "--platform" $2
            export PLATFORM=$2
            shift
            ;;
        --region-id)
            check_option_value "--region-id" $2
            export REGION=$2
            shift
        ;;
        --docker-registry)
            check_option_value "--docker-registry" $2
            export DOCKER_REGISTRY=$2
            shift
        ;;
        --registry-repo-namespace)
            check_option_value "--registry-repo-namespace" $2
            export REGISTRY_REPO_NAMESPACE=$2
            shift
        ;;
        --namespace)
            check_option_value "--namespace" $2
            export NAMESPACE=$2
            shift
        ;;
        --chart-value-file)
            check_option_value "--chart-value-file" $2
            export CHART_VALUE_FILE=$2
            shift
        ;;
        --help|-h)
            help
            exit 0
        ;;
        *)
            # unknown option
            logger error "unkonw option [$key]"
            help
            exit 3
        ;;
        esac
        shift
    done
}

function check_option_value() {
    option=$1
    value=$2
    if [[ $value == "" ]] || echo "$value" | grep -- "^--" &> /dev/null;then
        logger error "the option $option not set value,please set it"
        exit 3
    fi
}


function main() {
    parse_args "$@"
    set_sed_option
    set_sudo
    binary
    operators
    logger "debug" "--------------------------------"
    logger "debug" "Arena has been installed successfully!"
}

main "$@"
