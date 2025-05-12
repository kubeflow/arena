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

set -xe

function help() {
    echo -e "
Usage:

    arena-uninstall [OPTION1] [OPTION2] ...

Options:
	--kubeconfig string              Specify the kubeconfig file
	--namespace  string              Specify the namespace to delete arena
	--delete-binary                  Clean the client env,include ~/charts and /usr/local/bin/arena
	--delete-crds	                 Delete the CRDs,Warning: this option will delete the training jobs
	--chart-dir                      Specify the chart dir
"

}

function logger() {
    timestr=$(date +"%Y-%m-%d/%H:%M:%S")
    level=$(echo $1 | tr 'a-z' 'A-Z')
    echo ${timestr}"  "${level}"  "$2
}

function run() {
	detect_chart_dir
	delete 
	if [[ $DELETE_CRDS == "true" ]];then
		delete_crds $ARTIFACTS_DIR/all_crds
	fi
	if [[ $CLEAN_CLIENT == "true" ]];then
		delete_client
	fi    
}

function delete() {
	set +e 
	if arena-helm list -n $ARENA_NAMESPACE | grep arena-artifacts &> /dev/null;then
		arena-helm delete arena-artifacts -n $ARENA_NAMESPACE
	fi
	arena-helm template arena-artifacts -n $ARENA_NAMESPACE $ARTIFACTS_DIR > /tmp/arena-artifacts.yaml
	arena-kubectl delete -f /tmp/arena-artifacts.yaml
	arena-kubectl delete ns $ARENA_NAMESPACE
	set -e  
}

function delete_client() {
	rm -rf /charts
	rm -rf /usr/local/bin/arena*
}

function delete_crds() {
    for file in $(ls $1);do
        local path=$1"/"$file
        if [ -d $path ];then
            delete_crds $path
        else
			arena-kubectl delete -f $path || true 
        fi
    done
}

function detect_chart_dir() {
	ARTIFACTS_DIR=""
	if [[ $CHART_DIR !=  "" ]];then
		export ARTIFACTS_DIR=$CHART_DIR
		return 
	fi 
	if [ -d arena-artifacts ];then
		export ARTIFACTS_DIR=arena-artifacts
		return 
	fi
	if [ -d ~/charts/arena-artifacts ];then
		export ARTIFACTS_DIR=~/charts/arena-artifacts
		return 
	fi
	export ARTIFACTS_DIR=/charts/arena-artifacts    
}

function parse_args() {
    while [[ $# -gt 0 ]];do
        key="$1"
        case $key in
        --delete-binary)
            export CLEAN_CLIENT="true"
        ;;
        --delete-crds)
            export DELETE_CRDS="true"
        ;;
        --namespace)
			check_option_value "--namespace" $2
            export ARENA_NAMESPACE=$2
			shift
        ;;
        --chart-dir)
			check_option_value "--chart-dir" $2
            export CHART_DIR=$2
			shift
        ;;
        --help|-h)
            help
            exit 0
        ;;
        *)
            # unknown option
            logger error "unknown option [$key]"
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
    export ARENA_NAMESPACE="arena-system"
	parse_args $@
	run 
}

main $@
