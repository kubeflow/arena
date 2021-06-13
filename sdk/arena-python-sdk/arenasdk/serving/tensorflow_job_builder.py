#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import ServingJobType
from arenasdk.fields.fields import *
from arenasdk.serving.job_builder import JobBuilder

class TFServingJobBuilder(JobBuilder):
    def __init__(self):
        super().__init__(ServingJobType.TFServingJob)

    def with_restful_port(self,port: int) -> TFServingJobBuilder:
        self._options.append(StringField("--restful-port",str(port)))
        return self
    
    def with_port(self,port: int) -> TFServingJobBuilder:
        self._options.append(StringField("--port",str(port)))
        return self

    def with_version_policy(self,policy: str) -> TFServingJobBuilder:
        self._options.append(StringField("--version-policy",policy))
        return self

    def with_version_policy(self,policy: str) -> TFServingJobBuilder:
        self._options.append(StringField("--version-policy",policy))
        return self

    def with_model_name(self,name: str) -> TFServingJobBuilder:
        self._options.append(StringField("--model-name",name))
        return self
    
    def with_model_path(self,path: str) -> TFServingJobBuilder:
        self._options.append(StringField("--model-path",path))
        return self

    def with_model_config_file(self,file: str) -> TFServingJobBuilder:
        self._options.append(StringField("--model-config-file",file))
        return self

    #         
    def with_name(self,name: str) -> TFServingJobBuilder:
        self._job_name = name
        super().with_name(name)
        return self 

    def with_image(self,image: str) -> TFServingJobBuilder:
        super().with_image(image)
        return self 

    def with_version(self,version: str) -> TFServingJobBuilder:
        super().with_version(version)
        return self 

    def with_cpu(self,cpu: str) -> TFServingJobBuilder:
        super().with_cpu(cpu)
        return self 

    def with_memory(self,memory: str) -> TFServingJobBuilder:
        super().with_memory(memory)
        return self 

    def with_replicas(self,count: int) -> TFServingJobBuilder:
        super().with_replicas(count)
        return self

    def with_image_pull_policy(self,policy: List[str]) -> TFServingJobBuilder:
        super().with_image_pull_policy(policy)
        return self
    
    def with_gpus(self,count: int) -> TFServingJobBuilder:
        super().with_gpus(count)
        return self
    
    def with_gpu_memory(self,count: int) -> TFServingJobBuilder:
        super().with_gpu_memory(count)
        return self

    def with_envs(self,envs: Dict[str,str]) -> TFServingJobBuilder:
        super().env(envs)
        return self

    def with_node_selectors(self,selectors: Dict[str, str]) -> TFServingJobBuilder:
        super().with_node_selectors(selectors)
        return self

    def with_tolerations(self,tolerations: List[str]) -> TFServingJobBuilder:
        super().with_tolerations(tolerations)
        return self
    
    def with_annotations(self,annotions: Dict[str, str]) -> TFServingJobBuilder:
        super().with_annotations(annotions)
        return self

    def with_datas(self,datas: Dict[str,str]) -> TFServingJobBuilder:
        super().with_datas(datas)
        return self 
    
    def with_data_dirs(self,data_dirs: Dict[str, str]) -> TFServingJobBuilder:
        super().with_data_dirs(data_dirs)
        return  self

    def with_command(self,command: str) -> TFServingJobBuilder:
        super().with_command(command)
        return self
