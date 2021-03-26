package com.github.kubeflow.arena.model.training;

import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONArray;
import com.alibaba.fastjson.JSONObject;
import com.alibaba.fastjson.annotation.JSONType;
import com.github.kubeflow.arena.enums.TrainingJobStatus;
import com.github.kubeflow.arena.enums.TrainingJobType;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.HashMap;

public class TrainingJobInfo {
    private String name;
    private String namespace;
    private String duration;
    private TrainingJobStatus status;
    private TrainingJobType trainer;
    private String tensorboard;
    private String chiefName;
    private String priority;
    private int requestGPUs;
    private int allocatedGPUs;
    private Instance[] instances;
    private long creationTimestamp;

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

    public String getDuration() {
        return this.duration;
    }
    public void setDuration(String duration) {
        this.duration = duration;
    }

    public TrainingJobStatus getStatus() {
        return this.status;
    }
    public void setStatus(String status) {
        this.status = TrainingJobStatus.getByAlias(status);
    }

    public TrainingJobType getTrainer() {
        return this.trainer;
    }
    public void setTrainer(String trainer) {
        this.trainer = TrainingJobType.getByAlias(trainer);
    }

    public String getTensorboard() {
        return this.tensorboard;
    }
    public void setTensorboard(String tensorboard) {
        this.tensorboard = tensorboard;
    }

    public String getChiefName() {
        return this.chiefName;
    }
    public void setChiefName(String name) {
        this.chiefName = name;
    }

    public String getPriority() {
        return this.priority;
    }
    public void setPriority(String priority) {
        this.priority = priority;
    }

    public int getRequestGPUs() {
        return this.requestGPUs;
    }
    public void setRequestGPUs(int requestGPUs) {
        this.requestGPUs = requestGPUs;
    }

    public int getAllocatedGPUs() {
        return this.allocatedGPUs;
    }
    public void setAllocatedGPUs(int allocatedGPUs) {
        this.allocatedGPUs = allocatedGPUs;
    }

    public Instance[] getInstances() {
        return this.instances;
    }
    public void setInstances(Instance[] instances) {
        this.instances = instances;
    }

    public long getCreationTimestamp() {
        return creationTimestamp;
    }

    public void setCreationTimestamp(long creationTimestamp) {
        this.creationTimestamp = creationTimestamp;
    }

    public TrainingJobInfo complete(String json) {
        TrainingJobInfo trainingJobInfo = JSON.parseObject(json, TrainingJobInfo.class);
        for(int i = 0;i < trainingJobInfo.instances.length;i++) {
            trainingJobInfo.instances[i].setNamespace(trainingJobInfo.getNamespace());
            trainingJobInfo.instances[i].setOwner(trainingJobInfo.getName());
            trainingJobInfo.instances[i].setOwnerType(trainingJobInfo.getTrainer());
        }
        return trainingJobInfo;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this, true);
    }
}
