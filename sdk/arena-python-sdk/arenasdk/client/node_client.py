#!/usr/bin/env python
from __future__ import annotations
import json
from typing import List
from typing import Dict
from arenasdk.enums.types import *
from arenasdk.nodes.gpu_exclusive_node import GPUExclusiveNode
from arenasdk.nodes.gpu_topology_node  import GPUTopologyNode
from arenasdk.nodes.gpushare_node import GPUShareNode
from arenasdk.nodes.normal_node import NormalNode
from arenasdk.nodes.node_set import NodeSet
from arenasdk.nodes.node_set import build_nodes
from arenasdk.common.util import Command
from arenasdk.exceptions.arena_exception import ArenaException

class NodeClient(object):
    def __init__(self,kubeconfig: str,namespace: str,loglevel: str,arena_namespace: str):
        self.kubeconfig = kubeconfig
        self.namespace = namespace
        self.loglevel = loglevel
        self.arena_namespace = arena_namespace

    def namespace(self,namespace: str) -> TrainingClient:
        return NodeClient(self.kubeconfig,namespace,self.loglevel,self.arena_namespace)
    
    def all(self,*node_names: List[str]) -> NodeSet:
        return self._filter(NodeType.AllNodeType,*node_names)
    
    def gpushare_nodes(self,*node_names: List[str]) -> List[GPUShareNode]:
        return self._filter(NodeType.GPUShareNodeType,*node_names).get_gpushare_nodes()
    
    def gpu_exclusive_nodes(self,*node_names: List[str]) -> List[GPUExclusiveNode]:
        return self._filter(NodeType.GPUExclusiveNodeType,*node_names).get_gpu_exclusive_nodes()
    
    def gpu_topology_nodes(self,*node_names: List[str]) -> List[GPUTopologyNode]:
        return self._filter(NodeType.GPUTopologyNodeType,*node_names).get_gpu_topology_nodes()
    
    def normal_nodes(self,*node_names: List[str]) -> List[NormalNode]:
        return self._filter(NodeType.NormalNodeType,*node_names).get_normal_nodes()

    def _filter(self,node_type: NodeType,*node_names: List[str]) -> NodeSet:
        cmds = self.__generate_commands("top","node","-d","-o","json")
        if node_type != NodeType.AllNodeType and node_type != NodeType.UnknownNodeType:
            cmds.append("-m=" + node_type.value[0])
        try:
            status,stdout,stderr = Command(*cmds).run()
            if status != 0:
                raise ArenaException(ArenaErrorType.TopNodeError,stdout + stderr)
            data = json.loads(stdout)
            return build_nodes(data)
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