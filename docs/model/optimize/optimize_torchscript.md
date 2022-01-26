# Submit a model optimize job

This guide walks through the steps required to optimize a pytorch torchscript module with tensorrt.

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

3\. Submit a model optimize job.

```shell
$ arena model optimize \
    --name=resnet18-optimize \
    --namespace=default \
    --image=registry.cn-beijing.aliyuncs.com/kube-ai/easy-inference:1.0.0 \
    --image-pull-policy=Always \
    --gpus=1 \
    --optimizer=tensorrt \
    --data=model-pvc:/data \
    --model-config-file=/data/models/resnet18/config.json \
    --export-path=/data/models/resnet18 
    
job.batch/resnet18-optimize created
INFO[0002] The model optimize job resnet18-optimize has been submitted successfully
INFO[0002] You can run `arena model get resnet18-optimize` to check the job status
```

4\. List all the model optimize jobs.

```shell
$ arena model list

NAMESPACE      NAME               STATUS   TYPE      DURATION  AGE  GPU(Requested)
default-group  resnet18-optimize  RUNNING  Optimize  0s        1m   1
```

5\. Get model optimize job detail info.

```shell
$ arena model get resnet18-profile
Name:       resnet18-optimize
Namespace:  default-group
Type:       Optimize
Status:     RUNNING
Duration:   1m
Age:        1m
Parameters:
  --model-config-file  /data/models/resnet18/config.json
  --export-path        /data/models/resnet18


Instances:
  NAME                     STATUS             AGE  READY  RESTARTS  NODE
  ----                     ------             ---  -----  --------  ----
  resnet18-optimize-xrd6w  ContainerCreating  1m   0/1    0         cn-shenzhen.192.168.1.209
```

6\. After the optimize job finished, you can see a new torchscript modue named opt_resnet18.pt in --export-path.