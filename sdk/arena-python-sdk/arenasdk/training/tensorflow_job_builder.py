#!/usr/bin/env python
from __future__ import annotations
from arenasdk.enums.types import TrainingJobType
from arenasdk.fields.fields import *
from arenasdk.training.job_builder import JobBuilder

class TensorflowJobBuilder(JobBuilder):
    def __init__(self):
        super().__init__(TrainingJobType.TFTrainingJob)

    def witch_workers(self,count: int) -> TensorflowJobBuilder:
        self._options.append(StringField("--workers",str(count)))
        return self

    def witch_worker_selectors(self,selectors: Dict[str,str]) -> TensorflowJobBuilder:
        self._options.append(StringMapField("--worker-selector",selectors,"="))
        return self

    def witch_worker_port(self,port: int) -> TensorflowJobBuilder:
        self._options.append(StringField("--worker-port",str(port)))
        return self

    def witch_worker_memory(self,memory: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--worker-memory",memory))
        return self

    def witch_worker_image(self,image: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--worker-image",image))
        return self

    def witch_worker_cpu(self,cpu: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--worker-cpu",cpu))
        return self

    def witch_ps_selectors(self,selectors: Dict[str,str]) -> TensorflowJobBuilder:
        self._options.append(StringMapField("--ps-selector",selectors,"="))
        return self
    
    def witch_ps_port(self,port: int) -> TensorflowJobBuilder:
        self._options.append(StringField("--ps-port",str(port)))
        return self

    def witch_ps_memory(self,memory: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--ps-memory",memory))
        return self

    def witch_ps_image(self,image: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--ps-image",image))
        return self

    def witch_ps_cpu(self,cpu: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--ps-cpu",cpu))
        return self

    def witch_ps_count(self,count: int) -> TensorflowJobBuilder:
        self._options.append(StringField("--ps",str(count)))
        return self

    def witch_evaluator_selector(self,selectors: Dict[str,str]) -> TensorflowJobBuilder:
        self._options.append(StringMapField("--evaluator-selector",selectors))
        return self

    def witch_evaluator_memory(self,memory: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--evaluator-memory",memory))
        return self

    def witch_evaluator_cpu(self,cpu: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--evaluator-cpu",cpu))
        return self

    def witch_enable_evaluator(self) -> TensorflowJobBuilder:
        self._options.append(BoolField("--evaluator"))
        return self

    def witch_chief_selectors(self, selectors: Dict[str,str]) -> TensorflowJobBuilder:
        self._options.append(StringMapField("--chief-selector",selectors))
        return self

    def with_chief_port(self,port: int) -> TensorflowJobBuilder:
        self._options.append(StringField("--chief-port",str(port)))
        return self 

    def with_chief_memory(self,memory: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--chief-memory",memory))
        return self 

    def with_chief_cpu(self,cpu: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--chief-cpu",cpu))
        return self 
    
    def with_enable_chief(self) -> TensorflowJobBuilder:
        self._options.append(BoolField("--chief"))
        return self
            
    def with_cpu(self,cpu: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--cpu",cpu))
        return self 

    def with_memory(self,memory: str) -> TensorflowJobBuilder:
        self._options.append(StringField("--memory",memory))
        return self
    
    def with_name(self,name: str) -> TensorflowJobBuilder:
        super().with_name(name)
        return self 

    def with_image(self,image: str) -> TensorflowJobBuilder:
        super().with_image(image)
        return self 

    def with_workers(self,count: int) -> TensorflowJobBuilder:
        super().with_workers(count)
        return self

    def with_image_pull_secrets(self,secrets: List[str]) -> TensorflowJobBuilder:
        super().with_image_pull_secrets(secrets)
        return self
    
    def with_gpus(self,count: int) -> TensorflowJobBuilder:
        super().with_gpus(count)
        return self

    def with_envs(self,envs: Dict[str,str]) -> TensorflowJobBuilder:
        super().with_envs(envs)
        return self

    def with_node_selectors(self,selectors: Dict[str, str]) -> TensorflowJobBuilder:
        super().with_node_selectors(selectors)
        return self

    def with_tolerations(self,tolerations: List[str]) -> TensorflowJobBuilder:
        super().with_tolerations(tolerations)
        return self
    
    def with_config_files(self,files: Dict[str, str]) -> TensorflowJobBuilder:
        super().with_config_files(files)
        return self 

    def with_annotations(self,annotions: Dict[str, str]) -> TensorflowJobBuilder:
        super().with_annotations(annotions)
        return self

    def with_datas(self,datas: Dict[str,str]) -> TensorflowJobBuilder:
        super().with_datas(datas)
        return self 
    
    def with_data_dirs(self,data_dirs: Dict[str, str]) -> TensorflowJobBuilder:
        super().with_data_dirs(data_dirs)
        return  self

    def with_log_dir(self,dir: str) -> TensorflowJobBuilder:
        super().with_log_dir(dir)
        return self

    def with_priority(self,priority: str) -> TensorflowJobBuilder:
        super().with_priority(priority)
        return  self
    
    def enable_rdma(self) -> TensorflowJobBuilder:
        super().enable_rdma()
        return  self
    
    def with_sync_image(self,image: str) -> TensorflowJobBuilder:
        super().with_sync_image(image)
        return  self

    def with_sync_mode(self,mode: str) -> TensorflowJobBuilder:
        super().with_sync_mode(mode)
        return  self 

    def with_sync_source(self,source: str) -> TensorflowJobBuilder:
        super().with_sync_source(source)
        return  self
    
    def enable_tensorboard(self) -> TensorflowJobBuilder:
        super().enable_tensorboard()
        return self 

    def with_tensorboard_image(self,image: str) -> TensorflowJobBuilder:
        super().with_tensorboard_image(image)
        return self

    def with_working_dir(self,dir: str) -> TensorflowJobBuilder:
        super().with_working_dir(dir)
        return self

    def with_retry_count(self,count: int) -> TensorflowJobBuilder:
        super().with_retry_count(count)
        return self 

    def enable_coscheduling(self) -> TensorflowJobBuilder:
        super().enable_coscheduling()
        return self

    def with_command(self,command: str) -> TensorflowJobBuilder:
        super().with_command(command)
        return self
