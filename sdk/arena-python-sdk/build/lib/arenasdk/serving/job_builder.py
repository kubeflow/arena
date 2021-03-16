#!/usr/bin/env python
from __future__ import annotations
import abc 
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.fields.fields import *
from arenasdk.enums.types import ArenaErrorType
from arenasdk.enums.types import ServingJobType
from arenasdk.serving.job import ServingJob
from typing import List
from typing import Dict

def process_items(items: List[str]):
    for item in items:
        print(item) 

class JobBuilder(metaclass=abc.ABCMeta):
    def __init__(self,job_type: ServingJobType):
        self._job_type: ServingJobType = job_type
        self._job_name = ""
        self._version = ""
        self._options: List[Field] = list()
        self._command = ""
        
    def build(self) -> ServingJob:
        args = list()
        try:
            for field in self._options:
                if not isinstance(field,Field):
                    raise ArenaException(ArenaErrorType.Unknown,"the type of option {} is not Field".format(field))
                field.validate()
                for opt in field.options():
                    args.append(opt)
            return ServingJob(self._job_name,self._job_type,self._version,args,self._command)
        except ArenaException as e:
            raise e 
    
    def with_name(self,name: str) -> JobBuilder:
        self._job_name = name
        self._options.append(StringField("--name",name))
        return self 

    def with_image(self,image: str) -> JobBuilder:
        self._options.append(StringField("--image",image))
        return self 

    def with_version(self,version: str) -> JobBuilder:
        self._options.append(StringField("--version",version))
        return self 

    def with_cpu(self,cpu: str) -> JobBuilder:
        self._options.append(StringField("--cpu",cpu))
        return self 

    def with_memory(self,memory: str) -> JobBuilder:
        self._options.append(StringField("--memory",memory))
        return self 

    def with_replicas(self,count: int) -> JobBuilder:
        self._options.append(StringField("--replicas",str(count)))
        return self

    def with_image_pull_policy(self,policy: List[str]) -> JobBuilder:
        self._options.append(StringListField("--image-pull-policy",policy))
        return self
    
    def with_gpus(self,count: int) -> JobBuilder:
        self._options.append(StringField("--gpus",str(count)))
        return self
    
    def with_gpu_memory(self,count: int) -> JobBuilder:
        self._options.append(StringField("--gpumemory",str(count)))
        return self

    def with_envs(self,envs: Dict[str,str]) -> JobBuilder:
        self._options.append(StringMapField("--env",envs,"="))
        return self

    def with_node_selectors(self,selectors: Dict[str, str]) -> JobBuilder:
        self._options.append(StringMapField("--selector",selectors,"="))
        return self

    def with_tolerations(self,tolerations: List[str]) -> JobBuilder:
        self._options.append(StringListField("--toleration",tolerations))
        return self
    
    def with_annotations(self,annotions: Dict[str, str]) -> JobBuilder:
        self._options.append(StringMapField("--annotation",annotions,"="))
        return self

    def with_datas(self,datas: Dict[str,str]) -> JobBuilder:
        self._options.append(StringMapField("--data",datas,":"))
        return self 
    
    def with_data_dirs(self,data_dirs: Dict[str, str]) -> JobBuilder:
        self._options.append(StringMapField("--data-dir",data_dirs,":"))
        return  self

    def with_command(self,command: str) -> JobBuilder:
        self._command = command
        return self