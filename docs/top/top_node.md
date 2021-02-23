# Display GPU Usage For Nodes

The `arena top node` command allows you to see the gpu resource consumption for nodes.

## Supported GPU Modes

The `arena top node` command supports to display node details, which has different GPU modes. Currently supports 4 GPU Modes:

* none: the node has no gpus 
* exclusive: the node has gpus and owns kubernetes extend resource "nvidia.com/gpu".
* share: the node has gpus and owns kubernetes extend resource "aliyun.com/gpu-mem".
* topology: the node has gpus and owns kubernetes extend resource "aliyun.com/gpu".


## Usage

1\. The following command will help you to display gpu resource consumption for all nodes:

```
$ arena top node
NAME                      IPADDRESS      ROLE    STATUS    GPU(Total)  GPU(Allocated)
cn-beijing.192.168.8.10   192.168.8.10   <none>  Ready     0           0
virtual-kubelet           172.27.5.28    agent   Ready     0           0
cn-beijing.192.168.1.135  192.168.1.135  <none>  NotReady  1           0
cn-beijing.192.168.1.136  192.168.1.136  <none>  NotReady  1           0
cn-beijing.192.168.1.137  192.168.1.137  <none>  Ready     1           0
cn-beijing.192.168.8.3    192.168.8.3    <none>  Ready     1           1
---------------------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
1/4 (25.0%)
```

2\. If you only care gpu resource consumption for some nodes, you can specify the nodes:

```
$ arena top node cn-beijing.192.168.8.10 cn-beijing.192.168.1.136
NAME                      IPADDRESS      ROLE    STATUS    GPU(Total)  GPU(Allocated)
cn-beijing.192.168.8.10   192.168.8.10   <none>  Ready     0           0
cn-beijing.192.168.1.136  192.168.1.136  <none>  NotReady  1           0
```

3\. If you want to display gpu resource consumption for nodes with specified gpu mode,you can use '-m' to filter:

```
$ arena top node -m e
NAME                      IPADDRESS      ROLE    STATUS    GPU(Total)  GPU(Allocated)
cn-beijing.192.168.1.135  192.168.1.135  <none>  NotReady  1           0
cn-beijing.192.168.1.136  192.168.1.136  <none>  NotReady  1           0
cn-beijing.192.168.1.137  192.168.1.137  <none>  Ready     1           0
cn-beijing.192.168.8.3    192.168.8.3    <none>  Ready     1           1
---------------------------------------------------------------------------------------------------
Allocated/Total GPUs of nodes which own resource nvidia.com/gpu In Cluster:
1/4 (25.0%)
```

"e" represents "exclusive", This command only display node which owns kubernetes resource "nvidia.com/gpu". The following command can help you to get the supported gpu modes:

```
$ arena top node  -h | grep mode
  -m, --gpu-mode string   Display node information with following gpu mode:[n(none)|e(exclusive)|t(topology)|s(share)]
```

4\. If you want to get more information of the node, "-d" is requried:

```
$ arena top node -d cn-beijing.192.168.8.3

Name:    cn-beijing.192.168.8.3
Status:  Ready
Role:    <none>
Type:    GPUExclusive
Address: 192.168.8.3
Description:
  1.This node is enabled gpu exclusive mode.
  2.Pods can request resource 'nvidia.com/gpu' to use gpu exclusive feature on this node
Instances:
  NAMESPACE  NAME                                                       STATUS   GPU(Requested)
  ---------  ----                                                       ------   --------------
  default    fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4  Running  1
GPU Summary:
  Total GPUs:     1
  Allocated GPUs: 1
  Unhealthy GPUs: 0
```

5\. If you need to monitor nodes in real time, "-r" will help you("-r" must work with "-d"):

```
$ arena top node  cn-beijing.192.168.8.3 -r -d

Name:    cn-beijing.192.168.8.3
Status:  Ready
Role:    <none>
Type:    GPUExclusive
Address: 192.168.8.3
Description:
  1.This node is enabled gpu exclusive mode.
  2.Pods can request resource 'nvidia.com/gpu' to use gpu exclusive feature on this node
Instances:
  NAMESPACE  NAME                                                       STATUS   GPU(Requested)
  ---------  ----                                                       ------   --------------
  default    fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4  Running  1
GPU Summary:
  Total GPUs:     1
  Allocated GPUs: 1
  Unhealthy GPUs: 0

------------------------- 2021-02-22 17:26:29 -------------------------------------

Name:    cn-beijing.192.168.8.3
Status:  Ready
Role:    <none>
Type:    GPUExclusive
Address: 192.168.8.3
Description:
  1.This node is enabled gpu exclusive mode.
  2.Pods can request resource 'nvidia.com/gpu' to use gpu exclusive feature on this node
Instances:
  NAMESPACE  NAME                                                       STATUS   GPU(Requested)
  ---------  ----                                                       ------   --------------
  default    fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4  Running  1
GPU Summary:
  Total GPUs:     1
  Allocated GPUs: 1
  Unhealthy GPUs: 0

------------------------- 2021-02-22 17:26:31 -------------------------------------
```

6\. Arena supports to show gpu metrics of nodes when "--metric" is enabled, this feature requires Prometheus and gpu-exporter has been existed in cluster.

```
$ arena top node cn-beijing.192.168.8.3 --metric -d

Name:    cn-beijing.192.168.8.3
Status:  Ready
Role:    <none>
Type:    GPUExclusive
Address: 192.168.8.3
Description:
  1.This node is enabled gpu exclusive mode.
  2.Pods can request resource 'nvidia.com/gpu' to use gpu exclusive feature on this node
Instances:
  NAMESPACE  NAME                                                       STATUS   GPU(Requested)  GPU(Allocated)
  ---------  ----                                                       ------   --------------  --------------
  default    fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4  Running  1               gpu0
GPUs:
  INDEX  MEMORY(Total)  MEMORY(Allocated)  MEMORY(Used)  DUTY_CYCLE
  -----  -------------  -----------------  ------------  ----------
  0      14.7 GiB       14.7 GiB           0.0 GiB       0.0%
GPU Summary:
  Total GPUs:           1
  Allocated GPUs:       1
  Unhealthy GPUs:       0
  Total GPU Memory:     14.7 GiB
  Allocated GPU Memory: 14.7 GiB
  Used GPU Memory:      0.0 GiB
```

