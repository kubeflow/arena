#!/usr/bin/env python
import os
import sys
from arenasdk.client.client import ArenaClient
from arenasdk.enums.types import *
from arenasdk.exceptions.arena_exception import *
from arenasdk.training.mpi_job_builder import *
from arenasdk.logger.logger import LoggerBuilder
from arenasdk.common.log import Log

logger = Log(__name__).get_logger()

def main():
    print("start to test arena-python-sdk")
    client = ArenaClient("~/.kube/config-gpushare","default","debug","arena-system")
    print("create ArenaClient succeed.")
    print("start to get node details")
    try:
        nodes = client.nodes().all()
        logger.debug("nodes: %s",nodes)
        normal_nodes = nodes.get_normal_nodes()
        for node in normal_nodes:
            print(node.to_dict())
        gpushare_nodes = nodes.get_gpushare_nodes()
        for node in gpushare_nodes:
            print(node.to_dict())
            for i in node.get_instances():
                print(i.to_dict())
            for d in node.get_devices():
                print(d.to_dict())
        gpu_topology_nodes = nodes.get_gpu_topology_nodes()
        for node in gpu_topology_nodes:
            print(node.to_dict())
            for i in node.get_instances():
                print(i.to_dict())
            for d in node.get_devices():
                print(d.to_dict())
        gpu_exclusive_nodes = nodes.get_gpu_exclusive_nodes()
        for node in gpu_exclusive_nodes:
            for i in node.get_instances():
                print(i.to_dict())
            print(node.to_dict())
    except ArenaException as e:
        print(e)

main()
    