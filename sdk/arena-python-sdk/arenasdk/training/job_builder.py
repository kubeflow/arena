#!/usr/bin/env python
from __future__ import annotations
import abc 
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.fields.fields import *
from arenasdk.enums.types import ArenaErrorType
from arenasdk.enums.types import TrainingJobType
from arenasdk.training.job import TrainingJob
from typing import List
from typing import Dict

def process_items(items: List[str]):
    for item in items:
        print(item) 

class JobBuilder(metaclass=abc.ABCMeta):
    def __init__(self,job_type: TrainingJobType):
        self._job_type: TrainingJobType = job_type
        self._job_name = ""
        self._options: List[Field] = list()
        self._command = ""
        
    def build(self) -> TrainingJob:
        args = list()
        try:
            for field in self._options:
                if not isinstance(field,Field):
                    raise ArenaException(ArenaErrorType.Unknown,"the type of option {} is not Field".format(field))
                field.validate()
                for opt in field.options():
                    args.append(opt)
            return TrainingJob(self._job_name,self._job_type,args,self._command)
        except ArenaException as e:
            raise e 
    
    
    def with_name(self,name: str) -> JobBuilder:
        self._job_name = name
        self._options.append(StringField("--name",name))
        return self 

    def with_image(self,image: str) -> JobBuilder:
        self._options.append(StringField("--image",image))
        return self 

    def with_workers(self,count: int) -> JobBuilder:
        self._options.append(StringField("--workers",str(count)))
        return self

    def with_image_pull_secrets(self,secrets: List[str]) -> JobBuilder:
        self._options.append(StringListField("--image-pull-secret",secrets))
        return self
    
    def with_gpus(self,count: int) -> JobBuilder:
        self._options.append(StringField("--gpus",str(count)))
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
    
    def with_config_files(self,files: Dict[str, str]) -> JobBuilder:
        self._options.append(StringMapField("--config-file",files,":"))
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

    def with_log_dir(self,dir: str) -> JobBuilder:
        self._options.append(StringField("--logdir",dir))
        return self

    def with_priority(self,priority: str) -> JobBuilder:
        self._options.append(StringField("--priority",priority))
        return  self
    
    def enable_rdma(self) -> JobBuilder:
        self._options.append(BoolField("--rdma"))
        return  self
    
    def with_sync_image(self,image: str) -> JobBuilder:
        self._options.append(StringField("--sync-image",image))
        return  self

    def with_sync_mode(self,mode: str) -> JobBuilder:
        self._options.append(StringField("--sync-mode",mode))
        return  self 

    def with_sync_source(self,source: str) -> JobBuilder:
        self._options.append(StringField("--sync-source",source))
        return  self
    
    def enable_tensorboard(self) -> JobBuilder:
        self._options.append(BoolField("--tensorboard"))
        return self 

    def with_tensorboard_image(self,image: str) -> JobBuilder:
        self._options.append(StringField("--tensorboard-image",image))
        return self

    def with_working_dir(self,dir: str) -> JobBuilder:
        self._options.append(StringField("--working-dir",dir))
        return self

    def with_retry_count(self,count: int) -> JobBuilder:
        self._options.append(StringField("--retry",str(count)))
        return self 

    def enable_coscheduling(self) -> JobBuilder:
       self._options.append(BoolField("--gang"))
       return self

    def with_command(self,command: str) -> JobBuilder:
        self._command = command
        return self