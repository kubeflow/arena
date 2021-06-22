package com.github.kubeflow.arena.model.serving;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.enums.ServingJobType;

public class ServingJobInfo {
    private String uuid;
    private String name;
    private String namespace;
    private ServingJobType type;
    private String version;
    private String age;
    private String ip;
    private int desiredInstances;
    private int availableInstances;
    private float requestCPUs;
    private int requestGPUs;
    private int requestGPUMemory;
    private Instance[] instances;
    private Endpoint[] endpoints;

    public String getUuid() {
        return uuid;
    }

    public void setUuid(String uuid) {
        this.uuid = uuid;
    }

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

    public void setType(ServingJobType type) {
        this.type = type;
    }

    public String getAge() {
        return this.age;
    }

    public void setAge(String duration) {
        this.age = duration;
    }

    public ServingJobType getType() {
        return this.type;
    }

    public void setType(String jobType) {
        this.type = ServingJobType.getByAlias(jobType);
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public String getVersion() {
        return version;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getIp() {
        return ip;
    }

    public void setAvailableInstances(int availableInstances) {
        this.availableInstances = availableInstances;
    }

    public int getAvailableInstances() {
        return availableInstances;
    }

    public void setDesiredInstances(int desiredInstances) {
        this.desiredInstances = desiredInstances;
    }

    public int getDesiredInstances() {
        return desiredInstances;
    }

    public float getRequestCPUs() {
        return requestCPUs;
    }

    public void setRequestCPUs(float requestCPUs) {
        this.requestCPUs = requestCPUs;
    }

    public int getRequestGPUs() {
        return this.requestGPUs;
    }

    public void setRequestGPUs(int requestGPUs) {
        this.requestGPUs = requestGPUs;
    }

    public void setEndpoints(Endpoint[] endpoints) {
        this.endpoints = endpoints;
    }

    public Endpoint[] getEndpoints() {
        return endpoints;
    }

    public void setRequestGPUMemory(int requestGPUMemory) {
        this.requestGPUMemory = requestGPUMemory;
    }

    public int getRequestGPUMemory() {
        return requestGPUMemory;
    }

    public Instance[] getInstances() {
        return this.instances;
    }

    public void setInstances(Instance[] instances) {
        this.instances = instances;
    }

    public ServingJobInfo complete(String json) {
        ServingJobInfo servingJobInfo = JSON.parseObject(json, ServingJobInfo.class);
        for (int i = 0; i < servingJobInfo.instances.length; i++) {
            servingJobInfo.instances[i].setNamespace(servingJobInfo.getNamespace());
            servingJobInfo.instances[i].setOwner(servingJobInfo.getName());
            servingJobInfo.instances[i].setOwnerType(servingJobInfo.getType());
            servingJobInfo.instances[i].setOwnerVersion(servingJobInfo.getVersion());
        }
        return servingJobInfo;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this, true);
    }
}
