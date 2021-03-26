#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import ServingJobType
from arenasdk.fields.fields import *
from arenasdk.serving.job_builder import JobBuilder

class CustomServingJobBuilder(JobBuilder):
    def __init__(self):
        super().__init__(ServingJobType.CustomServingJob)

    def with_restful_port(self,port: int) -> CustomServingJobBuilder:
        self._options.append(StringField("--restful-port",str(port)))
        return self
    
    def with_port(self,port: int) -> CustomServingJobBuilder:
        self._options.append(StringField("--port",str(port)))
        return self

    #         
    def with_name(self,name: str) -> CustomServingJobBuilder:
        self._job_name = name
        super().with_name(name)
        return self 

    def with_image(self,image: str) -> CustomServingJobBuilder:
        super().with_image(image)
        return self 

    def with_version(self,version: str) -> CustomServingJobBuilder:
        super().with_version(version)
        return self 

    def with_cpu(self,cpu: str) -> CustomServingJobBuilder:
        super().with_cpu(cpu)
        return self 

    def with_memory(self,memory: str) -> CustomServingJobBuilder:
        super().with_memory(memory)
        return self 

    def with_replicas(self,count: int) -> CustomServingJobBuilder:
        super().with_replicas(count)
        return self

    def with_image_pull_policy(self,policy: List[str]) -> CustomServingJobBuilder:
        super().with_image_pull_policy(policy)
        return self
    
    def with_gpus(self,count: int) -> CustomServingJobBuilder:
        super().with_gpus(count)
        return self
    
    def with_gpu_memory(self,count: int) -> CustomServingJobBuilder:
        super().with_gpu_memory(count)
        return self

    def with_envs(self,envs: Dict[str,str]) -> CustomServingJobBuilder:
        super().env(envs)
        return self

    def with_node_selectors(self,selectors: Dict[str, str]) -> CustomServingJobBuilder:
        super().with_node_selectors(selectors)
        return self

    def with_tolerations(self,tolerations: List[str]) -> CustomServingJobBuilder:
        super().with_tolerations(tolerations)
        return self
    
    def with_annotations(self,annotions: Dict[str, str]) -> CustomServingJobBuilder:
        super().with_annotations(annotions)
        return self

    def with_datas(self,datas: Dict[str,str]) -> CustomServingJobBuilder:
        super().with_datas(datas)
        return self 
    
    def with_data_dirs(self,data_dirs: Dict[str, str]) -> CustomServingJobBuilder:
        super().with_data_dirs(data_dirs)
        return  self

    def with_command(self,command: str) -> CustomServingJobBuilder:
        super().with_command(command)
        return self
