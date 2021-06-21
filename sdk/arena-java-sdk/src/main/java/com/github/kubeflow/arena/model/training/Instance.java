package com.github.kubeflow.arena.model.training;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.common.Logger;
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.Configuration;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1Pod;
import okhttp3.Call;
import okhttp3.Response;

import java.io.IOException;
import java.io.InputStream;
import java.util.Map;

public class Instance {

    private String owner;
    private TrainingJobType ownerType;
    private String name;
    private String ip;
    private String age;
    private String status;
    private String node;
    private String nodeIP;
    private String namespace;
    private int requestGPUs;
    private boolean chief;
    private Map<String, GPUMetric> gpuMetrics;

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

    public void setOwnerType(TrainingJobType ownerType) {
        this.ownerType = ownerType;
    }

    public TrainingJobType getOwnerType() {
        return ownerType;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getAge() {
        return age;
    }

    public void setAge(String age) {
        this.age = age;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getIp() {
        return ip;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getNode() {
        return node;
    }

    public void setNode(String node) {
        this.node = node;
    }

    public void setNodeIP(String nodeIP) {
        this.nodeIP = nodeIP;
    }

    public String getNodeIP() {
        return this.nodeIP;
    }

    public boolean isChief() {
        return this.chief;
    }

    public void setChief(boolean chief) {
        this.chief = chief;
    }

    public void setRequestGPUs(int requestGPUs) {
        this.requestGPUs = requestGPUs;
    }

    public int getRequestGPUs() {
        return this.requestGPUs;
    }

    public void setGpuMetrics(Map<String, GPUMetric> gpuMetrics) {
        this.gpuMetrics = gpuMetrics;
    }

    public Map<String, GPUMetric> getGpuMetrics() {
        return this.gpuMetrics;
    }

    public InputStream getLog(Logger logger) throws ArenaException, IOException {
        ApiClient apiClient = Configuration.getDefaultApiClient();
        int defaultTimeout = apiClient.getReadTimeout();
        if (logger.getFollowTimeout() != null && logger.getFollowTimeout() != 0) {
            apiClient.setReadTimeout(logger.getFollowTimeout() * 1000);
        }
        CoreV1Api coreClient = new CoreV1Api(apiClient);
        V1Pod pod;
        try {
            //CoreV1Api api = new CoreV1Api(apiClient);
            pod = coreClient.readNamespacedPod(this.name, this.namespace, null, null, false);
        } catch (ApiException e) {
            apiClient.setReadTimeout(defaultTimeout);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS, e.getMessage());
        }
        if (pod.getSpec() == null) {
            apiClient.setReadTimeout(defaultTimeout);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS, "pod.spec is null and container isn't specified.");
        }
        if (pod.getSpec().getContainers() == null || pod.getSpec().getContainers().size() < 1) {
            apiClient.setReadTimeout(defaultTimeout);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS, "pod.spec.containers has no containers");
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
                    null,
                    "false",
                    false,
                    sinceSeconds,
                    tailLines,
                    timestamps,
                    null);
            response = call.execute();
        } catch (ApiException e) {
            apiClient.setReadTimeout(defaultTimeout);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS, e.getMessage());
        }
        if (!response.isSuccessful()) {
            apiClient.setReadTimeout(defaultTimeout);
            throw new ArenaException(ArenaErrorEnum.TRAINING_LOGS, "Logs request failed: " + response.code());
        }
        apiClient.setReadTimeout(defaultTimeout);
        return response.body().byteStream();
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this, true);
    }
}
