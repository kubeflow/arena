package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUMetric {
    private String id;
    private String uuid;
    private String status;
    private double gpuDutyCycle;
    private double usedGPUMemory;
    private double totalGPUMemory;

    public void setId(String id) {
        this.id = id;
    }

    public String getId() {
        return id;
    }

    public void setUuid(String uuid) {
        this.uuid = uuid;
    }

    public String getUuid() {
        return uuid;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getStatus() {
        return status;
    }

    public void setGpuDutyCycle(double gpuDutyCycle) {
        this.gpuDutyCycle = gpuDutyCycle;
    }

    public double getGpuDutyCycle() {
        return gpuDutyCycle;
    }

    public void setUsedGPUMemory(double usedGPUMemory) {
        this.usedGPUMemory = usedGPUMemory;
    }

    public double getUsedGPUMemory() {
        return usedGPUMemory;
    }

    public void setTotalGPUMemory(double totalGPUMemory) {
        this.totalGPUMemory = totalGPUMemory;
    }

    public double getTotalGPUMemory() {
        return totalGPUMemory;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
