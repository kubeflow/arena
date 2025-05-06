# Arena

[![GitHub release](https://img.shields.io/github/v/release/kubeflow/arena)](https://github.com/kubeflow/arena/releases) [![Integration Test](https://github.com/kubeflow/arena/actions/workflows/integration.yaml/badge.svg)](https://github.com/kubeflow/arena/actions/workflows/integration.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/arena)](https://goreportcard.com/report/github.com/kubeflow/arena)

View the [Arena documentation](https://arena-docs.readthedocs.io/en/latest).

## Overview

Arena is a command-line interface for the data scientists to run and monitor the machine learning training jobs and check their results in an easy way. Currently it supports solo/distributed TensorFlow training. In the backend, it is based on Kubernetes, helm and Kubeflow. But the data scientists can have very little knowledge about kubernetes.

Meanwhile, the end users require GPU resource and node management. Arena also provides `top` command to check available GPU resources in the Kubernetes cluster.

In one word, Arena's goal is to make the data scientists feel like to work on a single machine but with the Power of GPU clusters indeed.

For the Chinese version, please refer to [中文文档](README_cn.md)

## Setup

You can follow up the [Installation guide](https://arena-docs.readthedocs.io/en/latest/installation)

## User Guide

Arena is a command-line interface to run and monitor the machine learning training jobs and check their results in an easy way. Please refer the [User Guide](https://arena-docs.readthedocs.io/en/latest/training) to manage your training jobs.

## Demo

[![arena demo](demo.jpg)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/50210690772.mp4)

## Developing

Prerequisites:

- Go >= 1.8

```shell
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/kubeflow/arena.git
cd arena
make
```

`arena` binary is located in directory `arena/bin`. You may want to add the directory to `$PATH`.

Then you can follow [Installation guide for developer](https://arena-docs.readthedocs.io/en/latest/installation)

## CPU Profiling

```shell
# set profile rate (HZ)
export PROFILE_RATE=1000

# arena {command} --pprof
arena list --pprof
INFO[0000] Dump cpu profile file into /tmp/cpu_profile
```

Then you can analyze the profile by following [Go CPU profiling: pprof and speedscope](https://coder.today/go-profiling-pprof-and-speedscope-b05b812cc429)

## Adopters

If you are interested in Arena and would like to share your experiences with others, you are warmly welcome to add your information on [ADOPTERS.md](docs/about/ADOPTERS.md) page. We will continuously discuss new requirements and feature design with you in advance.

## FAQ

Please refer to [FAQ](https://arena-docs.readthedocs.io/en/latest/faq).

## CLI Document

Please refer to [arena.md](docs/cli/arena.md).

## RoadMap

See [RoadMap](ROADMAP.md).
