#!/bin/bash
set -e

function help() {
    echo -e "
Usage:

    install.sh [OPTION1] [OPTION2] ...

Options:
    --kubeconfig string              Specify the kubeconfig file
	--delete-binary                  Clean the client env,include ~/charts and /usr/local/bin/arena
	--delete-crds	                 Delete the CRDs,Warning: this option will delete the training jobs
	--chart-dir                      Specify the chart dir
"

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
	if arena-helm list -n arena-system | grep arena-artifacts &> /dev/null;then
		arena-helm delete arena-artifacts -n arena-system
	fi
	arena-helm template arena-artifacts -n arena-system $ARTIFACTS_DIR > /tmp/arena-artifacts.yaml
	arena-kubectl delete -f /tmp/arena-artifacts.yaml
	arena-kubectl delete ns arena-system
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
	parse_args $@
	run 
}

main $@