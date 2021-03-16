#!/usr/bin/env python

from arenasdk.enums.types import ArenaErrorType
from arenasdk.enums.types import TrainingJobType
from typing import List
from typing import Dict

class TrainingJob(object):
    def __init__(self,job_name: str,job_type: TrainingJobType,args: List[str],command: str):
        self._job_name: str = job_name
        self._job_type: TrainingJobType = job_type
        self._args: List[str] = args
        self._command: str = command
        
    def get_name(self) -> str:
        return self._job_name
    
    def get_type(self) -> TrainingJobType:
        return self._job_type
    
    def get_args(self)-> List[str]:
        return self._args
    
    def get_command(self) -> str:
        return self._command
		
