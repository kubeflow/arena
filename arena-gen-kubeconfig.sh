#!/bin/bash

# Copyright 2024 The Kubeflow Authors All rights reserved.
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

set -e
TEMPDIR=$( mktemp -d )

trap "{ rm -rf $TEMPDIR ; }" EXIT

function help() {
    echo -e "
Usage:

    arena-gen-kubeconfig.sh [OPTION1] [OPTION2] ...

Options:
    --user-name <USER_NAME>                    Specify the user name
    --user-namespace <USER_NAMESPACE>          Specify the user namespace
    --user-config <USER_CONFIG>                Specify the user config,refer the ~/charts/user/values.yaml or /charts/user/values.yaml
    --force                                    If the user has been existed,force to update the user
    --delete                                   Delete the user
    --output <KUBECONFIG|USER_MANIFEST_YAML>   Specify the output kubeconfig file or the user manifest yaml
    --admin-kubeconfig <ADMIN_KUBECONFIG>      Specify the Admin kubeconfig file
    --cluster-url <CLUSTER_URL>                Specify the Cluster URL,if not specified,the script will detect the cluster url
    --create-user-yaml                         Only generate the user manifest yaml,don't apply it and create kubeconfig file

"

}

function run() {
    # check user is null or not
    if [[ $USER_NAME == "" ]];then
        logger error "user name is null,please use --user-name to set"
        exit 1 
    fi
    # check user namespace is not null
    if [[ $USER_NAMESPACE == "" ]];then
        logger debug "user namespace not set,please user --user-namespace to set"
        exit 1 
    fi
    check_output
    if [[ $DELETE_USER == "1" ]];then
        delete_user
        return 
    fi  
    if [[ $ONLY_GEN_USER_YAML == "1" ]];then
        get_chart_dir
        generate_manifest_yaml $OUTPUT
        logger debug "the user manifest yaml has been stored to $OUTPUT"
        return 
    fi 
    check_namespace
    check_user_is_existed
    if [[ $USER_EXISTED == "false" ]]  || [[ $FORCE == "1" ]];then
        get_chart_dir
        apply_manifest_yaml
    fi
    get_cluster_url
    generate_kubeconfig
}

function check_namespace() {
    if arena-kubectl get ns | grep "^${USER_NAMESPACE} " &> /dev/null;then
        logger debug "namespace ${USER_NAMESPACE} has been existed,skip to create it"
        return
    fi
    arena-kubectl create ns ${USER_NAMESPACE}
}

function check_admin_kubeconfig() {
    if [[ $ADMIN_KUBECONFIG != "" ]];then
        export KUBECONFIG=$ADMIN_KUBECONFIG
    fi
    if ! arena-kubectl cluster-info &> /dev/null;then
        logger error "failed to execute 'arena-kubectl cluster-info',please make sure the role of kubeconfig file is admin"
        exit 1
    fi
}

function check_user_is_existed() {
    export USER_EXISTED="false" 
    if arena-kubectl get ClusterRole "arena:$USER_NAMESPACE:$USER_NAME" &> /dev/null;then
        export USER_EXISTED="true"
    fi
}

function check_output() {
    if [[ $OUTPUT == "" ]];then
        if [[ $ONLY_GEN_USER_YAML == "1" ]];then
            export OUTPUT="./${USER_NAME}.yaml"
        else
            export OUTPUT="./${USER_NAME}.kubeconfig"
        fi  
    else
        output_path=$(dirname $OUTPUT)
        if [[ ! -d $output_path ]];then
            logger error "the output path $output_path is invalid"
            exit 3
        fi  
    fi   
}

function generate_manifest_yaml() {
    if [[ $USER_VALUES == "" ]];then
        export USER_VALUES=$CHART_DIR/user/values.yaml  
        logger debug "the user configuration not set,use the default configuration file"
    fi
    if [ ! -f $USER_VALUES ];then
        logger error "the file $USER_VALUES doesn't exist"
        exit 2
    fi    
    output=$1
    arena-helm template --name $USER_NAME --namespace $USER_NAMESPACE -f $USER_VALUES $CHART_DIR/user > $output
}

function apply_manifest_yaml() {
    namespace=$USER_NAMESPACE
    user=$USER_NAME
    yaml_file=$TEMPDIR/${user}.yaml
    generate_manifest_yaml $yaml_file
    arena-kubectl apply -f $yaml_file
    if arena-kubectl get configmap arena-user-${user} -n $namespace &> /dev/null;then
        logger debug "delete old configmap of user $user"
        arena-kubectl delete configmap arena-user-${user} -n $namespace
    fi
    arena-kubectl create configmap arena-user-${user}  -n $namespace --from-file=config=$yaml_file 
}

