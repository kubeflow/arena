package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUShareNode extends Node {
    private int totalGPUs;
    private int unhealthyGPUs;
    private int allocatedGPUs;
    private double totalGPUMemory;
    private double allocatedGPUMemory;
    private GPUShareNodePod[] instances;
    private GPUShareNodeDevice[] devices;
    private GPUMetric[] gpuMetrics;

    public GPUShareNode() {
        super();
    }

    public void setAllocatedGPUMemory(double allocatedGPUMemory) {
        this.allocatedGPUMemory = allocatedGPUMemory;
    }

    public double getAllocatedGPUMemory() {
        return allocatedGPUMemory;
    }

    public void setTotalGPUMemory(double totalGPUMemory) {
        this.totalGPUMemory = totalGPUMemory;
    }

    public double getTotalGPUMemory() {
        return totalGPUMemory;
    }

    public GPUShareNodeDevice[] getDevices() {
        return devices;
    }
    public void setDevices(GPUShareNodeDevice[] devices) {
        this.devices = devices;
    }

    public GPUShareNodePod[] getInstances() {
        return instances;
    }
    public void setInstances(GPUShareNodePod[] instances) {
        this.instances = instances;
    }

    public void setTotalGPUs(int totalGPUs) {
        this.totalGPUs = totalGPUs;
    }

    public int getTotalGPUs() {
        return this.totalGPUs;
    }

    public void setUnhealthyGPUs(int unhealthyGPUs) {
        this.unhealthyGPUs = unhealthyGPUs;
    }

    public int getUnhealthyGPUs() {
        return unhealthyGPUs;
    }

    public void setAllocatedGPUs(int allocatedGPUs) {
        this.allocatedGPUs = allocatedGPUs;
    }

    public int getAllocatedGPUs() {
        return allocatedGPUs;
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
