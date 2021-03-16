package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUExclusiveNode extends Node {

    private int totalGPUs;
    private int allocatedGPUs;
    private int unhealthyGPUs;
    private GPUExclusiveNodePod[] instances;
    private GPUMetric[] gpuMetrics;

    public GPUExclusiveNode() {
        super();
    }

    public void setAllocatedGPUs(int allocatedGPUs) {
        this.allocatedGPUs = allocatedGPUs;
    }

    public int getAllocatedGPUs() {
        return allocatedGPUs;
    }

    public void setUnhealthyGPUs(int unhealthyGPUs) {
        this.unhealthyGPUs = unhealthyGPUs;
    }

    public int getUnhealthyGPUs() {
        return unhealthyGPUs;
    }

    public void setTotalGPUs(int totalGPUs) {
        this.totalGPUs = totalGPUs;
    }

    public int getTotalGPUs() {
        return totalGPUs;
    }

    public void setInstances(GPUExclusiveNodePod[] instances) {
        this.instances = instances;
    }

    public GPUExclusiveNodePod[] getInstances() {
        return instances;
    }

    public void setGpuMetrics(GPUMetric[] gpuMetrics) {
        this.gpuMetrics = gpuMetrics;
    }

    public GPUMetric[] getGpuMetrics() {
        return gpuMetrics;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
