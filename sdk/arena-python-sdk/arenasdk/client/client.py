#!/usr/bin/env python
import os
from arenasdk.client.training_client import TrainingClient
from arenasdk.client.serving_client import ServingClient
from arenasdk.client.node_client import NodeClient
class ArenaClient(object):
    def __init__(self,kubeconfig: str = "",namespace: str = "default",loglevel: str = "info",arena_namespace: str = "arena-system"):
        self.kubeconfig = kubeconfig
        self.namespace = namespace
        self.loglevel = loglevel
        self.arena_namespace = arena_namespace
        if self.kubeconfig and self.kubeconfig != "":
            os.environ['KUBECONFIG']= self.kubeconfig
        if self.arena_namespace and self.arena_namespace != "":
            os.environ['ARENA_NAMESPACE']= self.arena_namespace
        if self.loglevel and self.loglevel != "":
            os.environ['ARENA_LOG_LEVEL']= self.loglevel
        if self.namespace and self.namespace != "":
            os.environ['DEFAULT_NAMESPACE']= self.namespace

    def training(self) -> TrainingClient:
        return TrainingClient(self.kubeconfig,self.namespace,self.loglevel,self.arena_namespace)

    def serving(self) -> ServingClient:
        return ServingClient(self.kubeconfig,self.namespace,self.loglevel,self.arena_namespace)
    
    def nodes(self) -> NodeClient:
        return NodeClient(self.kubeconfig,self.namespace,self.loglevel,self.arena_namespace)
		