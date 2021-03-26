package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUTopologyNodePod {

    private String name;
    private String namespace;
    private int requestGPUs;
    private String[] allocation;
    private String[] visibleGPUs;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }


    public String getNamespace() {
        return namespace;
    }

    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }

    public void setAllocation(String[] allocation) {
        this.allocation = allocation;
    }

    public String[] getAllocation() {
        return allocation;
    }

    public void setRequestGPUs(int requestGPUs) {
        this.requestGPUs = requestGPUs;
    }

    public int getRequestGPUs() {
        return requestGPUs;
    }

    public String[] getVisibleGPUs() {
        return visibleGPUs;
    }

    public void setVisibleGPUs(String[] visibleGPUs) {
        this.visibleGPUs = visibleGPUs;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
