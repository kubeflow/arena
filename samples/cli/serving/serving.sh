#!/bin/sh

# kserve custom runtime
arena serve kserve \
    --name=sklearn-iris-2 \
    --model-format=sklearn \
    --image=kube-ai-registry.cn-shanghai.cr.aliyuncs.com/ai-sample/kserve-sklearn-server:v0.12.0 \
    --memory=100Mi \
    --command="python -m klearnserver --model_name=sklearn-iris --model_dir=/models --http_port=8080"