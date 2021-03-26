#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import TrainingJobType
from arenasdk.fields.fields import *
from arenasdk.training.job_builder import JobBuilder
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.enums.types import ArenaErrorType
from arenasdk.training.job import TrainingJob

class SparkJobBuilder(object):
    def __init__(self):
        self._job_type: TrainingJobType = TrainingJobType.SparkTrainingJob
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
        
    def with_name(self,name: str) -> SparkJobBuilder:
        self._options.append(StringField("--name",name))
        return self 

    def with_image(self,image: str) -> SparkJobBuilder:
        self._options.append(StringField("--image",image))
        return self 

    def with_replicas(self,replicas: int) -> SparkJobBuilder:
        self._options.append(StringField("--replicas",str(replicas)))
        return self 

    def with_main_class(self,main_class: str) -> SparkJobBuilder:
        self._options.append(StringField("--main-class",main_class))
        return self 

    def with_jar(self,jar: str) -> SparkJobBuilder:
        self._options.append(StringField("--jar",jar))
        return self 

    def with_driver_cpu(self,driver_cpu: str) -> SparkJobBuilder:
        self._options.append(StringField("--driver-cpu-request",driver_cpu))
        return self 

    def with_driver_memory(self,driver_memory: str) -> SparkJobBuilder:
        self._options.append(StringField("--driver-memory-request",driver_memory))
        return self 
    
    def with_executor_memory(self,executor_memory: str) -> SparkJobBuilder:
        self._options.append(StringField("--executor-memory-request",executor_memory))
        return self 

    def with_executor_cpu(self,executor_cpu: str) -> SparkJobBuilder:
        self._options.append(StringField("--executor-cpu-request",executor_cpu))
        return self 

    def with_command(self,command: str) -> SparkJobBuilder:
        self._command = command
        return self
