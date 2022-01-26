# Submit a model benchmark job

This guide walks through the steps required to benchmark a pytorch torchscript module.

1\. The first step is to check the available resources:

```shell
$ arena top node

NAME                       IPADDRESS      ROLE    STATUS  GPU(Total)  GPU(Allocated)
cn-shenzhen.192.168.1.209  192.168.1.209  <none>  Ready   1           0
cn-shenzhen.192.168.1.210  192.168.1.210  <none>  Ready   1           0
cn-shenzhen.192.168.1.211  192.168.1.211  <none>  Ready   1           0
---------------------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
0/3 (0.0%)
```

There are 3 available nodes with GPU for running model profile job.

2\. Prepare the model to profile and configuration.

In this example, we will profile a pytorch resnet18 model. We need save the resnet18 model as a torchscript module firstly.

```python
import torch
import torchvision

# An instance of your model.
model = torchvision.models.resnet18()

# An example input you would normally provide to your model's forward() method.
dummy_input = torch.rand(1, 3, 224, 224)

# Use torch.jit.trace to generate a torch.jit.ScriptModule via tracing.
traced_script_module = torch.jit.trace(model, dummy_input)

torch.jit.save(traced_script_module, "resnet18.pt")
```

Then give a profile configuration file named config.json like below.

```json
{
  "model_name": "resnet18",
  "model_platform": "torchscript",
  "model_path": "/data/models/resnet18/resnet18.pt",
  "inputs": [
    {
      "name": "input",
      "data_type": "float32",
      "shape": [1, 3, 224, 224]
    }
  ],
  "outputs": [
    {
      "name": "output",
      "data_type": "float32",
      "shape": [ 1000 ]
    }
  ]
}
```

3\. Submit a model benchmark job.

```shell
$ arena model benchmark \
  --name=resnet18-benchmark \
  --namespace=default \
  --image=registry.cn-beijing.aliyuncs.com/kube-ai/easy-inference:1.0.0 \
  --image-pull-policy=Always \
  --gpus=1 \
  --data=model-pvc:/data \
  --model-config-file=/data/modeljob/models/resnet18/benchmark.json \
  --report-path=/data/modeljob/models/resnet18 \
  --concurrency=5 \
  --requests=1000 \
  --duration=60
    
job.batch/resnet18-benchmark created
INFO[0000] The model benchmark job resnet18-benchmark has been submitted successfully
INFO[0000] You can run `arena model get resnet18-benchmark` to check the job status
```

4\. List all the model benchmark jobs.

```shell
$ arena model list

NAMESPACE      NAME                STATUS   TYPE       DURATION  AGE  GPU(Requested)
default  resnet18-benchmark  RUNNING  Benchmark  23s       23s  1
```

5\. Get model benchmark job detail info.

```shell
$ arena model get resnet18-benchmark
Name:       resnet18-benchmark
Namespace:  default
Type:       Benchmark
Status:     RUNNING
Duration:   45s
Age:        45s
Parameters:
  --model-config-file  /data/models/resnet18/benchmark.json
  --report-path        /data/models/resnet18
  --concurrency        5
  --requests           1000
  --duration           60
GPU:        1

Instances:
  NAME                      STATUS   AGE  READY  RESTARTS  GPU  NODE
  ----                      ------   ---  -----  --------  ---  ----
  resnet18-benchmark-gvj97  Running  45s  1/1    0         1    cn-beijing.192.168.94.82
```

6\. After the benchmark job finished, you can find a file named benchmark_result.txt which contains the benchmark result int the specified --report-path.

```
Benchmark options:
{"batch_size": 1, "concurrency": 5, "duration": 60, "requests": 1000, "model_config": {"model_name": "resnet18", "model_platform": "torchscript", "model_path": "/data/modeljob/models/resnet18/resnet18.pt", "inputs": [{"name": "input", "data_type": "float32", "shape": [1, 3, 224, 224]}], "outputs": [{"name": "output", "data_type": "float32", "shape": [1000]}]}, "report_path": "/data/modeljob/models/resnet18"}
Benchmark finished, cost 60.00157570838928 s
Benchmark result:
{"p90_latency": 3.806, "p95_latency": 3.924, "p99_latency": 4.781, "min_latency": 3.665, "max_latency": 1555.418, "mean_latency": 3.88, "median_latency": 3.731, "throughput": 257, "gpu_mem_used": 1.47, "gpu_utilization": 38.39514839785918}
```




