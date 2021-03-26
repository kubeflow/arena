#!/usr/bin/env python
from __future__ import annotations
import os
import json
from typing import List
from typing import Dict
from arenasdk.enums.types import *
from arenasdk.logger.logger import LoggerBuilder
from arenasdk.common.util import Command
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.common.log import Log

logger = Log(__name__).get_logger()

class ServingJobInfo(object):
    def __init__(self):
        self._name: str
        self._namespace: str
        self._job_type: ServingJobType 
        self._version: str 
        self._duration: str
        self._desired_instances: int
        self._available_instances: int   
        self._ip: str
        self._request_gpus: int
        self._request_gpu_memory: int
        self._instances: List[Instance]
        self._endpoints: List[Endpoint]
    
    def set_name(self,name: str) -> None:
        self._name = name
        
    def get_name(self) -> str:
        return self._name
    
    def set_namespace(self,namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace
    
    def set_version(self,version: str) -> None:
        self._version = version
    
    def get_version(self) -> str:
        return self._version
    
    def set_desired_instances(self,count: int) -> None:
        self._desired_instances = count
    
    def get_desired_instances(self) -> int:
        return self._desired_instances
    
    def set_available_instances(self,count: int) -> None:
        self._available_instances = count
    
    def get_available_instances(self) -> int:
        return self._available_instances

    def set_duration(self,duration: str) -> None:
        self._duration = duration
    
    def get_duration(self) -> str:
        return self._duration

    def set_type(self,job_type: str) -> None:
        self._job_type = ServingJobType.convert(job_type)
    
    def get_type(self) -> ServingJobType:
        return self._job_type
    
    def set_request_gpus(self,request_gpus: int) -> None:
        self._request_gpus = request_gpus
    
    def get_request_gpus(self) -> int:
        return self._request_gpus
    
    def set_instances(self,instances: List[Instance]) -> None:
        self._instances = instances
    
    def get_instances(self) -> List[Instance]:
        return self._instances
    
    def set_endpoints(self,endpoints: List[Endpoint]) -> None:
        self._endpoints = endpoints
    
    def get_endpoints(self) -> List[Endpoint]:
        return self._endpoints
    
    def set_request_gpu_memory(self,gpu_memory: int) -> None:
        self._request_gpu_memory = gpu_memory

    def get_request_gpu_memory(self) -> int:
        return self._request_gpu_memory
    
    def set_ip(self,ip: str) -> None:
        self._ip = ip
    
    def get_ip(self) -> str:
        return self._ip
    
    def __str__(self) -> str:
        data = dict()
        data["name"] = self.get_name()
        data["namespace"] = self.get_namespace()
        data["duration"] = self.get_duration()
        data["type"] = self.get_type().value[0]
        data["version"] = self.get_version()
        data["desired_instances"] = self.get_desired_instances()
        data["available_instances"] = self.get_available_instances()
        data["ip"] = self.get_ip()
        data["request_gpus"] = self.get_request_gpus()
        data["request_gpu_memory"] = self.get_request_gpu_memory()
        endpoints = list()
        for e in self.get_endpoints():
            endpoint = dict()
            endpoint["name"] = e.get_name()
            endpoint["port"] =  e.get_port()
            endpoint["node_port"] = e.get_node_port()
            endpoints.append(endpoint)
        data["endpoints"] = endpoints
        instances = list()
        for i in self.get_instances():
            instance = dict()
            instance["name"] = i.get_name()
            instance["age"] = i.get_age()
            instance["namespace"] = i.get_namespace()
            instance["status"] = i.get_status()
            instance["ready_containers"] = i.get_ready_containers()
            instance["total_containers"] = i.get_total_containers()
            instance["restart_count"] = i.get_restart_count()
            instance["ip"] = i.get_ip()
            instance["node_ip"]  = i.get_node_ip()
            instance["node_name"] = i.get_node_name()
            instance["request_gpus"] = i.get_request_gpus()
            instance["request_gpu_memory"] = i.get_request_gpu_memory()
            instances.append(instance)
        data["instances"] = instances
        return json.dumps(data, sort_keys=True, indent=4) 

class Instance(object):
    def __init__(self):
        self._owner: str 
        self._owner_type: ServingJobType
        self._owner_version: str
        self._name: str
        self._age: str
        self._namespace: str
        self._status: str
        self._ready_containers: int
        self._total_containers: int
        self._restart: int
        self._node_name: str
        self._node_ip: str
        self._ip: str     
        self._request_gpus: int
        self._request_gpu_memory: int
    
    def set_owner(self, owner: str) -> None:
        self._owner = owner
        
    def get_owner(self) -> str:
        return self._owner
    
    def set_owner_type(self, owner_type: ServingJobType) -> None:
        self._owner_type = owner_type
    
    def get_owner_type(self) -> ServingJobType:
        return self._owner_type
    
    def set_owner_version(self,version: str) -> None:
        self._owner_version = version
    
    def get_owner_version(self) -> str:
        return self._owner_version
    
    def set_name(self,name: str) -> None:
        self._name = name
    
    def get_name(self) -> str:
        return self._name
    
    def set_age(self,age: str) -> None:
        self._age = age 

    def get_age(self) -> str:
        return self._age

    def set_namespace(self,namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace
    
    def set_status(self,status: str) -> None:
        self._status = status
    
    def get_status(self) -> str:
        return self._status

    def set_ready_containers(self,count: int) -> None:
        self._ready_containers = count
    
    def get_ready_containers(self) -> int:
        return self._ready_containers
    
    def set_total_containers(self,count: int) -> None:
        self._total_containers = count
    
    def get_total_containers(self) -> int:
        return self._total_containers
    
    def set_restart_count(self,count: int) -> None:
        self._restart = count
    
    def get_restart_count(self) -> int:
        return self._restart
    
    def set_ip(self,ip: str) -> None:
        self._ip = ip
    
    def get_ip(self) -> str:
        return self._ip
    
    def set_node_name(self,node_name: str) -> None:
        self._node = node_name
    
    def get_node_name(self) -> str:
        return self._node
    
    def set_node_ip(self,ip: str) -> None:
        self._node_ip = ip
    
    def get_node_ip(self) -> str:
        return self._node_ip
    
    def set_request_gpus(self,gpus: int) -> None:
        self._request_gpus = gpus 
    
    def get_request_gpus(self) -> int:
        return self._request_gpus
    
    def set_request_gpu_memory(self,gpu_memory: int) -> None:
        self._request_gpu_memory = gpu_memory
    
    def get_request_gpu_memory(self) -> int:
        return self._request_gpu_memory
    
    def get_logs(self,logger: LoggerBuilder) -> None:
        cmds = list()
        cmds.append(ARENA_BINARY)
        cmds.append("serve")
        cmds.append("logs")
        cmds.append(self.get_owner())
        cmds.append("-T=" + self.get_owner_type().value[0])
        cmds.append("--version=" + self.get_owner_version())
        for opt in logger.get_args():
            cmds.append(opt)
        cmds.append("-i=" + self._name)
        kubeconfig = os.getenv("KUBECONFIG")
        if kubeconfig and kubeconfig != "":
            cmds.append("--config=" + kubeconfig)
        cmds.append("--namespace=" + self._namespace)
        arena_namespace = os.getenv("ARENA_NAMESPACE")
        if arena_namespace and arena_namespace != "":
            cmds.append("--arena-namespace=" + arena_namespace)
        try:
            status,stdout,stderr = Command(*cmds).run_with_communicate(logger.get_accepter())
            if status != 0:
                raise ArenaException(ArenaErrorType.LogsServingJobError,stdout + stderr)
        except ArenaException as e:
            raise e 
        
class Endpoint(object):
    def __init__(self):
        self._name: str
        self._port: int
        self._node_port: int
    
    def set_name(self,name: str) -> None:
        self._name = name
    
    def get_name(self) -> str:
        return self._name
    
    def set_port(self,port:int) -> None:
        self._port = port
    
    def get_port(self) -> None:
        return self._port
    
    def set_node_port(self,port: int) -> None:
        self._node_port = port
    
    def get_node_port(self) -> None:
        return self._node_port
    
def generate_serving_job_info(data) -> ServingJobInfo:
    logger.debug("get serving job information with json: %s",data)
    job_info = ServingJobInfo()
    job_info.set_name(data["name"])
    job_info.set_namespace(data["namespace"])
    job_info.set_duration(data["age"])
    job_info.set_type(data["type"])
    job_info.set_version(data["version"])
    job_info.set_desired_instances(data["desiredInstances"])
    job_info.set_available_instances(data["availableInstances"])
    job_info.set_ip(data["ip"])
    job_info.set_request_gpus(data["requestGPUs"])
    job_info.set_request_gpu_memory(data["requestGPUMemory"])
    instances = list()
    for i in data["instances"]:
        instance = Instance()
        instance.set_owner(job_info.get_name())
        instance.set_owner_type(job_info.get_type())
        instance.set_owner_version(job_info.get_version())
        instance.set_namespace(job_info.get_namespace())
        instance.set_name(i["name"])
        instance.set_status(i["status"])
        instance.set_age(i["age"])
        instance.set_node_name(i["nodeName"])
        instance.set_node_ip(i["nodeIP"])
        instance.set_ip(i["ip"])
        instance.set_ready_containers(i["readyContainers"])
        instance.set_total_containers(i["totalContainers"])
        instance.set_restart_count(i["restartCount"])
        instance.set_request_gpus(i["requestGPUs"])
        instance.set_request_gpu_memory(i["requestGPUMemory"])
        instances.append(instance)
    job_info.set_instances(instances)
    endpoints = list()
    for e in data["endpoints"]:
        endpoint = Endpoint()
        endpoint.set_name(e["name"])
        endpoint.set_port(e["port"])
        endpoint.set_node_port(e["nodePort"])
        endpoints.append(endpoint)
    job_info.set_endpoints(endpoints)
    return job_info
