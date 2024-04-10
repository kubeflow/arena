# Model Manage Guide

Welcome to the Arena Model Manage Guide! This guide covers how to use the `arena model` subcommand to manage registered model and model versions. This page outlines the most common situations and questions that bring readers to this section.

## Who Should Use this Guide?

If you want to use arena to manage models, this guide is for you. We have included detailed usages for managing models.

## Prerequisites

Arena now use [MLflow](https://mlflow.org/) as model registry backend, so you first need to run MLflow tracking server with database as storage backend beforehand. See [MLflow Tracking Server](https://mlflow.org/docs/latest/tracking/server.html) for detailed information.

## Setup

### Access MLflow Tracking Server In Non-proxied Mode

To access MLflow tracking server in non-proxied mode, you need to set up the `MLFLOW_TRACKING_URI` environment variable as follows:

```shell
export MLFLOW_TRACKING_URI=http://<tracking-server-host>:<port>
```

Replace `<tracking-server-host>` with the hostname or IP address of your MLflow tracking server, and `<port>` with the port number on which the tracking server is listening to.

### Access MLflow Tracking Server In Proxied Mode

If you run the MLflow tracking server within a Kubernetes cluster and do not set up the `MLFLOW_TRACKING_URI` environment variable, then Arena will search for services named `ack-mlflow` or `mlflow` across all namespaces and create a model client proxied by Kubernetes API server. If no such service is found, an error will be thrown. If multiple services are found, the first one will be used.

### Configure Basic Authentication

When the MLflow tracking server is secured with basic authentication, set up the `MLFLOW_TRACKING_USERNAME` and `MLFLOW_TRACKING_PASSWORD` environment variables to ensure that your MLflow client can authenticate with the tracking server successfully:

```shell
export MLFLOW_TRACKING_USERNAME=<username>
export MLFLOW_TRACKING_PASSWORD=<password>
```

Remember to replace `<username>` and `<password>` with your actual username and password for the MLflow tracking server.

<div style="background-color: #ffcccc; padding: 10px; border-left: 5px solid red;">
    <strong>Warning</strong><br />
    When accessing MLflow tracking server in proxied mode, basic authentication is not supported because the API server proxy will strip out Authorization HTTP header.
</div>

## Model Management

### Create a Model Version

```shell
$ arena model create \
    --name my-model \
    --tags key1,key2=value2 \
    --description "This is some description about my-model" \
    --version-tags key3,key4=value4 \
    --version-description "This is some description about my-model v1" \
    --source pvc://my-pvc/models/my-model/1
INFO[0000] registered model "my-model" created     
INFO[0000] model version 1 for "my-model" created  
```

### Get a Registered Model or Model Version

Get a registered model named `my-model`:

```shell
$ arena model get \
    --name my-model
Name:                my-model
LatestVersion        1
CreationTime:        2024-04-09T19:53:15+08:00
LastUpdatedTime:     2024-04-09T19:53:15+08:00
Description:
  This is some description about my-model
Tags:
  createdBy: arena
  key1: 
  key2: value2
Versions:
  Version    Source
  ---        ---
  1          pvc://my-pvc/models/my-model/1
```

Get model version `1` of registered model named `my-model`:

```shell
$ arena model get \
    --name my-model \
    --version 1
Name:                my-model
Version:             1
CreationTime:        2024-04-09T19:53:15+08:00
LastUpdatedTime:     2024-04-09T19:53:15+08:00
Source:              pvc://my-pvc/models/my-model/1
Description:
  This is some description about my-model v1
Tags:
  createdBy: arena
  key4: value4
  key3: 
```

### List All Registered Models

```shell
$ arena model list 
NAME                 LATEST_VERSION       LAST_UPDATED_TIME  
my-model             1                    2024-04-09T19:53:15+08:00 
```

### Update a Registered Model or Model Version

Update registered model named `my-model`:

```shell
$ arena model update \
    --name my-model \
    --description "This is some updated description" \
    --tags key1=updatedValue1,key2=updatedValue2 
INFO[0000] registered model "my-model" updated  
```

Update version `1` of model named `my-model`:

```shell
$ arena model update \
    --name my-model \
    --version 1 \
    --version-description "This is some updated description about version 1" \
    --version-tags key3=newValue3,key4=newValue4
INFO[0000] model version "my-model/1" updated 
```

If you want to delete tags, do as follows:

```shell
$ arena model update \
    --name my-model \
    --tags key1-,key2=value2- \
    --version 1 \
    --version-tags key3-,key4=value4-
INFO[0000] registered model "my-model" updated          
INFO[0000] model version "my-model/1" updated  
```

This will delete tag with key `key1` and `key2` of registered model named `my-model` and delete tag `key3` and `key4` of model version `1`.

### Delete a Registered Model or Model Version

Delete a registered model named `my-model` with confirmation:

```shell
$ arena model delete \
  --name my-model
Delete a registered model will cascade delete all its model versions. Are you sure you want to perform this operation? (yes/no)
yes
registered model "my-model" deleted
```

Or you can delete a registered model without confirmation by adding `--force` flag:

```shell
$ arena model delete \
  --name my-model \
  --force
registered model "my-model" deleted
```

Delete model version `1` of registered model named `my-model` with confirmation:

```shell
$ arena model delete \
    --name my-model \
    --version 1
Are you sure you want to perform this operation? (yes/no)
yes
model version "my-model/1" deleted
```

Or you can delete a model version without confirmation by adding `--force` flag:

```shell
$ arena model delete \
    --name my-model \
    --version 1 \
    --force
model version "my-model/1" deleted
```

<div style="background-color: #ffcccc; padding: 10px; border-left: 5px solid red;">
    <strong>Warning:</strong><br />
    Delete a registered model will cascade delete all its model versions, so you should do it carefully.
</div>

## Register a Model Version When Submitting a Training Job

### Submit a Training Job

When submitting a training job, you can register a model version at the same time as follows:

- `--model-name`: The name of the model to be registered. Upon successful submission of the training job, the model (if it doesn't exist) and a new model version will be created.
- `--model-source`: The model source is a URI that specifies the location of the model, for example `s3://my-bucket/path/to/model`, `pvc://namespace/pvc-name/path/to/model`. In this example, the model produced by the training is stored in the `/bloom-560m-sft` directory on the `training-data` pvc in the `default` namespace.

```shell
$ arena submit pytorchjob \
  --name=bloom-sft \
  --namespace=default \
  --gpus=1 \
  --image=registry.cn-hangzhou.aliyuncs.com/acs/deepspeed:v0.9.0-chat \
  --data=training-data:/model \
  --model-name=my-model \
  --model-source=pvc://default/training-data/bloom-560m-sft \
  "cd /model/DeepSpeedExamples/applications/DeepSpeed-Chat/training/step1_supervised_finetuning && bash training_scripts/other_language/run_chinese.sh /model/bloom-560m-sft"
pytorchjob.kubeflow.org/bloom-sft created
INFO[0001] The Job bloom-sft has been submitted successfully 
INFO[0001] You can run `arena get bloom-sft --type pytorchjob -n default` to check the job status 
INFO[0001] registered model "my-model" created
INFO[0001] model version 1 for "my-model" created
```

### Get Information About the Training Job

By querying information about the training job, we can know that this job is associated with version `1` of model named `my-model`:

```shell
$ arena get bloom-sft      
Name:          bloom-sft
Status:        PENDING
Namespace:     default
Priority:      N/A
Trainer:       PYTORCHJOB
Duration:      37s
CreateTime:    2024-04-10 16:36:39
EndTime:       
ModelName:     my-model
ModelVersion:  1
ModelSource:   pvc://default/training-data/bloom-560m-sft

Instances:
  NAME                STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
  ----                ------   ---  --------  --------------  ----
  bloom-sft-master-0  Pending  37s  true      1               N/A
```

### Get Information About the Model Version Associated with the Training Job

```shell
$ arena model get \
    --name my-model \
    --version 1
Name:                my-model
Version:             1
CreationTime:        2024-04-10T16:36:39+08:00
LastUpdatedTime:     2024-04-10T16:36:39+08:00
Source:              pvc://default/training-data/bloom-560m-sft
Description:
  arena submit pytorchjob \
      --data training-data:/model \
      --gpus 1 \
      --image registry.cn-hangzhou.aliyuncs.com/acs/deepspeed:v0.9.0-chat \
      --model-name my-model \
      --model-source pvc://default/training-data/bloom-560m-sft \
      --name bloom-sft \
      --namespace=default \
      "cd /model/DeepSpeedExamples/applications/DeepSpeed-Chat/training/step1_supervised_finetuning && bash training_scripts/other_language/run_chinese.sh /model/bloom-560m-sft"
Tags:
  createdBy: arena
  arena.kubeflow.org/uid: 3399d840e8b371ed7ca45dda29debeb1
  modelName: my-model
```

## Refer a Model Version When Submitting a Serving Job

### Submit a Serving Job

When submitting a serving job, you can associate it with a model by specifying `--model-name` and `--model-version` flags. It is necessary to ensure that the model used by the serving job is the one specified.

```shell
$ arena serve custom \
    --name=bloom-tgi-inference \
    --namespace=default \
    --gpus=1 \
    --version=v1 \
    --replicas=1 \
    --restful-port=8080 \
    --data=training-data:/model \
    --model-name=my-model \
    --model-version=1 \
    --image=text-generation-inference:0.8 \
    "text-generation-launcher --disable-custom-kernels --model-id /model/bloom-560m-sft --num-shard 1 -p 8080"
service/bloom-tgi-inference-v1 created
deployment.apps/bloom-tgi-inference-v1-custom-serving created
INFO[0001] The Job bloom-tgi-inference has been submitted successfully 
INFO[0001] You can run `arena serve get bloom-tgi-inference --type custom-serving -n default` to check the job status
```

### Get Information About the Serving Job

By querying information about the serving job, we can know that this job is associated with version `1` of model named `my-model`:

```shell
$ arena serve get bloom-tgi-inference
Name:          bloom-tgi-inference
Namespace:     default
Type:          Custom
Version:       v1
Desired:       1
Available:     0
Age:           7s
Address:       172.16.166.93
Port:          RESTFUL:8080
ModelName:     my-model
ModelVersion:  1
ModelSource:   pvc://default/training-data/bloom-560m-sft

Instances:
  NAME                                                    STATUS   AGE  READY  RESTARTS  NODE
  ----                                                    ------   ---  -----  --------  ----
  bloom-tgi-inference-v1-custom-serving-86cc9fb59c-dcxdp  Pending  7s   0/1    0
```

### Get Information About the Model Associated With the Serving Job

```shell
$ arena model get \
    --name my-model \
    --version 1
Name:                my-model
Version:             1
CreationTime:        2024-04-10T16:36:39+08:00
LastUpdatedTime:     2024-04-10T16:36:39+08:00
Source:              pvc://default/training-data/bloom-560m-sft
Description:
  arena submit pytorchjob \
      --data training-data:/model \
      --gpus 1 \
      --image registry.cn-hangzhou.aliyuncs.com/acs/deepspeed:v0.9.0-chat \
      --model-name my-model \
      --model-source pvc://default/training-data/bloom-560m-sft \
      --name bloom-sft \
      --namespace=default \
      "cd /model/DeepSpeedExamples/applications/DeepSpeed-Chat/training/step1_supervised_finetuning && bash training_scripts/other_language/run_chinese.sh /model/bloom-560m-sft"
Tags:
  createdBy: arena
  arena.kubeflow.org/uid: 3399d840e8b371ed7ca45dda29debeb1
  modelName: my-model
```
