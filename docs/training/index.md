# Training Job Guide

Welcome to the Arena Training Job Guide! This guide covers how to use the `arena cli` to manage the training job. This page outlines the most common situations and questions that bring readers to this section.


## Who should use this guide?

If you want to use arena to manage training jobs, this guide is for you. we have included detailed usages for managing training jobs.

## Manage The Training Jobs

* How to [list all training jobs](common/list_jobs.md).
* How to [get the training job details](common/get_job.md).
* How to [attach the training job](common/attach_job.md).
* How to [get the training job logs](common/get_job_logs.md). 
* How to [delete the training jobs](common/delete_jobs.md).
* How to [clean up the finished training jobs](common/prune_jobs.md). 

## Tensorflow Training Job Guide

* I want to [submit a standalone tensorflow training job](tfjob/standalone.md).
* I want to [submit a tensorflow training job with specified a tensorboard](tfjob/tensorboard.md).
* I want to [submit a distributed tensorflow training job](tfjob/distributed.md).
* I want to [submit a tensorflow training job with specified datasets](tfjob/dataset.md).
* I want to [submit a tensorflow training job with gang scheduling enabled](tfjob/gangschd.md).
* I want to [submit a tensorflow training job with specified estimator](tfjob/estimator.md).
* I want to [submit a tensorflow training job with specified node selectors](tfjob/selector.md).
* I want to [submit a tensorflow training job with specified taint nodes](tfjob/toleration.md).
* I want to [submit a tensorflow training job with specified configuration files](tfjob/assign_config_file.md).
* I want to [submit Tensorflow Job with specified role sequence](tfjob/role-sequence.md).

## MPI Training Job Guide

* I want to [submit a distributed MPI training job](mpijob/distributed.md).
* I want to [submit a distributed MPI training with gpu topology scheduling](mpijob/gputopology.md).
* I want to [preempt the MPI training job](mpijob/preempted.md).
* I want to [submit a MPI training job with specified tolerations](mpijob/toleration.md).
* I want to [submit a MPI training job with specified node selectors](mpijob/selector.md).
* I want to [submit a MPI training job with specified configuration files](mpijob/assign_config_file.md).
* I want to [submit a MPI training job with specified rdma devices](mpijob/rdma.md).


## Pytorch Training Job Guide

* I want to [submit a standalone pytorch training job](pytorchjob/standalone.md).
* I want to [submit a distributed pytorch training job](pytorchjob/distributed.md).
* I want to [submit a pytorch training job with specified tensorboard](pytorchjob/tensorboard.md).
* I want to [submit a pytorch training job with specified datasets](pytorchjob/distributed-data.md).
* I want to [submit a pytorch training job with specified node selectors](pytorchjob/node-selector.md).
* I want to [submit a pytorch training job with specified node tolerations](pytorchjob/node-toleration.md).
* I want to [submit a pytorch training job with specified configuration files](pytorchjob/assign-config-file.md).
* I want to [preempt the pytorch training job](pytorchjob/preempted.md).
* I want to [submit a pytorch training job with specified cleaning task policy](pytorchjob/clean-pod-policy.md).  

## Elastic Training Job Guide

* I want to [submit a elastic training job(pytorch)](etjob/elastictraining-pytorch-synthetic.md).
* I want to [submit a elastic training job(tensorflow)](etjob/elastictraining-tensorflow2-mnist.md).

## Cron Training Job Guide

* I want to [submit a cron training job(tensorflow)](cron/cron-tfjob.md).

## Spark Training Job Guide

* I want to [submit a distributed spark training job](sparkjob/distributed.md).

## Volcano Training Job Guide

* I want to [submit a volcano training job](volcanojob/volcanojob.md).

## Ray Training Job Guide

* I want to [submit a ray training job](rayjob/rayjob.md).

## Other Usage

* I want to [submit a training job with specified the imagePullSecrets](common/image-pull-secret.md).
