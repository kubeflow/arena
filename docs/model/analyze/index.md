# Model Analyze Guide

<div style="background-color: #e0f2f4; padding: 10px; border-left: 5px solid #e0f2f4;">
    <strong>Note</strong><br />
    This feature is still experimental and may change in a future release without warning.
</div>

Welcome to the Arena Model Analyze Guide! This guide covers how to use the `arena cli` to profile the model to find performance bottleneck, and how to use tensorrt to optimize the inference performance, you can also benchmark the model to get inference metrics like qps, latency, gpu usage and so on. This page outlines the most common situations and questions that bring readers to this section.

## Who should use this guide?

After training you may get some models. If you want to know the model performance, and get some guidance to optimize the model if the performance is not meet you requirements, this guide is for you. we have included detailed usages for managing model profile and optimize job.

## Profile the model

* How to [profile the pytorch torchscript module](profile/profile_torchscript.md).

## Optimize the model

* I want to [optimize the torchscript module with tensorrt](optimize/optimize_torchscript.md).

## Benchmark the model inference

* I want to [benchmark the torchscript inference performance](benchmark/benchmark_torchscript.md).
