package com.github.kubeflow.arena.client;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

import com.alibaba.fastjson.JSONObject;
import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.exceptions.ExitCodeException;
import com.github.kubeflow.arena.model.nodes.*;
import com.github.kubeflow.arena.enums.NodeType;
import com.github.kubeflow.arena.model.serving.ServingJobInfo;
import com.github.kubeflow.arena.utils.Command;

public class NodeClient {
    private String namespace;
    private String kubeConfig;
    private String loglevel;
    private String arenaSystemNamespace;
    private static String  arenaBinary = "arena";

    public NodeClient(String kubeConfig,String namespace,String loglevel,String arenaSystemNamespace) {
        this.namespace = namespace;
        this.kubeConfig = kubeConfig;
        this.loglevel = loglevel;
        this.arenaSystemNamespace = arenaSystemNamespace;
    }

    public NodeClient namespace(String namespace) {
        return new NodeClient(this.kubeConfig,namespace,this.loglevel,this.arenaSystemNamespace);
    }

    public NodeSet all(String... nodeNames)  throws ArenaException,IOException {
        return this.filter(NodeType.AllNodeType,nodeNames);
    }

    public NodeSet filter(NodeType nodeType,String... nodeNames) throws ArenaException,IOException {
        ArrayList<String> cmds = this.generateCommands("top","node","-d","-o","json");
        if (!nodeType.equals(NodeType.AllNodeType) && !nodeType.equals(NodeType.UnknownNodeType)){
            cmds.add("-m="+nodeType.shortHand());
        }
        if (nodeNames != null && nodeNames.length != 0) {
            for(int i = 0;i < nodeNames.length;i++) {
                cmds.add(nodeNames[i]);
            }
        }
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            NodeSet nodeSet = JSON.parseObject(output, NodeSet.class);
            return nodeSet;
        }catch(ExitCodeException e){
            throw new ArenaException(ArenaErrorEnum.TOP_NODE,e.getMessage());
        }
    }

    public GPUExclusiveNode[] gpuExclusiveNodes(String... nodeNames) throws ArenaException,IOException {
        NodeSet nodeSet = this.filter(NodeType.GPUExclusiveNodeType,nodeNames);
        return nodeSet.getGpuExclusiveNodes();
    }

    public GPUTopologyNode[] gpuTopologyNodes(String... nodeNames) throws ArenaException,IOException {
        NodeSet nodeSet = this.filter(NodeType.GPUTopologyNodeType,nodeNames);
        return nodeSet.getGpuTopologyNodes();
    }

    public GPUShareNode[] gpuShareNodes(String... nodeNames)  throws ArenaException,IOException {
        NodeSet nodeSet = this.filter(NodeType.GPUShareNodeType,nodeNames);
        return nodeSet.getGpuShareNodes();
    }

    public NormalNode[] normalNodes(String... nodeNames) throws ArenaException,IOException {
        NodeSet nodeSet = this.filter(NodeType.NormalNodeType,nodeNames);
        return nodeSet.getNormalNodes();
    }


    private ArrayList<String> generateCommands(String... subCommand) {
        ArrayList<String> cmds = new ArrayList<String>();
        cmds.add(arenaBinary);
        for(int i = 0;i < subCommand.length;i++) {
            cmds.add(subCommand[i]);
        }
        if (this.namespace.length() != 0) {
            cmds.add("--namespace=" + this.namespace);
        }
        if (this.kubeConfig.length() != 0) {
            cmds.add("--config=" + this.kubeConfig);
        }
        if (this.loglevel.length() != 0) {
            cmds.add("--loglevel=" + this.loglevel);
        }
        if (this.arenaSystemNamespace.length() != 0) {
            cmds.add("--arena-namespace=" + this.arenaSystemNamespace);
        }
        return cmds;
    }
}
