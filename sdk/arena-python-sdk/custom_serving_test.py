#!/usr/bin/env python
import os
import sys
from arenasdk.client.client import ArenaClient
from arenasdk.enums.types import *
from arenasdk.exceptions.arena_exception import *
from arenasdk.serving.custom_job_builder import *
from arenasdk.logger.logger import LoggerBuilder

def main():
    print("start to test arena-python-sdk")
    client = ArenaClient("~/.kube/config-gpushare","default","info","arena-system")
    print("create ArenaClient succeed.")
    print("start to create mpi job")
    jobName = "custom-serving-test"
    jobType = ServingJobType.CustomServingJob
    version = "alpha"
    try:
        job = CustomServingJobBuilder().with_name(jobName).with_version(version).with_gpus(1).with_replicas(1).with_restful_port(5000).with_image("happy365/fast-style-transfer:latest").with_command("python app.py").build()
        if not client.serving().get(jobName,jobType,version):
            output = client.serving().submit(job)
            print(output)
        else:
            print("the job {} has been created,skip to create it".format(jobName))
        jobInfo = client.serving().get(jobName,jobType,version)
        logger = LoggerBuilder().with_accepter(sys.stdout).with_since("5m")
        jobInfo.get_instances()[0].get_logs(logger)
        client.serving().delete(jobName,jobType,version)
    except ArenaException as e:
        print(e)


main()
    