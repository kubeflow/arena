#!/usr/bin/env python
from __future__ import annotations
from arenasdk.nodes.node import Node
from arenasdk.nodes.node import AdvancedGPUMetric
from typing import List
from typing import Dict

class GPUTopologyNode(Node):
    def __init__(self):
        super().__init__()
        self._total_gpus: int
        self._allocated_gpus: int 
        self._unhealthy_gpus: int
        self._instances: List[GPUTopologyNodePod]
        self._linketype_matrix: List[List[str]]
        self._bandwidth_matrix: List[List[float]]
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
    
    def set_instances(self,instances: List[GPUTopologyNodePod]) -> None:
        self._instances = instances
    
    def get_instances(self) -> List[GPUTopologyNodePod]:
        return self._instances
    
    def set_devices(self,devices: List[PGUTopologyNodeDevice]) -> None:
        self._devices = devices
    
    def get_devices(self) -> List[PGUTopologyNodeDevice]:
        return self._devices 
    
    def set_linktype_matrix(self,matrix:  List[List[str]]) -> None:
        self._linktype_matrix = matrix
    
    def get_linktype_matrix(self) -> List[List[str]]:
        return self._linktype_matrix
    
    def set_bandwidth_matrix(self,matrix: List[List[float]]) -> None:
        self._bandwidth_matrix = matrix
    
    def get_bandwidth_matrix(self) -> List[List[float]]:
        return self._bandwidth_matrix
    
    def set_gpu_metrics(self,metrics: List[AdvancedGPUMetric]) -> None:
        self._metrics = metrics
    
    def get_gpu_metrics(self) -> List[AdvancedGPUMetric]:
        return self._metrics
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()

def build_gpu_topology_nodes(data: dict) -> GPUTopologyNode:
    cls = GPUTopologyNode()
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
        instance = GPUTopologyNodePod()
        instance.set_name(i["name"])
        instance.set_namespace(i["namespace"])
        instance.set_request_gpus(i["requestGPUs"])
        instance.set_allocation(i["allocation"])
        instance.set_visible_gpus(i["visibleGPUs"])
        instances.append(instance)
    cls.set_instances(instances)
    devices = list()
    for d in data["devices"]:
        device = GPUTopologyNodeDevice()
        device.set_id(d["gpuIndex"])
        device.set_status(d["status"])
        device.set_healthy(d["healthy"])
        devices.append(device)
    cls.set_devices(devices)
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
    cls.set_linktype_matrix(data["gpuTopology"]["linkMatrix"])
    cls.set_bandwidth_matrix(data["gpuTopology"]["bandwidthMatrix"])
    return cls
    

class GPUTopologyNodePod(object):
    def __init__(self):
        self._name: str
        self._namespace: str
        self._request_gpus: int 
        self._allocation: List[str]
        self._visible_gpus: List[str]
    
    def set_name(self, name: str) -> None:
        self._name = name
    
    def get_name(self) -> str:
        return self._name
    
    def set_namespace(self, namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace
    
    def set_request_gpus(self, request_gpus: int) -> None:
        self._request_gpus = request_gpus
    
    def get_request_gpus(self) -> int:
        return self._request_gpus
    
    def set_allocation(self,allocation: List[str]) -> None:
        self._allocation = allocation
    
    def get_allocation(self) -> List[str]:
        return self._allocation
    
    def set_visible_gpus(self,visible_gpus: List[str]) -> None:
        self._visible_gpus = visible_gpus
    
    def get_visible_gpus(self) -> List[str]:
        return self._visible_gpus
    
    def to_dict(self) -> dict:
        return self.__dict__.copy() 


class GPUTopologyNodeDevice(object):
    def __init__(self):
        self._id: str
        self._status: str
        self._healthy: bool
    
    def set_id(self, id: str) -> None:
        self._id = id
    
    def get_id(self) -> str:
        return self._id
    
    def set_status(self,status: str) -> None:
        self._status = status
    
    def get_status(self) -> str:
        return self._status
    
    def set_healthy(self,healthy: bool) -> None:
        self._healthy = healthy
    
    def is_healthy(self) -> bool:
        return self._healthy
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()
    