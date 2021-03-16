package com.github.kubeflow.arena.model.training;

import com.alibaba.fastjson.JSON;

public class GPUMetric {
    private double gpuDutyCycle;
    private double usedGPUMemory;
    private double totalGPUMemory;

    public double getGpuDutyCycle() {
        return this.gpuDutyCycle;
    }

    public void setGpuDutyCycle(double gpuDutyCycle) {
        this.gpuDutyCycle = gpuDutyCycle;
    }

    public double getTotalGPUMemory() {
        return totalGPUMemory;
    }

    public void setTotalGPUMemory(double totalGPUMemory) {
        this.totalGPUMemory = totalGPUMemory;
    }

    public double getUsedGPUMemory() {
        return usedGPUMemory;
    }

    public void setUsedGPUMemory(double usedGPUMemory) {
        this.usedGPUMemory = usedGPUMemory;
    }
    @Override
    public String toString() {
        return JSON.toJSONString(this, true);
    }
}
