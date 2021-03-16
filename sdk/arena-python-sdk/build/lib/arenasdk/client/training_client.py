#!/usr/bin/env python
from __future__ import annotations
import json
from typing import List
from typing import Dict
from arenasdk.enums.types import *
from arenasdk.training.job import TrainingJob
from arenasdk.training.training_job_info import TrainingJobInfo
from arenasdk.training.training_job_info import generate_training_job_info
from arenasdk.common.util import Command
from arenasdk.exceptions.arena_exception import ArenaException

class TrainingClient(object):
    def __init__(self,kubeconfig: str,namespace: str,loglevel: str,arena_namespace: str):
        self.kubeconfig = kubeconfig
        self.namespace = namespace
        self.loglevel = loglevel
        self.arena_namespace = arena_namespace

    def namespace(self,namespace: str) -> TrainingClient:
        return TrainingClient(self.kubeconfig,namespace,self.loglevel,self.arena_namespace)

    def submit(self,job: TrainingJob) -> str:
        cmds = self.__generate_commands("submit")
        cmds.append(job.get_type().value)
        for opt in job.get_args():
            cmds.append(opt)
        cmds.append(job.get_command())
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                err_message = "the job {} is already exist, please delete it first.".format(job.get_name())
                if stdout.find(err_message) != -1 or stderr.find(err_message) != -1:
                    raise ArenaException(ArenaErrorType.TrainingJobExistError,stdout + stderr)
                else:
                    raise ArenaException(ArenaErrorType.SubmitTrainingJobError,stdout + stderr)
            return stdout
        except ArenaException as e:
            raise e
    
    def list(self,job_type: TrainingJobType,all_namespaces: bool) -> List[TrainingJobInfo]:
        def convert(json_object):
            job_infos = list()
            for obj in json_object:
                job_infos.append(generate_training_job_info(obj))
            return job_infos
        cmds = self.__generate_commands("list")
        if job_type != TrainingJobType.AllTrainingJob and job_type != TrainingJobType.UnknownTrainingJob:
            cmds.append("--type=" + job_type.value)
        if all_namespaces:
            cmds.append("-A")
        cmds.append("-o")
        cmds.append("json")
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                raise ArenaException(ArenaErrorType.ListTrainingJobError,stdout + stderr)
            return json.loads(stdout,object_hook=convert)
        except ArenaException as e:
            raise e

    def get(self,job_name: str,job_type: TrainingJobType) -> TrainingJobInfo:
        cmds = self.__generate_commands("get")
        cmds.append(job_name)
        if job_type != TrainingJobType.AllTrainingJob and job_type != TrainingJobType.UnknownTrainingJob:
            cmds.append("--type=" + job_type.value)
        cmds.append("-o")
        cmds.append("json")
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                err_message = "Not found training job {} in namespace {},please use 'arena submit' to create it.".format(job_name,self.namespace)
                if stdout.find(err_message) != -1 or stderr.find(err_message) != -1:
                    return None
                else:
                    raise ArenaException(ArenaErrorType.GetTrainingJobError,stdout + stderr)
            data = json.loads(stdout)
            return generate_training_job_info(data)
        except ArenaException as e:
            raise e
    
    def delete(self, job_name: str,job_type: TrainingJobType) -> str:
        cmds = self.__generate_commands("delete")
        cmds.append(job_name)
        if job_type != TrainingJobType.AllTrainingJob and job_type != TrainingJobType.UnknownTrainingJob:
            cmds.append("--type=" + job_type.value)
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                raise ArenaException(ArenaErrorType.DeleteTrainingJobError,stdout + stderr)
            return stdout
        except ArenaException as e:
            raise e
    
    def prune(self,since: str,all_namespaces: bool) -> str:
        cmds = self.__generate_commands("prune")
        cmds.append("--since=" + since)
        if all_namespaces:
            cmds.append("-A")
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                raise ArenaException(ArenaErrorType.PruneTrainingJobError,stdout + stderr)
            return stdout
        except ArenaException as e:
            raise e
        

    def scalein(self,job: TrainingJob) -> str:
        cmds = self.__generate_commands("scalein")
        cmds.append(job.get_type().value)
        for opt in job.get_args():
            cmds.append(opt)
        cmds.append(job.get_command())
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                raise ArenaException(ArenaErrorType.ScaleinTrainingJobError,stdout + stderr)
            return stdout
        except ArenaException as e:
            raise e        

    def scaleout(self,job: TrainingJob) -> str:
        cmds = self.__generate_commands("scaleout")
        cmds.append(job.get_type().value)
        for opt in job.get_args():
            cmds.append(opt)
        cmds.append(job.get_command())
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                raise ArenaException(ArenaErrorType.ScaleoutTrainingJobError,stdout + stderr)
            return stdout
        except ArenaException as e:
            raise e        
         
    def __generate_commands(self,*sub_commands: List[str]) -> List[str]:
        arena_cmds = list()
        arena_cmds.append(ARENA_BINARY)
        for c in sub_commands:
            arena_cmds.append(c)
        if self.kubeconfig != "":
            arena_cmds.append("--config=" + self.kubeconfig)
        if self.namespace != "":
            arena_cmds.append("--namespace=" + self.namespace)
        if self.arena_namespace != "":
            arena_cmds.append("--arena-namespace=" + self.arena_namespace)
        if self.loglevel != "":
            arena_cmds.append("--loglevel=" + self.loglevel)
        return arena_cmds
        		
