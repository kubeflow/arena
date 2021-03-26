#!/usr/bin/env python
import os
import sys
import time
from arenasdk.client.client import ArenaClient
from arenasdk.enums.types import *
from arenasdk.exceptions.arena_exception import *
from arenasdk.training.mpi_job_builder import *
from arenasdk.logger.logger import LoggerBuilder

def main():
    print("start to test arena-python-sdk")
    client = ArenaClient("","default","info","arena-system")
    print("create ArenaClient succeed.")
    print("start to create mpi job")
    job_name = "mpi-dist"
    job_type = TrainingJobType.MPITrainingJob
    try:
        # build the training job 
        job = MPIJobBuilder().with_name(job_name).with_workers(1).enable_tensorboard().with_image("registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5").with_command('''"sh -c -- 'mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10'"''').build()
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
    
