#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import TrainingJobType
from arenasdk.fields.fields import *
from arenasdk.training.job_builder import JobBuilder
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.enums.types import ArenaErrorType
from arenasdk.training.job import TrainingJob

class VolcanoJobBuilder(object):
    def __init__(self):
        self._job_type: TrainingJobType = TrainingJobType.VolcanoTrainingJob
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
        
    def with_min_available(self,min_available: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--min-available",min_available))
        return self 

    def with_name(self,name: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--name",name))
        return self 

    def with_queue(self,queue: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--queue",queue))
        return self 

    def with_scheduler_name(self,name: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--scheduler-name",name))
        return self 

    def with_task_cpu(self,cpu: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--task-cpu",cpu))
        return self

    def with_task_images(self,images: List[str]) -> VolcanoJobBuilder:
        self._options.append(StringListField("--task-images",images))
        return self
    
    def with_task_memory(self,memory: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--task-memory",memory))
        return self

    def with_task_name(self,name: str) -> VolcanoJobBuilder:
        self._options.append(StringField("--task-name",name))
        return self

    def with_task_port(self,port: int) -> VolcanoJobBuilder:
        self._options.append(StringField("--task-name",str(port)))
        return self

    def with_task_replicas(self,replicas: int) -> VolcanoJobBuilder:
        self._options.append(StringField("--task-replicas",str(replicas)))
        return self
    
    def with_command(self,command: str) -> VolcanoJobBuilder:
        self._command = command
        return self
