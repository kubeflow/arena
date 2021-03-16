#!/usr/bin/env python
import os
import sys
import time
from arenasdk.client.client import ArenaClient
from arenasdk.enums.types import *
from arenasdk.exceptions.arena_exception import *
from arenasdk.training.tensorflow_job_builder import *
from arenasdk.logger.logger import LoggerBuilder

def main():
    print("start to test arena-python-sdk")
    client = ArenaClient("","default","info","arena-system")
    print("create ArenaClient succeed.")
    print("start to create tfjob")
    job_name = "tf-distributed-test"
    job_type = TrainingJobType.TFTrainingJob
    try:
        # build the training job 
        job =  TensorflowJobBuilder().with_name(job_name).with_workers(1).with_gpus(1).with_worker_image("cheyang/tf-mnist-distributed:gpu").with_ps_image("cheyang/tf-mnist-distributed:cpu").with_ps_count(1).enable_tensorboard().with_tensorboard_image("registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.12.0-devel").with_command("python /app/main.py").build()
        # if training job is not existed,create it 
        if not client.training().get(job_name,job_type):
            output = client.training().submit(job)
            print(output)
        else:
            print("the job {} has been created,skip to create it".format(job_name))
        count = 0
        # waiting training job to be running 
        while True:
            if count > 160:
                raise Exception("timeout for waiting job to be running")
            jobInfo = client.training().get(job_name,job_type)
            if jobInfo.get_status() == TrainingJobStatus.TrainingJobPending:
                print("job status is PENDING,waiting...")
                count = count + 1
                time.sleep(5)
                continue
            print("current status is {} of job {}".format(jobInfo.get_status().value,job_name))
            break
        # get the training job logs 
        logger = LoggerBuilder().with_accepter(sys.stdout).with_follow().with_since("5m")
        jobInfo.get_instances()[0].get_logs(logger)
        # display the training job information
        print(str(jobInfo))
        # delete the training job 
        client.training().delete(job_name,job_type)
    except ArenaException as e:
        print(e)

main()
    
