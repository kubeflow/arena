package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class GPUTopologyNodeDevice {
    private String id;
    private String status;
    private Boolean healthy;

    public void setStatus(String status) {
        this.status = status;
    }

    public String getStatus() {
        return status;
    }

    public Boolean getHealthy() {
        return healthy;
    }

    public void setHealthy(Boolean healthy) {
        this.healthy = healthy;
    }

    public String getId() {
        return id;
    }

    public void setId(String gpuIndex) {
        this.id = gpuIndex;
    }
    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
