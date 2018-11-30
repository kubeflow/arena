The command `arena top job <job name>` can display GPU monitoring metrics. Before using it, you must deploy a Prometheus and nodeExporter for GPU Metrics.

1\. Deploy a Prometheus

```
kubectl apply -f kubernetes-artifacts/prometheus/prometheus.yaml
```

2\. Deploy GPU node exporter

If your cluster is ACK (Alibaba Cloud Kubernetes) cluster, you can just exec command:

```
kubectl apply -f kubernetes-artifacts/prometheus/gpu-expoter.yaml
```

If your cluster is not ACK cluster, exec command:

```
# label all your GPU nodes
kubectl label node <your node> accelerator/nvidia_gpu=true
# change gpu export nodeSelector to your label
sed 's|aliyun.accelerator/nvidia_count|accelerator/nvidia_gpu|g' kubernetes-artifacts/prometheus/gpu-expoter.yaml
# deploy gpu expoter
kubectl apply -f kubernetes-artifacts/prometheus/gpu-expoter.yaml
```

> Notice: the prometheus and gpu-exporter components should be deployed in namespace `kube-system`, and `arena top job` can work. 

3\. You can check the GPU metrics by prometheus SQL request

```
# kubectl get --raw '/api/v1/namespaces/kube-system/services/prometheus-svc:prometheus/proxy/api/v1/query?query=nvidia_gpu_num_devices'

{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"nvidia_gpu_num_devices","app":"node-gpu-exporter","instance":"172.16.1.144:9445","job":"kubernetes-service-endpoints","k8s_app":"node-gpu-exporter","kubernetes_name":"node-gpu-exporter","node_name":"mynode"},"value":[1543202894.919,"2"]}]}}

```

4\. Submit a traing job by arena

```
arena submit tf --name=style-transfer              \
              --gpus=2              \
              --workers=2              \
              --workerImage=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/neural-style:gpu \
              --workingDir=/neural-style \
              --ps=1              \
              --psImage=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/style-transfer:ps   \
              "python neural_style.py --styles /neural-style/examples/1-style.jpg --iterations 1000000"
```

5\. Check GPU metrics for the job you deployed

```
# arena top job style-transfer
INSTANCE NAME                  STATUS   NODE          GPU(Device Index)  GPU(Duty Cycle)  GPU(Memory MiB)
style-transfer-tfjob-ps-0      Running  192.168.0.95  N/A                N/A              N/A
style-transfer-tfjob-worker-0  Running  192.168.0.98  0                  98%              15641MiB / 16276MiB
                                                      1                  0%               15481MiB / 16276MiB
style-transfer-tfjob-worker-1  Running  192.168.0.99  0                  98%              15641MiB / 16276MiB
                                                      1                  0%               15481MiB / 16276MiB
```