# Run:AI
## Overview

Run:AI CLI is a command-line interface for the data scientists to run and monitor the machine learning training jobs on top of Run:AI software and Kubernestes.

## Prerequisites
* Kubernetes 1.15+
* Kubectl installed and configured to acceess your cluster. Please refer to https://kubernetes.io/docs/tasks/tools/install-kubectl/
* Install Helm. See https://v2.helm.sh/docs/using_helm/#quickstart . Run:AI currently supports Helm 2 only. Helm 3 (the default) is not supported
* Run:AI software installed on your Kubernetes cluster. Please refer to https://support.run.ai/hc/en-us/articles/360010280179-Installing-Run-AI-on-an-on-premise-Kubernetes-Cluster for installation, if you haven't done so already.
## Setup

Download latest release from the releases page. https://github.com/run-ai/arena/releases. Unarchive the downloaded file.
For installation:
```
sudo ./install-runai.sh
```
To verify installation:
```
runai --help
```
## Quickstart

For help on Run:AI CLI run
```
runai --help
```
To verify the status of your cluster, use the `top` command.
```
runai top job
runai top node
```
These commands will give you valuble information about your cluster's and jobs' GPUs allocation status.

To run a sample job using runai sample training container please run:
```
runai submit -g 1 --name runai-test --project {your_project} -i gcr.io/run-ai-lab/quickstart -g 1
```
This will run a job using Run:AI scheduler on top of the kubernetes cluster using Run:AI quickstart image and requesting 1 GPU for the job. Now to see the status of the running job please run:
```
runai list
```
Once the job in running, you can view its logs by running:
```
runai logs runai-test
```
At last, to delete the job prior to it's compleation you can run:
```
runai delete runai-test
```