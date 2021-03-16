#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import TrainingJobType
from arenasdk.fields.fields import *
from arenasdk.training.job_builder import JobBuilder

class HorovodJobBuilder(JobBuilder):
    def __init__(self):
        super().__init__(TrainingJobType.HorovodTrainingJob)
        
    def with_ssh_port(self,port: int) ->HorovodJobBuilder:
        self._options.append(StringField("--ssh-port",str(port)))
        return self 
    
    def with_cpu(self,cpu: str) ->HorovodJobBuilder:
        self._options.append(StringField("--cpu",cpu))
        return self 

    def with_memory(self,memory: str) ->HorovodJobBuilder:
        self._options.append(StringField("--memory",memory))
        return self
    
    def with_name(self,name: str) ->HorovodJobBuilder:
        super().with_name(name)
        return self 

    def with_image(self,image: str) ->HorovodJobBuilder:
        super().with_image(image)
        return self 

    def with_workers(self,count: int) ->HorovodJobBuilder:
        super().with_workers(count)
        return self

    def with_image_pull_secrets(self,secrets: List[str]) ->HorovodJobBuilder:
        super().with_image_pull_secrets(secrets)
        return self
    
    def with_gpus(self,count: int) ->HorovodJobBuilder:
        super().with_gpus(count)
        return self

    def with_envs(self,envs: Dict[str,str]) ->HorovodJobBuilder:
        super().with_envs(envs)
        return self

    def with_node_selectors(self,selectors: Dict[str, str]) ->HorovodJobBuilder:
        super().with_node_selectors(selectors)
        return self

    def with_tolerations(self,tolerations: List[str]) ->HorovodJobBuilder:
        super().with_tolerations(tolerations)
        return self
    
    def with_config_files(self,files: Dict[str, str]) ->HorovodJobBuilder:
        super().with_config_files(files)
        return self 

    def with_annotations(self,annotions: Dict[str, str]) ->HorovodJobBuilder:
        super().with_annotations(annotions)
        return self

    def with_datas(self,datas: Dict[str,str]) ->HorovodJobBuilder:
        super().with_datas(datas)
        return self 
    
    def with_data_dirs(self,data_dirs: Dict[str, str]) ->HorovodJobBuilder:
        super().with_data_dirs(data_dirs)
        return  self

    def with_log_dir(self,dir: str) ->HorovodJobBuilder:
        super().with_log_dir(dir)
        return self

    def with_priority(self,priority: str) ->HorovodJobBuilder:
        super().with_priority(priority)
        return  self
    
    def enable_rdma(self) ->HorovodJobBuilder:
        super().enable_rdma()
        return  self
    
    def with_sync_image(self,image: str) ->HorovodJobBuilder:
        super().with_sync_image(image)
        return  self

    def with_sync_mode(self,mode: str) ->HorovodJobBuilder:
        super().with_sync_mode(mode)
        return  self 

    def with_sync_source(self,source: str) ->HorovodJobBuilder:
        super().with_sync_source(source)
        return  self
    
    def enable_tensorboard(self) ->HorovodJobBuilder:
        super().enable_tensorboard()
        return self 

    def with_tensorboard_image(self,image: str) ->HorovodJobBuilder:
        super().with_tensorboard_image(image)
        return self

    def with_working_dir(self,dir: str) ->HorovodJobBuilder:
        super().with_working_dir(dir)
        return self

    def with_retry_count(self,count: int) ->HorovodJobBuilder:
        super().with_retry_count(count)
        return self 

    def enable_coscheduling(self) ->HorovodJobBuilder:
        super().enable_coscheduling()
        return self

    def with_command(self,command: str) ->HorovodJobBuilder:
        super().with_command(command)
        return self
