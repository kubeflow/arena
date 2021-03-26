#!/usr/bin/env python
import json
from enum import Enum

ARENA_BINARY = "arena"

class TrainingJobType(Enum):
    TFTrainingJob = "tfjob"
    MPITrainingJob = "mpijob"
    PytorchTrainingJob = "pytorchjob"
    HorovodTrainingJob = "horovodjob"
    VolcanoTrainingJob = "volcanojob"
    SparkTrainingJob = "sparkjob"
    ETTrainingJob = "etjob"
    AllTrainingJob = ""
    UnknownTrainingJob = "unknown"
    @classmethod
    def convert(cls,alias):
        if alias == "":
            return TrainingJobType.AllTrainingJob
        for name,value in TrainingJobType.__members__.items():
            if alias == value.value:
                return value 
        return TrainingJobType.UnknownTrainingJob


class ServingJobType(Enum):
    TFServingJob = ("tf","Tensorflow")
    TRTServingJob = ("trt","Tensorrt")
    KFServingJob = ("kf","KFServing")
    CustomServingJob = ("custom","Custom")
    AllServingJob = ("","")
    UnknownServingJob = ("unknown","unknown")
    @classmethod
    def convert(cls,alias):
        if alias == "":
            return ServingJobType.AllServingJob 
        for name,value in ServingJobType.__members__.items():
            if alias == value.value[0]:
                return value
            if alias.lower() == value.value[1].lower():
                return value 
        return ServingJobType.UnknownServingJob

class NodeType(Enum):
    GPUExeclusiveNodeType = ("e","GPUExeclusive") 
    GPUShareNodeType = ("s","GPUShare")
    GPUTopologyNodeType = ("t","GPUTopology")
    NormalNodeType = ("n","Normal")
    AllNodeType = ("","")
    UnknownNodeType = ("unknown","unknown")  
    @classmethod
    def convert(cls,alias):
        if alias == "":
            return NodeType.AllNodeType
        for name,value in NodeType.__members__.items():
            if alias == value.value[0]:
                return value
            if alias.lower() == value.value[1].lower():
                return value 
        return NodeType.UnknownNodeType
    
class ArenaErrorType(Enum):
    Unknown = ("unknown","unknown type")
    SubmitTrainingJobError = ("training_job_submit","failed to submit training jobs")
    GetTrainingJobError = ("training_job_get","failed to get training jobs")
    ListTrainingJobError = ("training_job_list","failed to list training jobs")
    LogsTrainingJobError = ("training_job_logs","failed to get training job logs")
    DeleteTrainingJobError = ("training_job_delete","failed to delete training jobs")
    PruneTrainingJobError = ("prune_training_jobs","failed to prune training jobs")
    ScaleinTrainingJobError = ("scalein_training_jobs","failed to scale in training jobs")
    ScaleoutTrainingJobError = ("scaleout_training_jobs","failed to scale out training jobs")
    SubmitServingJobError = ("serving_job_submit","failed to submit serving jobs")
    GetServingJobError = ("serving_job_get","failed to get serving jobs")
    ListServingJobsError = ("serving_job_list","failed to list serving jobs")
    LogsServingJobError = ("serving_job_logs","failed to get serving job logs")
    DeleteServingJobError = ("serving_job_delete","failed to delete serving jobs")
    TrafficRouterSplitServingJobError = ("traffic_router_split","failed to split traffic router of the serving job")
    ValidateArgsError = ("validate_args","failed to validate args of submiting jobs")
    ServingJobExistError = ("serving_job_exist","serving job is existed")
    ServingJobNotFoundError = ("serving_job_not_found","not found serving job")
    TrainingJobExistError = ("training_job_exist","training job is existed")
    TrainingJobNotFoundError = ("training_job_not_found","not found training job")
    TopNodeError = ("top_node","failed to get node information")
    InvalidTrainingJobType = ("invalid_training_job_type","training job type is invalid")
    InvalidServingJobType = ("invalid_serving_job_type","serving job type is invalid")
 
class TrainingJobStatus(Enum):
    TrainingJobPending = "PENDING"
    TrainingJobRunning = "RUNNING"
    TrainingJobSucceeded = "SUCCEEDED"
    TrainingJobFailed = "FAILED"
    TrainingJobScaling = "SCALING"
    UnknownTrainingJobStatus = "UNKNOWN"
    @classmethod
    def convert(cls,alias):
        for name,value in TrainingJobStatus.__members__.items():
            if alias == value.value:
                return value 
        return TrainingJobStatus.UnknownTrainingJobStatus
