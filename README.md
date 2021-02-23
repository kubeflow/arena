# Arena

[![CircleCI](https://circleci.com/gh/kubeflow/arena.svg?style=svg)](https://circleci.com/gh/kubeflow/arena)
[![Build Status](https://travis-ci.org/kubeflow/arena.svg?branch=master)](https://travis-ci.org/kubeflow/arena) 
[![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/arena)](https://goreportcard.com/report/github.com/kubeflow/arena)


## Overview

Arena is a command-line interface for the data scientists to run and monitor the machine learning training jobs and check their results in an easy way. Currently it supports solo/distributed TensorFlow training. In the backend, it is based on Kubernetes, helm and Kubeflow. But the data scientists can have very little knowledge about kubernetes.

Meanwhile, the end users require GPU resource and node management. Arena also provides `top` command to check available GPU resources in the Kubernetes cluster.

In one word, Arena's goal is to make the data scientists feel like to work on a single machine but with the Power of GPU clusters indeed.

For the Chinese version, please refer to [中文文档](README_cn.md)

## Setup

You can follow up the [Installation guide](docs/installation/INSTALL_FROM_BINARY.md)

## User Guide

Arena is a command-line interface to run and monitor the machine learning training jobs and check their results in an easy way. Currently it supports solo/distributed training.

- [1. Run a training Job with source code from git](docs/userguide/1-tfjob-standalone.md)
- [2. Run a training Job with tensorboard](docs/userguide/2-tfjob-tensorboard.md)
- [3. Run a distributed training Job](docs/userguide/3-tfjob-distributed.md)
- [4. Run a distributed training Job with external data](docs/userguide/4-tfjob-distributed-data.md)
- [5. Run a distributed training Job based on MPI](docs/userguide/5-mpijob-distributed.md)
- [6. Run a distributed TensorFlow training job with gang scheduler](docs/userguide/6-tfjob-gangschd.md)
- [7. Run TensorFlow Serving](docs/userguide/7-tf-serving.md)
- [8. Run TensorFlow Estimator](docs/userguide/8-tfjob-estimator.md)
- [9. Monitor GPUs of the training job ](docs/userguide/9-top-job-gpu-metric.md)
- [10. Run a distributed training job with RDMA](docs/userguide/10-rdma-integration.md)
- [11. Run a distributed spark job](docs/userguide/11-sparkjob-distributed.md)
- [12. Run a Volcano job](docs/userguide/12-volcanojob.md)
- [13. Preempted mpi job](docs/userguide/13-preempted-mpijob.md)
- [14. Submit jobs with node selectors](docs/userguide/14-submit-with-node-selector.md)
- [15. Submit jobs with tolerating taints](docs/userguide/14-submit-with-node-toleration.md)
- [16. Run a custom serving job](docs/userguide/15-custom-serving-sample.md)
- [17. Run a training Job with configuration files](docs/userguide/16-assign-config-file.md)
- [18. Run a standalone Pytorch Job](docs/userguide/17-pytorchjob-standalone.md)
- [19. Run a distributed Pytorch Job](docs/userguide/18-pytorchjob-distributed.md)
- [20. Run a KFServing Job](docs/userguide/27-kfserving-custom.md)
- [21. Run a Elastic Training Job](docs/userguide/28-elastictraining-tensorflow2-mnist.md)
- [21. Run a Seldon Core Job](docs/userguide/32-seldon-serving.md)
## Demo

[![](demo.jpg)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/50210690772.mp4)


## Developing

Prerequisites:

- Go >= 1.8

```
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/kubeflow/arena.git
cd arena
make
```

`arena` binary is located in directory `arena/bin`. You may want to add the directory to `$PATH`.

Then you can follow [Installation guide for developer](docs/installation/INSTALL_FROM_SOURCE.md)

## CPU Profiling

```
# set profile rate (HZ)
export PROFILE_RATE=1000

# arena {command} --pprof
arena list --pprof
INFO[0000] Dump cpu profile file into /tmp/cpu_profile
```

Then you can analyze the profile by following [Go CPU profiling: pprof and speedscope](https://coder.today/go-profiling-pprof-and-speedscope-b05b812cc429)


## Adopters

If you are intrested in Arena and would like to share your experiences with others, you are warmly welcome to add your information on [ADOPTERS.md](docs/ADOPTERS.md) page. We will continuousely discuss new requirements and feature design with you in advance.


## FAQ

Please refer to [FAQ](FAQ.md)

## CLI Document

Please refer to [arena.md](docs/cli/arena.md)

## RoadMap

See [RoadMap](ROADMAP.md)
