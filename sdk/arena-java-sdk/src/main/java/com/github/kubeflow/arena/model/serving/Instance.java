package com.github.kubeflow.arena.model.serving;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.exceptions.ArenaException;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.Configuration;
import io.kubernetes.client.ApiClient;
import com.github.kubeflow.arena.model.common.*;
import com.squareup.okhttp.Call;
import com.squareup.okhttp.Response;
import io.kubernetes.client.models.V1Pod;

import java.io.InputStream;

import java.io.IOException;
import java.util.concurrent.TimeUnit;
import com.github.kubeflow.arena.enums.ServingJobType;
import  com.github.kubeflow.arena.enums.ArenaErrorEnum;

public class Instance {

    private String owner;
    private ServingJobType ownerType;
    private String ownerVersion;
    private String name;
    private String namespace;
    private String age;
    private String status;
    private int readyContainers;
    private int totalContainers;
    private int restartCount;
    private String nodeIP;
    private String nodeName;
    private String ip;
    private int requestGPUs;
    private int requestGPUMemory;

    public String getOwner() {
        return owner;
    }
    public void setOwner(String owner) {
        this.owner = owner;
    }


    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }
    public String getNamespace() {
        return namespace;
    }

    public void setOwnerType(ServingJobType ownerType) {
        this.ownerType = ownerType;
    }
    public ServingJobType getOwnerType() {
        return ownerType;
    }

    public String getName() {
        return name;
    }
    public void   setName(String name) {
        this.name = name;
    }

    public String getOwnerVersion() {
        return this.ownerVersion;
    }
    public void setOwnerVersion(String version) {
       this.ownerVersion = version;
    }

    public String getAge() {
        return age;
    }
    public void   setAge(String age) {
        this.age = age;
    }

    public String getStatus() {
        return status;
    }
    public void   setStatus(String status) {
        this.status = status;
    }

    public int getReadyContainers() {
        return readyContainers;
    }
    public void setReadyContainers(int count) {
        this.readyContainers = count;
    }

    public int getTotalContainers() {
        return this.totalContainers;
    }
    public void setTotalContainers(int count) {
        this.totalContainers = count;
    }

    public int getRestartCount() {
        return this.restartCount;
    }
    public void setRestartCount(int count) {
        this.restartCount = count;
    }

    public String getNodeName() {
        return nodeName;
    }
    public void   setNodeName(String node) {
        this.nodeName = node;
    }

    public void setNodeIP(String nodeIP) {
        this.nodeIP = nodeIP;
    }
    public String getNodeIP() {
        return this.nodeIP;
    }

    public String getIp() {
        return this.ip;
    }
    public void setIp(String ip) {
        this.ip = ip;
    }

    public int getRequestGPUMemory() {
        return this.requestGPUMemory;
    }
    public void setRequestGPUMemory(int count) {
        this.requestGPUMemory = count;
    }

    public void setRequestGPUs(int requestGPUs) {
        this.requestGPUs = requestGPUs;
    }

    public int getRequestGPUs() {
        return this.requestGPUs;
    }

    public InputStream getLog(Logger logger) throws ArenaException, IOException  {
        ApiClient apiClient = Configuration.getDefaultApiClient();
        int defaultTimeout = apiClient.getHttpClient().getReadTimeout()/1000;
        if (logger.getFollowTimeout() != null && logger.getFollowTimeout() != 0) {
            apiClient.getHttpClient().setReadTimeout(logger.getFollowTimeout(),TimeUnit.SECONDS);
        }
        CoreV1Api coreClient = new CoreV1Api(apiClient);
        V1Pod pod;
        try{
            //CoreV1Api api = new CoreV1Api(apiClient);
            pod = coreClient.readNamespacedPod(this.name, this.namespace, null, null, false);
        }catch (ApiException e) {
            apiClient.getHttpClient().setReadTimeout(defaultTimeout,TimeUnit.SECONDS);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS,e.getMessage());
        }
        if (pod.getSpec() == null) {
            apiClient.getHttpClient().setReadTimeout(defaultTimeout,TimeUnit.SECONDS);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS,"pod.spec is null and container isn't specified.");
        }
        if (pod.getSpec().getContainers() == null || pod.getSpec().getContainers().size() < 1) {
            apiClient.getHttpClient().setReadTimeout(defaultTimeout,TimeUnit.SECONDS);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS,"pod.spec.containers has no containers");
        }
        String container = pod.getSpec().getContainers().get(0).getName();
        String namespace = this.namespace;
        String name = this.name;
        Integer sinceSeconds = logger.getSinceSeconds();
        Integer tailLines = logger.getTailLines();
        boolean timestamps = logger.getTimestamps();
        boolean follow = logger.getFollow();
        if (logger.getContainer() != null) {
            container = logger.getContainer();
        }
        Response response;
        try {
            Call call = coreClient.readNamespacedPodLogCall(
                    name,
                    namespace,
                    container,
                    follow,
                    null,
                    "false",
                    false,
                    sinceSeconds,
                    tailLines,
                    timestamps,
                    null, null);
            response = call.execute();
        }catch (ApiException e) {
            apiClient.getHttpClient().setReadTimeout(defaultTimeout,TimeUnit.SECONDS);
            throw  new ArenaException(ArenaErrorEnum.TRAINING_LOGS,e.getMessage());
        }
        if (!response.isSuccessful()) {
            apiClient.getHttpClient().setReadTimeout(defaultTimeout,TimeUnit.SECONDS);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS,"Logs request failed: " + response.code());
        }
        apiClient.getHttpClient().setReadTimeout(defaultTimeout,TimeUnit.SECONDS);
        return response.body().byteStream();
    }


    @Override
    public String toString() {
       return JSON.toJSONString(this,true);
    }
}
