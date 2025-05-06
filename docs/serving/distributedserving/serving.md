# Submit a distributed serving job

This guide walks through the steps to deploy and serve a model on two nodes and each node has 1 gpu. To illustrate usage, we use the [Qwen2-1.5B](https://modelscope.cn/models/Qwen/Qwen2-1.5B) downloaded from modelscope and deploy it by using vllm with pipeline-parallel-size equals to 2.

## Prerequisites

- Install LeaderWorkerSet API to your k8s cluster following this [guide](https://github.com/kubernetes-sigs/lws/blob/main/docs/setup/install.md) (required)
- Create a pvc named `test-pvc` with models to deploy

## Steps

1\. Submit vllm distributed serving job with:

    $ arena serve distributed \
        --name=vllm \
        --version=alpha \
        --restful-port=5000 \
        --image=vllm/vllm-openai:latest \
        --data=test-pvc:/mnt/models \
        --masters=1 \
        --master-gpus=1
        --master-command="ray start --head --port=6379; vllm serve /mnt/models/Qwen2-1.5B \
            --port 5000 \
            --dtype half \
            --pipeline-parallel-size 2" \
        --workers=1 \
        --worker-gpus=1 \
        --worker-command="ray start --address=\$(MASTER_ADDR):6379 --block" \
        --share-memory=4Gi \
        --startup-probe-action=httpGet \
        --startup-probe-action-option="path: /health" \
        --startup-probe-action-option="port: 5000" \
        --startup-probe-option="periodSeconds: 60" \
        --startup-probe-option="failureThreshold: 5"
    configmap/vllm-alpha-cm created
    service/vllm-alpha created
    leaderworkerset.leaderworkerset.x-k8s.io/vllm-alpha-distributed-serving created
    INFO[0002] The Job vllm has been submitted successfully 
    INFO[0002] You can run `arena serve get vllm --type distributed-serving -n default` to check the job status
    
In this example, we use `MASTER_ADDR` to get the address of the master pod in worker command. This environment variable is automatically injected into all pods by Arena when it creates the job. Besides this variable, there are several others environment variables created by Arena in order to help user to deploy the distributed serving job:

- `WORLD_SIZE`: The number of pod in one replica, equals to `#masters + #workers`.
- `POD_NAME`: The name of current pod.
- `POD_INDEX`: The index of current pod in the replica, starts from 0 which can only be master pod
- `GROUP_INDEX`: The index of current replica, starts from 0.
- `HOSTFILE`: The hostfile path of current replica, which contains the hostname of all pods in the replica.
- `GPU_COUNT`: The number of gpu of current pod.
- `ROLE`: The role of current pod, can only be master or worker.

!!! note

    To run a distributed serving job, you need to specify:

    - `--workers`: The number of workers (default is 0).
    - `--master-gpus`: GPUs of master (default is 0).
    - `--worker-gpus`: GPUs of each worker (default is 0).
    - `--master-command`: The command to run on master (required).
    - `--worker-command`: The command to run on each worker (required).
    
    If you do not explicitly specify the command for the master and worker but run the command like following format:

        $ arena serve distributed --name=test ... "command"

    arena will automatically run this command on both the master and the worker.

!!! warning
    To avoid ambiguity, the distributed serving job exposes the `--masters` parameter to user. But currently arena does not support modifying master numbers in distributed serving job. By default, one replica can only have one master pod.

2\. List the job you were just serving

    $ arena serve list --type=distributed
    NAME  TYPE         VERSION  DESIRED  AVAILABLE  ADDRESS      PORTS         GPU
    vllm  Distributed  alpha    1        1          172.21.5.50  RESTFUL:5000  2

3\. Test the model service 
    
    $ kubectl get svc | grep vllm
    vllm-alpha                        ClusterIP   172.21.13.60    <none>        5000/TCP         3m24s
    vllm-alpha-distributed-serving    ClusterIP   None            <none>        <none>           3m24s

    $ kubectl port-forward svc/vllm-alpha 5000:5000
    Forwarding from 127.0.0.1:5000 -> 5000
    Forwarding from [::1]:5000 -> 5000

    # check model service
    $ curl -X POST http://localhost:5000/v1/completions \
           -H "Content-Type: application/json" \
           -d '{
                    "model": "/mnt/oss/models/Qwen2-1.5B",
                    "prompt": "Please count from 1 to 10: 1, 2",
                    "max_tokens": 32
               }'
    {"id":"cmpl-9f6c8d3be9ae476ca0cc3e393ec370a6","object":"text_completion","created":1730190958,"model":"/mnt/oss/models/Qwen2-1.5B","choices":[{"index":0,"text":", 3, 4, 5, 6, 7, 8, 9, 10. These observations boiling point. According","logprobs":null,"finish_reason":"length","stop_reason":null}],"usage":{"prompt_tokens":15,"total_tokens":47,"completion_tokens":32}}


4\. Delete the inference service
    
    $ arena serve delete vllm
    INFO[0001] The serving job vllm with version alpha has been deleted successfully