package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUExclusiveNodePod {
    private String name;
    private String namespace;
    private int requestGPUs;

    public int getRequestGPUs() {
        return requestGPUs;
    }

    public void setRequestGPUs(int requestGPUs) {
        this.requestGPUs = requestGPUs;
    }

    public String getNamespace() {
        return namespace;
    }

    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
