#!/usr/bin/env python
from __future__ import annotations
from arenasdk.nodes.node import Node
from arenasdk.nodes.node import AdvancedGPUMetric
from typing import List
from typing import Dict

class GPUShareNode(Node):
    def __init__(self):
        super().__init__()
        self._total_gpus: int
        self._allocated_gpus: int  
        self._unhealthy_gpus: int
        self._total_gpu_memory: int
        self._allocated_gpu_memory: int
        self._instances: List[GPUShareNodePod]
        self._devices: List[GPUShareNodeDevice]
        self._gpu_metrics: List[AdvancedGPUMetric]
    
    def set_total_gpus(self, count: int) -> None:
        self._total_gpus = count
    
    def get_total_gpus(self) -> int:
        return self._total_gpus
    
    def set_allocated_gpus(self,count: int) -> None:
        self._allocated_gpus = count
    
    def get_allocated_gpus(self) -> int:
        return self._allocated_gpus
    
    def set_total_gpu_memory(self,count: int) -> None:
        self._total_gpu_memory = count
    
    def get_total_gpu_memory(self) -> int:
        return self._total_gpu_memory
    
    def set_allocated_gpu_memory(self,count: int) -> None:
        self._allocated_gpu_memory = count
    
    def get_allocated_gpu_memory(self) -> int:
        return self._allocated_gpu_memory
    
    def set_unhealthy_gpus(self,count: int) -> None:
        self._unhealthy_gpus = count
    
    def get_unhealthy_gpus(self) -> int:
        return self._unhealthy_gpus

    def set_instances(self,instances: List[GPUShareNodePod]) -> None:
        self._instances = instances
    
    def get_instances(self) -> List[GPUShareNodePod]:
        return self._instances
    
    def set_devices(self,devices: List[GPUShareNodeDevice]) -> None:
        self._devices = devices
    
    def get_devices(self) -> List[GPUShareNodeDevice]:
        return self._devices
    
    def set_gpu_metrics(self,metrics: List[AdvancedGPUMetric]) -> None:
        self._metrics = metrics
    
    def get_gpu_metrics(self) -> List[AdvancedGPUMetric]:
        return self._metrics
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()

def build_gpushare_nodes(data: dict) -> GPUShareNode:
    cls = GPUShareNode()
    cls.set_total_gpus(data["totalGPUs"])
    cls.set_allocated_gpus(data["allocatedGPUs"])
    cls.set_unhealthy_gpus(data["unhealthyGPUs"])
    cls.set_total_gpu_memory(data["totalGPUMemory"])
    cls.set_allocated_gpu_memory(data["allocatedGPUMemory"])
    cls.set_name(data["name"])
    cls.set_ip(data["ip"])
    cls.set_status(data["status"])
    cls.set_role(data["role"])
    cls.set_node_type(data["type"])
    devices = list()
    for d in data["devices"]:
        device = GPUShareNodeDevice()
        device.set_id(d["id"])
        device.set_allocated_gpu_memory(d["allocatedGPUMemory"])
        device.set_total_gpu_memory(d["totalGPUMemory"])
        devices.append(device)
    cls.set_devices(devices)
    instances = list()
    for i in data["instances"]:
        instance = GPUShareNodePod()
        instance.set_name(i["name"])
        instance.set_namespace(i["namespace"])
        instance.set_request_gpu_memory(i["requestGPUMemory"])
        instance.set_allocation(i["allocation"])
        instances.append(instance)
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
    cls.set_instances(instances)
    return cls 
    

class GPUShareNodePod(object):
    def __init__(self): 
        self._name: str
        self._namespace: str
        self._request_gpu_memory: int 
        self._allocation: Dict[str,int]
    def set_name(self, name: str) -> None:
        self._name = name
        
    def get_name(self) -> str:
        return self._name
    
    def set_namespace(self, namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace
    
    def set_request_gpu_memory(self, request_gpu_memory: int) -> None:
        self._request_gpu_memory = request_gpu_memory
    
    def get_request_gpu_memory(self) -> int:
        return self._request_gpu_memory
    
    def set_allocation(self, allocation: Dict[str,int]) -> None:
        self._allocation = allocation
    
    def get_allocation(self) -> Dict[str,int]:
        return self._allocation
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()

class GPUShareNodeDevice(object):
    def __init__(self):
        self._id: str
        self._allocated_gpu_memory: int 
        self._total_gpu_memory: int
    
    def set_id(self, id: str) -> None:
        self._id = id
    
    def get_id(self) -> str:
        return self._id
    
    def set_allocated_gpu_memory(self,allocated_gpu_memory: int) -> None:
        self._allocated_gpu_memory = allocated_gpu_memory
    
    def get_allocated_gpu_memory(self) -> int:
        return self._allocated_gpu_memory
    
    def set_total_gpu_memory(self,total_gpu_memory: int) -> None:
        self._total_gpu_memory = total_gpu_memory
    
    def get_total_gpu_memory(self) -> int:
        return self._total_gpu_memory
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()