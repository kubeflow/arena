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

class TrainingJobInfo(object):
    def __init__(self):
        self._name: str
        self._namespace: str
        self._duration: str 
        self._status: TrainingJobStatus
        self._job_type: TrainingJobType 
        self._tensorboard: str
        self._chief_name: str
        self._priority: str
        self._request_gpus: int
        self._allocated_gpus: int
        self._instances: List[Instance]
    
    def set_name(self,name: str) -> None:
        self._name = name
        
    def get_name(self) -> str:
        return self._name
    
    def set_namespace(self,namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace

    def set_duration(self,duration: str) -> None:
        self._duration = duration
    
    def get_duration(self) -> str:
        return self._duration

    def set_status(self,status: str) -> None:
        self._status = TrainingJobStatus.convert(status)
    
    def get_status(self) -> TrainingJobStatus:
        return self._status
    
    def set_type(self,job_type: str) -> None:
        self._job_type = TrainingJobType.convert(job_type)
    
    def get_type(self) -> TrainingJobType:
        return self._job_type
    
    def set_tensorboard(self,tensorboard: str) -> None:
        self._tensorboard = tensorboard
    
    def get_tensorboard(self) -> str:
        return self._tensorboard
    
    def set_chief_name(self,chief_name: str) -> None:
        self._chief_name = chief_name
    
    def get_chief_name(self) -> str:
        return self._chief_name
    
    def set_priority(self,priority: str) -> None:
        self._priority = priority
    
    def get_priority(self) -> str:
        return self._priority
    
    def set_request_gpus(self,request_gpus: int) -> None:
        self._request_gpus = request_gpus
    
    def get_request_gpus(self) -> int:
        return self._request_gpus
    
    def set_allocated_gpus(self,allocated_gpus: int) -> None:
        self._allocated_gpus = allocated_gpus

    def get_allocated_gpus(self) -> int:
        return self._allocated_gpus
    
    def set_instances(self,instances: List[Instance]) -> None:
        self._instances = instances
    
    def get_instances(self) -> List[Instance]:
        return self._instances
    
    def __str__(self) -> str:
        data = dict()
        data["name"] = self.get_name()
        data["namespace"] = self.get_namespace()
        data["duration"] = self.get_duration()
        data["type"] = self.get_type().value
        data["tensorboard"] = self.get_tensorboard()
        data["chief_name"] = self.get_chief_name()
        data["priority"] = self.get_priority()
        data["request_gpus"] = self.get_request_gpus()
        data["allocated_gpus"] = self.get_allocated_gpus()
        instances = list()
        for instance in self._instances:
            instance_data = dict()
            instance_data["owner"] = instance.get_owner()
            instance_data["owner_type"] = instance.get_owner_type().value
            instance_data["name"] = instance.get_name()
            instance_data["age"] = instance.get_age()
            instance_data["status"] = instance.get_status()
            instance_data["node_name"] = instance.get_node_name()
            instance_data["node_ip"] = instance.get_node_ip()
            instance_data["request_gpus"] = instance.get_request_gpus()
            instance_data["is_chief"] = instance.is_chief()
            metrics = list()
            for metric in instance.get_gpu_metrics():
                metric_data = dict()
                metric_data["gpu_duty_cycle"] = metric.get_gpu_duty_cycle()
                metric_data["gpu_used_memory"] = metric.get_used_gpu_memory()
                metric_data["gpu_total_memory"] = metric.get_gpu_total_memory()
                metrics.append(metric_data)
            instance_data["gpu_metrics"] = metrics
            instances.append(instance_data)
        data["instances"] = instances
        return json.dumps(data, sort_keys=True, indent=4)
            
class Instance(object):
    def __init__(self):
        self._owner: str 
        self._owner_type: TrainingJobType
        self._name: str
        self._age: str
        self._status: str
        self._node: str
        self._node_ip: str
        self._namespace: str
        self._request_gpus: int
        self._is_chief: bool 
        self._gpu_metrics: Dict[str,GPUMetric]
    
    def set_owner(self, owner: str) -> None:
        self._owner = owner
        
    def get_owner(self) -> str:
        return self._owner
    
    def set_owner_type(self, owner_type: TrainingJobType) -> None:
        self._owner_type = owner_type
    
    def get_owner_type(self) -> TrainingJobType:
        return self._owner_type
    
    def set_name(self,name: str) -> None:
        self._name = name
    
    def get_name(self) -> str:
        return self._name
    
    def set_age(self,age: str) -> None:
        self._age = age 
    
    def get_age(self) -> str:
        return self._age
    
    def set_status(self,status: str) -> None:
        self._status = status
    
    def get_status(self) -> str:
        return self._status
    
    def set_node_name(self,node_name: str) -> None:
        self._node = node_name
    
    def get_node_name(self) -> str:
        return self._node
    
    def set_node_ip(self,ip: str) -> None:
        self._node_ip = ip
    
    def get_node_ip(self) -> str:
        return self._node_ip
    
    def set_namespace(self,namespace: str) -> None:
        self._namespace = namespace
    
    def get_namespace(self) -> str:
        return self._namespace

    def set_request_gpus(self,gpus: int) -> None:
        self._request_gpus = gpus 
    
    def get_request_gpus(self) -> int:
        return self._request_gpus
    
    def set_is_chief(self,is_chief: bool) -> None:
        self._is_chief = is_chief
    
    def is_chief(self) -> bool:
        return self._is_chief
    
    def set_gpu_metrics(self,metrics: Dict[str,GPUMetric]) -> None:
        self._gpu_metrics = metrics
    
    def get_gpu_metrics(self) -> Dict[str,GPUMetric]:
        return self._gpu_metrics

    def get_logs(self,logger: LoggerBuilder) -> None:
        cmds = list()
        cmds.append(ARENA_BINARY)
        cmds.append("logs")
        cmds.append(self.get_owner())
        cmds.append("-T=" + self.get_owner_type().value)
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
                raise ArenaException(ArenaErrorType.LogsTrainingJobError,stdout + stderr)
        except ArenaException as e:
            raise e  

class GPUMetric(object):
    def __init__(self):
        self._gpu_duty_cycle: float
        self._used_gpu_memory: float
        self._total_gpu_memory: float
    
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
    
def generate_training_job_info(data) -> TrainingJobInfo:
    # logger.debug("get training job information with json: %s",data)
    job_info = TrainingJobInfo()
    job_info.set_name(data["name"])
    job_info.set_namespace(data["namespace"])
    job_info.set_duration(data["duration"])
    job_info.set_status(data["status"])
    job_info.set_type(data["trainer"])
    job_info.set_tensorboard(data["tensorboard"])
    job_info.set_chief_name(data["chiefName"])
    job_info.set_priority(data["priority"])
    job_info.set_request_gpus(data["requestGPUs"])
    job_info.set_allocated_gpus(data["allocatedGPUs"])
    instances = list()
    for i in data["instances"]:
        instance = Instance()
        instance.set_owner(job_info.get_name())
        instance.set_owner_type(job_info.get_type())
        instance.set_namespace(job_info.get_namespace())
        instance.set_name(i["name"])
        instance.set_status(i["status"])
        instance.set_age(i["age"])
        instance.set_node_name(i["node"])
        instance.set_node_ip(i["nodeIP"])
        instance.set_is_chief(i["chief"])
        instance.set_request_gpus(i["requestGPUs"])
        gpu_metrics = dict()
        for key,value in i["gpuMetrics"].items():
            gpu_metric = GPUMetric()
            gpu_metric.set_gpu_duty_cycle(value["gpuDutyCycle"])
            gpu_metric.set_used_gpu_memory(value["usedGPUMemory"])
            gpu_metric.set_total_gpu_memory(value["totalGPUMemory"])
            gpu_metrics[key] = gpu_metric
        instance.set_gpu_metrics(gpu_metrics)
        instances.append(instance)
    job_info.set_instances(instances)
    return job_info
