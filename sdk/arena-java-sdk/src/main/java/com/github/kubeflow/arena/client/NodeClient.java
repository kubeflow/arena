package com.github.kubeflow.arena.client;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.NodeType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.exceptions.ExitCodeException;
import com.github.kubeflow.arena.model.nodes.*;
import com.github.kubeflow.arena.utils.Command;

import java.io.IOException;
import java.util.List;

public class NodeClient extends BaseClient {

    public NodeClient(String kubeConfig, String namespace, String loglevel, String arenaSystemNamespace) {
        super(kubeConfig, namespace, loglevel, arenaSystemNamespace);
    }

    public NodeClient namespace(String namespace) {
        return new NodeClient(this.kubeConfig, namespace, this.loglevel, this.arenaSystemNamespace);
    }

    public NodeSet all(String... nodeNames) throws ArenaException, IOException {
        return this.filter(NodeType.AllNodeType, nodeNames);
    }

    public NodeSet filter(NodeType nodeType, String... nodeNames) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("top", "node", "-d", "-o", "json");
        if (!nodeType.equals(NodeType.AllNodeType) && !nodeType.equals(NodeType.UnknownNodeType)) {
            cmds.add("-m=" + nodeType.shortHand());
        }
        if (nodeNames != null && nodeNames.length != 0) {
            for (int i = 0; i < nodeNames.length; i++) {
                cmds.add(nodeNames[i]);
            }
        }
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            NodeSet nodeSet = JSON.parseObject(output, NodeSet.class);
            return nodeSet;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.TOP_NODE, e.getMessage());
        }
    }

    public GPUExclusiveNode[] gpuExclusiveNodes(String... nodeNames) throws ArenaException, IOException {
        NodeSet nodeSet = this.filter(NodeType.GPUExclusiveNodeType, nodeNames);
        return nodeSet.getGpuExclusiveNodes();
    }

    public GPUTopologyNode[] gpuTopologyNodes(String... nodeNames) throws ArenaException, IOException {
        NodeSet nodeSet = this.filter(NodeType.GPUTopologyNodeType, nodeNames);
        return nodeSet.getGpuTopologyNodes();
    }

    public GPUShareNode[] gpuShareNodes(String... nodeNames) throws ArenaException, IOException {
        NodeSet nodeSet = this.filter(NodeType.GPUShareNodeType, nodeNames);
        return nodeSet.getGpuShareNodes();
    }

    public NormalNode[] normalNodes(String... nodeNames) throws ArenaException, IOException {
        NodeSet nodeSet = this.filter(NodeType.NormalNodeType, nodeNames);
        return nodeSet.getNormalNodes();
    }

}