function delete_user() {
    namespace=$USER_NAMESPACE
    user=$USER_NAME
    if ! arena-kubectl get clusterrole arena:$namespace:$user &> /dev/null;then
        logger debug "not found the user $user,skip to delete it"
        exit 0
    fi  
    if arena-kubectl get configmap -n $namespace arena-user-${user} &> /dev/null;then
        arena-kubectl get configmap -n $namespace arena-user-${user} -o jsonpath='{.data.config}' > $TEMPDIR/${user}.yaml
        arena-kubectl delete -f $TEMPDIR/${user}.yaml
        return  
    fi 
    arena-kubectl delete ClusterRoleBinding arena:$namespace:$user
    arena-kubectl delete clusterrole arena:$namespace:$user 
    arena-kubectl delete RoleBinding  arena:$user -n $namespace
    arena-kubectl delete Role arena:$user -n $namespace  
    arena-kubectl delete ServiceAccount $user -n $namespace
    if arena-kubectl get quota  arena-quota-$user -n $namespace &> /dev/null;then
        arena-kubectl delete quota  arena-quota-$user -n $namespace
    fi  
}

function get_chart_dir() {
    export CHART_DIR=~/charts
    if [ -d $CHART_DIR ];then
        logger debug "found arena charts in $CHART_DIR"
        return 
    fi 
    export CHART_DIR=/charts
    if [ -d $CHART_DIR ];then
        logger debug "found arena charts in $CHART_DIR"
        return 
    fi
    logger error "failed to find arena charts in '~/charts' and '/charts'"
    exit 2
}

function get_cluster_url() {
    if [[ $CLUSTER_URL != "" ]];then
        return 
    fi
    export CLUSTER_URL=$( arena-kubectl config view -o jsonpath='{.clusters[0].cluster.server}')
    if [[ $CLUSTER_URL != "" ]];then
        return 
    fi
    logger warning "failed to get cluster url from the admin kubeconfig file"
    export CLUSTER_URL=$(arena-kubectl get endpoints | grep -E '\<kubernetes\>' | awk '{print $2}')
}

function generate_kubeconfig() {
    user=$USER_NAME
    namespace=$USER_NAMESPACE
    SA_SECRET=$( arena-kubectl get sa -n $namespace $user -o jsonpath='{.secrets[0].name}' )
    BEARER_TOKEN=$( arena-kubectl get secrets -n $namespace $SA_SECRET -o jsonpath='{.data.token}' | base64 --decode)
    arena-kubectl get secrets -n $namespace $SA_SECRET -o jsonpath='{.data.ca\.crt}' | base64 --decode > $TEMPDIR/ca.crt
    arena-kubectl config \
    --kubeconfig=$OUTPUT \
    set-cluster \
    $CLUSTER_URL \
    --server=$CLUSTER_URL \
    --certificate-authority=$TEMPDIR/ca.crt \
    --embed-certs=true

    arena-kubectl config  \
    --kubeconfig=$OUTPUT \
    set-credentials $user --token=$BEARER_TOKEN

    arena-kubectl config \
    --kubeconfig=$OUTPUT \
    set-context $USER_NAME \
    --namespace="$USER_NAMESPACE" \
    --cluster=$CLUSTER_URL \
    --user=$user

    arena-kubectl config \
    --kubeconfig=$OUTPUT \
    use-context $USER_NAME
    
    rm -rf $TEMPDIR
    logger debug "kubeconfig written to file $OUTPUT"
}



function parse_args() {
    while
		[[ $# -gt 0 ]]
	do
		key="$1"

		case $key in
		--admin-kubeconfig)
            check_option_value "--admin-kubeconfig" $2
            export ADMIN_KUBECONFIG=$2
            shift
			;;
        --output)
            check_option_value "--output" $2
            export OUTPUT=$2
            shift
            ;;
        --create-user-yaml)
            export ONLY_GEN_USER_YAML=1
            ;;
        --user-config)
            check_option_value "--user-config" $2
            export USER_VALUES=$2
            shift
            ;;
        --user-name)
            check_option_value "--user-name" $2
            export USER_NAME=$2
            shift
            ;;
        --user-namespace)
            check_option_value "--user-namespace" $2
            export USER_NAMESPACE=$2
            shift
            ;;
        --cluster-url)
            check_option_value "--cluster-url" $2
            export CLUSTER_URL=$2
            shift
            ;;
        --delete)
            export DELETE_USER=1
            ;;
        --force)
            export FORCE=1
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

function logger() {
    timestr=$(date +"%Y-%m-%d/%H:%M:%S")
    level=$(echo $1 | tr 'a-z' 'A-Z')
    echo ${timestr}"  "${level}"  "$2
}

function main() {
    if [[ "$@" == "" ]];then
        help
        exit 1
    fi
    parse_args "$@"
    check_admin_kubeconfig
    run
}

main "$@"
