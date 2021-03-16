#!/usr/bin/env python
from __future__ import annotations
import abc 
from arenasdk.enums.types import *

class Node(metaclass=abc.ABCMeta):
    def __init__(self):
        self._name: str
        self._ip: str 
        self._status: str
        self._role: str
        self._node_type: NodeType
        
    def set_name(self,name: str) -> None:
        self._name = name
        
    def get_name(self) -> str:
        return self._name
    
    def set_ip(self,ip: str) -> None:
        self._ip = ip
        
    def get_ip(self) -> str:
        return self._ip
    
    def set_status(self,status: str) -> None:
        self._status = status
        
    def get_status(self) -> str:
        return self._status
    
    def set_role(self,role: str) -> None:
        self._role = role
        
    def get_role(self) -> str:
        return self._role
    
    def set_node_type(self,node_type: str) -> None:
        self._node_type = NodeType.convert(node_type)
    
    def get_node_type(self) -> NodeType:
        return self._node_type


class AdvancedGPUMetric(object):
    def __init__(self):
        self._id: str
        self._uuid: str
        self._status: str 
        self._gpu_duty_cycle: float
        self._used_gpu_memory: float
        self._total_gpu_memory: float
    
    def set_id(self,id: str) -> None:
        self._id = id
    
    def get_id(self) -> str:
        return self._id
    
    def set_uuid(self,uuid: str) -> None:
        self._uuid = uuid
    
    def get_uuid(self) -> str:
        return self._uuid
    
    def set_status(self,status: str) -> None:
        self._status = status
        
    def get_status(self) -> str:
        return self._status
     
    def set_gpu_duty_cycle(self,gpu_duty_cycle: float) -> None:
        self._gpu_duty_cycle = gpu_duty_cycle
        
    def get_gpu_duty_cycle(self) -> float:
        return self._gpu_duty_cycle

    def set_used_gpu_memory(self,used_gpu_memory: float) -> None:
        self._used_gpu_memory = used_gpu_memory
        
    def get_used_gpu_memory(self) -> float:
        return self._used_gpu_memory
    
    def set_total_gpu_memory(self,total_gpu_memory: float) -> None:
        self._total_gpu_memory = total_gpu_memory
    
    def get_total_gpu_memory(self) -> float:
        return self._total_gpu_memory 