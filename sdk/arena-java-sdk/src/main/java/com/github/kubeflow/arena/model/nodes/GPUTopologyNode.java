package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUTopologyNode extends Node {
    private int totalGPUs;
    private int allocatedGPUs;
    private int unhealthyGPUs;
    private GPUTopologyNodePod[] instances;
    private GPUTopologyNodeDevice[] devices;
    private GPUMetric[] gpuMetrics;

    public GPUTopologyNode() {
        super();
    }

    public int getTotalGPUs() {
        return totalGPUs;
    }

    public void setTotalGPUs(int totalGPUs) {
        this.totalGPUs = totalGPUs;
    }

    public int getAllocatedGPUs() {
        return allocatedGPUs;
    }

    public void setAllocatedGPUs(int allocatedGPUs) {
        this.allocatedGPUs = allocatedGPUs;
    }

    public int getUnhealthyGPUs() {
        return unhealthyGPUs;
    }

    public void setUnhealthyGPUs(int unhealthyGPUs) {
        this.unhealthyGPUs = unhealthyGPUs;
    }

    public void setInstances(GPUTopologyNodePod[] instances) {
        this.instances = instances;
    }

    public GPUTopologyNodePod[] getInstances() {
        return instances;
    }

    public void setDevices(GPUTopologyNodeDevice[] devices) {
        this.devices = devices;
    }

    public GPUTopologyNodeDevice[] getDevices() {
        return devices;
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
