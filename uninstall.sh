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

set -eux

HELM=${HELM:-arena-helm}
KUBECTL=${KUBECTL:-arena-kubectl}
RELEASE_NAME=${ARENA_RELEASE_NAME:-arena-artifacts}
RELEASE_NAMESPACE=${ARENA_RELEASE_NAMESPACE:-arena-system}
CLEAN_CLIENT=false
ARTIFACTS_DIR=""
CHART_DIR=""

function help() {
    echo -e "
Usage:

    arena-uninstall [OPTION1] [OPTION2] ...

Options:
	--namespace  string              Specify the namespace to delete arena
	--delete-binary                  Clean the client env,include ~/charts and /usr/local/bin/arena
	--chart-dir                      Specify the chart dir
"
}

function logger() {
    timestr=$(date +"%Y-%m-%d/%H:%M:%S")
    level=$(echo $1 | tr 'a-z' 'A-Z')
    echo ${timestr}"  "${level}"  "$2
}

function delete() {
    if ${HELM} status -n ${RELEASE_NAMESPACE} ${RELEASE_NAME} &>/dev/null; then
        ${HELM} delete -n ${RELEASE_NAMESPACE} ${RELEASE_NAME}
    else
        ${HELM} template ${RELEASE_NAME} -n ${RELEASE_NAMESPACE} $ARTIFACTS_DIR | ${KUBECTL} delete --ignore-not-found -f -
    fi
}

function delete_client() {
    rm -rf /charts
    rm -rf /usr/local/bin/arena*
}

function detect_chart_dir() {
    if [[ $CHART_DIR != "" ]]; then
        ARTIFACTS_DIR=$CHART_DIR
    elif [ -d arena-artifacts ]; then
        ARTIFACTS_DIR=arena-artifacts
    elif [ -d ~/charts/arena-artifacts ]; then
        ARTIFACTS_DIR=~/charts/arena-artifacts
    else
        ARTIFACTS_DIR=/charts/arena-artifacts
    fi
}

function parse_args() {
    while [[ $# -gt 0 ]]; do
        key="$1"
        case $key in
        --delete-binary)
            CLEAN_CLIENT=true
            ;;
        --namespace)
            check_option_value "--namespace" $2
            RELEASE_NAMESPACE=$2
            shift
            ;;
        --chart-dir)
            check_option_value "--chart-dir" $2
            CHART_DIR=$2
            shift
            ;;
        --help | -h)
            help
            exit 0
            ;;
        *)
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
    if [[ $value == "" ]] || echo "$value" | grep -- "^--" &>/dev/null; then
        logger error "the option $option not set value,please set it"
        exit 3
    fi
}

parse_args "$@"
detect_chart_dir
delete
if [[ $CLEAN_CLIENT == "true" ]]; then
    delete_client
fi
