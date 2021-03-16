#!/usr/bin/env python
from __future__ import annotations
from arenasdk.nodes.gpu_exclusive_node import GPUExclusiveNode
from arenasdk.nodes.gpu_exclusive_node import build_gpu_exclusive_nodes
from arenasdk.nodes.gpu_topology_node  import GPUTopologyNode
from arenasdk.nodes.gpu_topology_node  import build_gpu_topology_nodes
from arenasdk.nodes.gpushare_node import GPUShareNode
from arenasdk.nodes.gpushare_node import build_gpushare_nodes
from arenasdk.nodes.normal_node import NormalNode
from arenasdk.nodes.normal_node import build_normal_nodes
from typing import List
from typing import Dict
from arenasdk.common.log import Log

logger = Log(__name__).get_logger()

class NodeSet(object):
    def __init__(self):
        self._normal_nodes: List[NormalNode]
        self._gpu_topology_nodes: List[GPUTopologyNode]
        self._gpu_exclusive_nodes: List[GPUExclusiveNode]
        self._gpushare_nodes: List[GPUShareNode]
    
    def set_normal_nodes(self,nodes: List[NormalNode]) -> None:
        self._normal_nodes = nodes
    
    def get_normal_nodes(self) -> List[NormalNode]:
        return self._normal_nodes
    
    def set_gpu_topology_nodes(self,nodes: List[GPUTopologyNode]) -> None:
        self._gpu_topology_nodes = nodes
    
    def get_gpu_topology_nodes(self) -> List[GPUTopologyNode]:
        return self._gpu_topology_nodes
    
    def set_gpu_exclusive_nodes(self,nodes: List[GPUExclusiveNode]) -> None:
        self._gpu_exclusive_nodes = nodes 
    
    def get_gpu_exclusive_nodes(self) -> List[GPUExclusiveNode]:
        return self._gpu_exclusive_nodes
    
    def set_gpushare_nodes(self,nodes: List[GPUShareNode]) -> None:
        self._gpushare_nodes = nodes
    
    def get_gpushare_nodes(self) -> List[GPUShareNode]:
        return self._gpushare_nodes


def build_nodes(data) -> NodeSet:
    cls = NodeSet()
    logger.debug("get all node informations: %s",data)
    normal_nodes = list()
    for node_info in data["normalNodes"]:
        normal_nodes.append(build_normal_nodes(node_info))
    cls.set_normal_nodes(normal_nodes)
    gpu_topology_nodes = list()
    for node_info in data["gpuTopologyNodes"]:
        gpu_topology_nodes.append(build_gpu_topology_nodes(node_info))
    cls.set_gpu_topology_nodes(gpu_topology_nodes)
    gpu_exclusive_nodes = list()
    for node_info in data["gpuExclusiveNodes"]:
        gpu_exclusive_nodes.append(build_gpu_exclusive_nodes(node_info))
    cls.set_gpu_exclusive_nodes(gpu_exclusive_nodes)
    gpu_share_nodes = list()
    for node_info in data["gpuShareNodes"]:
        gpu_share_nodes.append(build_gpushare_nodes(node_info))
    cls.set_gpushare_nodes(gpu_share_nodes)
    return cls
            