﻿# Arena

[![Integration Test](https://github.com/kubeflow/arena/actions/workflows/integration.yaml/badge.svg)](https://github.com/kubeflow/arena/actions/workflows/integration.yaml)[![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/arena)](https://goreportcard.com/report/github.com/kubeflow/arena)

## 概述

Arena 是一个命令行工具，可供数据科学家轻而易举地运行和监控机器学习训练作业，并便捷地检查结果。目前，它支持单机/分布式深度学习模型训练。在实现层面，它基于 Kubernetes、helm 和 Kubeflow。但数据科学家可能对于 kubernetes 知之甚少。

与此同时，用户需要 GPU 资源和节点管理。Arena 还提供了 `top` 命令，用于检查 Kubernetes 集群内的可用 GPU 资源。

简而言之，Arena 的目标是让数据科学家感觉自己就像是在一台机器上工作，而实际上还可以享受到 GPU 集群的强大力量。

## 设置

您可以按照 [安装指南](https://arena-docs.readthedocs.io/en/latest/installation) 执行操作

## 用户指南

Arena 是一种命令行界面，支持轻而易举地运行和监控机器学习训练作业，并便捷地检查结果。目前，它支持独立/分布式训练。

- [1.使用 git 上的源代码运行训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/standalone/)
- [2.使用 tensorboard 运行训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/tensorboard/)
- [3.运行分布式训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/distributed/)
- [4.使用外部数据运行分布式训练作业](https://arena-docs.readthedocs.io/en/latest/training/tfjob/dataset/)
- [5.运行基于 MPI 的分布式训练作业](https://arena-docs.readthedocs.io/en/latest/training/mpijob/distributed/)
- [6.使用群调度器运行分布式 TensorFlow 训练作业](https://arena-docs.readthedocs.io/en/latest/training/etjob/elastictraining-tensorflow2-mnist/)
- [7.运行 TensorFlow Serving](https://arena-docs.readthedocs.io/en/latest/serving/tfserving/serving/)

## 演示

[![arena demo](demo.jpg)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/50210690772.mp4)

## 开发

先决条件：

- Go >= 1.8

```shell
mkdir -p $(go env GOPATH)/src/github.com/kubeflow
cd $(go env GOPATH)/src/github.com/kubeflow
git clone https://github.com/kubeflow/arena.git
cd arena
make
```

`arena` 二进制文件位于 `arena/bin` 目录下。您可能希望将目录添加到 `$PATH`。

## 命令行文档

请参阅 [arena.md](docs/cli/arena.md)

## 路线图

请参阅[路线图](ROADMAP.md)
