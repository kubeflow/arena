#!/usr/bin/env python
import os
import sys
import time
from arenasdk.client.client import ArenaClient
from arenasdk.enums.types import *
from arenasdk.exceptions.arena_exception import *
from arenasdk.training.pytorch_job_builder import *
from arenasdk.logger.logger import LoggerBuilder

def main():
    print("start to test arena-python-sdk")
    client = ArenaClient("","default","info","arena-system")
    print("create ArenaClient succeed.")
    print("start to create pytorchjob")
    job_name = "pytorchjob-test"
    job_type = TrainingJobType.PytorchTrainingJob
    try:
        # build the training job 
        job =  PytorchJobBuilder().with_name(job_name).with_workers(1).with_gpus(1).with_sync_mode("git").with_sync_source("https://github.com/happy2048/mnist-pytorch.git").with_image("registry.cn-beijing.aliyuncs.com/ai-samples/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime").with_command("'python /root/code/mnist-pytorch/mnist.py --backend gloo'").build()
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
    
