package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class NodeSet {

    private GPUExclusiveNode[] gpuExclusiveNodes;
    private GPUShareNode[] gpuShareNodes;
    private GPUTopologyNode[] gpuTopologyNodes;
    private NormalNode[] normalNodes;

    public GPUExclusiveNode[] getGpuExclusiveNodes() {
        return gpuExclusiveNodes;
    }

    public void setGpuExclusiveNodes(GPUExclusiveNode[] gpuExclusiveNodes) {
        this.gpuExclusiveNodes = gpuExclusiveNodes;
    }

    public GPUShareNode[] getGpuShareNodes() {
        return gpuShareNodes;
    }

    public void setGpuShareNodes(GPUShareNode[] gpuShareNodes) {
        this.gpuShareNodes = gpuShareNodes;
    }

    public GPUTopologyNode[] getGpuTopologyNodes() {
        return gpuTopologyNodes;
    }

    public void setGpuTopologyNodes(GPUTopologyNode[] gpuTopologyNodes) {
        this.gpuTopologyNodes = gpuTopologyNodes;
    }

    public NormalNode[] getNormalNodes() {
        return normalNodes;
    }

    public void setNormalNodes(NormalNode[] normalNodes) {
        this.normalNodes = normalNodes;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
