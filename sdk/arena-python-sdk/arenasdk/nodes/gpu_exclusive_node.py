#!/usr/bin/env python
from __future__ import annotations
from arenasdk.nodes.node import Node
from arenasdk.nodes.node import AdvancedGPUMetric
from typing import List
from typing import Dict 

class GPUExclusiveNode(Node):
    def __init__(self):
        super().__init__()
        self._total_gpus: int
        self._allocated_gpus: int 
        self._unhealthy_gpus: int
        self._instances: List[GPUExclusiveNodePod]
        self._gpu_metrics: List[AdvancedGPUMetric]
    
    def set_total_gpus(self,total_gpus: int) -> None:
        self._total_gpus = total_gpus
    
    def get_total_gpus(self) -> int:
        return self._total_gpus
        
    def set_allocated_gpus(self,used_gpus: int) -> None:
        self._allocated_gpus = used_gpus
    
    def get_allocated_gpus(self) -> int:
        return self._allocated_gpus
        
    def set_unhealthy_gpus(self,unhealthy_gpus: int) -> None:
        self._unhealthy_gpus = unhealthy_gpus
        
    def get_unhealthy_gpus(self) -> int:
        return self._unhealthy_gpus
    
    def set_instances(self,instances: List[GPUExclusiveNodePod]) -> None:
        self._instances = instances
    
    def get_instances(self) -> List[GPUExclusiveNodePod]:
        return self._instances
    
    def set_gpu_metrics(self,metrics: List[AdvancedGPUMetric]) -> None:
        self._metrics = metrics
    
    def get_gpu_metrics(self) -> List[AdvancedGPUMetric]:
        return self._metrics
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()
    
def build_gpu_exclusive_nodes(data: dict) -> GPUExclusiveNode:
    cls = GPUExclusiveNode()
    cls.set_name(data["name"])
    cls.set_ip(data["ip"])
    cls.set_status(data["status"])
    cls.set_role(data["role"])
    cls.set_node_type(data["type"])
    cls.set_total_gpus(data["totalGPUs"])
    cls.set_allocated_gpus(data["allocatedGPUs"])
    cls.set_unhealthy_gpus(data["unhealthyGPUs"])
    instances = list()
    for i in data["instances"]:
        instance = GPUExclusiveNodePod()
        instance.set_name(i["name"])
        instance.set_namespace(i["namespace"])
        instance.set_request_gpus(i["requestGPUs"])
        instances.append(instance)
    cls.set_instances(instances)
    gpu_metrics = list()
    for g in data["gpuMetrics"]:
        gpu_metric = AdvancedGPUMetric()
        gpu_metric.set_id(g["id"])
        gpu_metric.set_uuid(g["uuid"])
        gpu_metric.set_status(g["status"])
        gpu_metric.set_gpu_duty_cycle(g["gpuDutyCycle"])
        gpu_metric.set_total_gpu_memory(g["totalGPUMemory"])
        gpu_metric.set_used_gpu_memory(g["usedGPUMemory"])
        gpu_metrics.append(gpu_metric)
    cls.set_gpu_metrics(gpu_metrics)
    return cls


class GPUExclusiveNodePod(object):
    def __init__(self):
        self._name: str
        self._namepsace: str
        self._request_gpus: int
    
    def set_name(self,name: str) -> None:
        self._name = name
    
    def get_name(self) -> str:
        return self._name
    
    def set_namespace(self,namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace
    
    def set_request_gpus(self,gpus: int) -> None:
        self._request_gpus = gpus
    
    def get_request_gpus(self) -> int:
        return self._request_gpus
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()
        