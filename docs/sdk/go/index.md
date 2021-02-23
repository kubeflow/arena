# Arena Go SDK Guide

Welcome to the Arena Go SDK Guide!

## **Who should use this guide?**

If you want to integrate arena to your GO project, for example, submit a training job and get the training job details, this guide is for you. The guide has included detailed information for you to how to use Arena Go SDK.

## **API for creating ArenaClient**

ArenaClient is the entry point of all APIs and it can create other sub client such as TrainingClient or ServingClient, It is important. Please refer:

* I want to [create a ArenaClient](./arena_client.md).

## **APIs for managing training jobs**

The APIs of managing training jobs have two parts: 

* Common operations for all training jobs
* Customly building a training job

If you want to get a completed example about how to use apis to managing training jobs,please refer examples stored in the directory "samples/sdk/tfjob". 

#### Common operations for all training jobs 

* How to use api to [list training jobs](./training/list.md).
* How to use api to [get the training job details](./training/get.md).
* How to use api to [get the training job logs](./training/logs.md).
* How to use api to [get the training job logviewer](./training/logviewer.md).
* How to use api to [delete the training job](./training/delete.md).
* How to use api to [clean up all finished training jobs](./training/prune.md).
* How to use api to [submit a training job](./training/submit.md).

#### Customly building a training job

* I want to customly [build a tensorflow training job](./training/tfjob.md).
* I want to customly [build a pytorch training job](./training/pytorchjob.md).
* I want to customly [build a mpi training job](./training/mpijob.md).
* I want to customly [build a elastic training job](./training/etjob.md).
* I want to customly [build a spark training job](./training/sparkjob.md).
* I want to customly [build a volcano training job](./training/volcanojob.md). 


## **APIs for managing serving jobs**

The APIs of managing serving jobs have two parts: 

* Common operations for all serving jobs
* Customly building a serving job

If you want to get a completed example about how to use apis to managing serving jobs,please refer examples stored in the directory "samples/sdk/custom-serving". 

#### **Common operations for all serving jobs**

* How to use api to [list serving jobs](./serving/list.md).
* How to use api to [get the serving job details](./serving/get.md).
* How to use api to [get the serving job logs](./serving/logs.md).
* How to use api to [delete the serving job](./serving/delete.md).
* How to use api to [submit a serving job](./serving/submit.md).

#### **Customly building a serving job**

* I want to customly [build a tensorflow serving job](./serving/tfserving.md).
* I want to customly [build a tensorrt serving job](./serving/trtserving.md).
* I want to customly [build a custom serving job](./serving/customserving.md).
* I want to customly [build a kubeflow serving job](./serving/kfserving.md).
  
## **APIs for querying cluster nodes**

The following APIs are used to query cluster nodes.

* I want to [query cluster nodes](./top_node.md).