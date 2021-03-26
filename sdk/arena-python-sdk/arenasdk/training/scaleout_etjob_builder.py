#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import TrainingJobType
from arenasdk.fields.fields import *
from arenasdk.training.job_builder import JobBuilder
from arenasdk.exceptions.arena_exception import ArenaException
from arenasdk.enums.types import ArenaErrorType
from arenasdk.training.job import TrainingJob

class ScaleOutETJobBuilder(object):
    def __init__(self):
        self._job_type: TrainingJobType = TrainingJobType.ETTrainingJob
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
        
    def with_name(self,name: str) -> ScaleOutETJobBuilder:
        self._options.append(StringField("--name",name))
        return self 

    def with_timeout(self,timeout: str) -> ScaleOutETJobBuilder:
        self._options.append(StringField("--timeout",timeout))
        return self 

    def with_retry(self,retry: int) -> ScaleOutETJobBuilder:
        self._options.append(StringField("--retry",str(retry)))
        return self 

    def with_envs(self,envs: Dict[str,str]) -> ScaleOutETJobBuilder:
        self._options.append(StringMapField("--env",envs,"="))
        return self 

    def with_script(self,script: str) -> ScaleOutETJobBuilder:
        self._options.append(StringField("--script",script))
        return self 

    def with_count(self,count: int) -> ScaleOutETJobBuilder:
        self._options.append(StringField("--count",str(count)))
        return self 
