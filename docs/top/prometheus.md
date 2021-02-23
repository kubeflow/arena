# Work with Prometheus

The command ``arena top job <job name>`` or ``arena top node -d`` can display GPU monitoring metrics. Before using it, you must deploy a Prometheus and nodeExporter for GPU Metrics.

1\. Deploy a Prometheus

```
$ kubectl apply -f kubernetes-artifacts/prometheus/prometheus.yaml
```

2\. Deploy GPU node exporter

If your cluster is ACK (Alibaba Cloud Kubernetes) cluster, you can just exec command:

```
# change gpu export nodeSelector to aliyun label
$ sed -i 's|accelerator/nvidia_gpu|aliyun.accelerator/nvidia_count|g' kubernetes-artifacts/prometheus/gpu-expoter.yaml
```


If your cluster is not ACK cluster, you need to label your GPU node:
```
# label all your GPU nodes
$ kubectl label node <your GPU node> accelerator/nvidia_gpu=true
```

Deploy gpu exporter

```
$ kubectl apply -f kubernetes-artifacts/prometheus/gpu-exporter.yaml
```

!!! note

    * the prometheus and gpu-exporter components should be deployed in namespace ``kube-system``, and so that ``arena top job <job name>`` can work.

    * if the your prometheus has been existed in cluster,please make sure the k8s service whose port is 9090 has the label `kubernetes.io/service-name=prometheus-server` 
  
3\. You can check the GPU metrics by prometheus SQL request

```
$ kubectl get --raw '/api/v1/namespaces/arena-system/services/prometheus-svc:prometheus/proxy/api/v1/query?query=nvidia_gpu_num_devices'

{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"nvidia_gpu_num_devices","app":"node-gpu-exporter","instance":"172.16.1.144:9445","job":"kubernetes-service-endpoints","k8s_app":"node-gpu-exporter","kubernetes_name":"node-gpu-exporter","node_name":"mynode"},"value":[1543202894.919,"2"]}]}}
```

4\. Submit a traing job by arena

```
$ arena submit tf --name=style-transfer \
    --gpus=2 \
    --workers=2 \
    --workerImage=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/neural-style:gpu \
    --workingDir=/neural-style \
    --ps=1 \
    --psImage=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/style-transfer:ps \
    "python neural_style.py --styles /neural-style/examples/1-style.jpg --iterations 1000000"
```


5\. Check GPU metrics for the job you deployed

```
$ arena top job style-transfer
INSTANCE NAME                  STATUS   NODE          GPU(Device Index)  GPU(Duty Cycle)  GPU(Memory MiB)
style-transfer-tfjob-ps-0      Running  192.168.0.95  N/A                N/A              N/A
style-transfer-tfjob-worker-0  Running  192.168.0.98  0                  98%              15641MiB / 16276MiB
                                                    1                  0%               15481MiB / 16276MiB
style-transfer-tfjob-worker-1  Running  192.168.0.99  0                  98%              15641MiB / 16276MiB
                                                    1                  0%               15481MiB / 16276MiB
```
