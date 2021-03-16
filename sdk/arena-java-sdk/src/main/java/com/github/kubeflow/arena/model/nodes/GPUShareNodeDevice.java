package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUShareNodeDevice {
    private String id;
    private double allocatedGPUMemory;
    private double totalGPUMemory;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public double getTotalGPUMemory() {
        return totalGPUMemory;
    }

    public void setTotalGPUMemory(double totalGPUMemory) {
        this.totalGPUMemory = totalGPUMemory;
    }

    public double getAllocatedGPUMemory() {
        return allocatedGPUMemory;
    }

    public void setAllocatedGPUMemory(double allocatedGPUMemory) {
        this.allocatedGPUMemory = allocatedGPUMemory;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
