#!/usr/bin/env python

from arenasdk.enums.types import ArenaErrorType
from arenasdk.enums.types import ServingJobType
from typing import List
from typing import Dict

class ServingJob(object):
    def __init__(self,job_name: str,job_type: ServingJobType,version: str,args: List[str],command: str):
        self._job_name: str = job_name
        self._job_type: ServingJobType = job_type
        self._version = version
        self._args: List[str] = args
        self._command: str = command
        
    def get_name(self) -> str:
        return self._job_name
    
    def get_type(self) -> ServingJobType:
        return self._job_type
    
    def get_args(self)-> List[str]:
        return self._args
    
    def get_command(self) -> str:
        return self._command
    
    def get_version(self) -> str:
        return self._version
		
