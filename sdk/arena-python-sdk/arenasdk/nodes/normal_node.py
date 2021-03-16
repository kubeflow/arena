#!/usr/bin/env python
from __future__ import annotations
from arenasdk.nodes.node import Node
from typing import List
from typing import Dict 

class NormalNode(Node):
    def __init__(self):
        super().__init__()
    
    def to_dict(self) -> dict:
        return self.__dict__.copy()
    
def build_normal_nodes(data: dict) -> NormalNode:
    cls = NormalNode()
    cls.set_name(data["name"])
    cls.set_ip(data["ip"])
    cls.set_status(data["status"])
    cls.set_role(data["role"])
    cls.set_node_type(data["type"])
    return cls
