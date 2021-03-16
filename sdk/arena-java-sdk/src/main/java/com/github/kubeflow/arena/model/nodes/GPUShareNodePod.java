package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

import java.util.Map;

public class GPUShareNodePod {
    private String name;
    private String namespace;
    private int requestGPUMemory;
    private Map<String,Integer> allocation;

    public void setName(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }

    public int getRequestGPUMemory() {
        return requestGPUMemory;
    }

    public void setRequestGPUMemory(int requestGPUMemory) {
        this.requestGPUMemory = requestGPUMemory;
    }

    public Map<String, Integer> getAllocation() {
        return allocation;
    }

    public void setAllocation(Map<String, Integer> allocation) {
        this.allocation = allocation;
    }

    public String getNamespace() {
        return namespace;
    }

    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
