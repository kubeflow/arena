package com.github.kubeflow.arena.client;

import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.utils.Utils;
import io.kubernetes.client.openapi.ApiClient;

import java.io.File;
import java.io.IOException;

public class ArenaClient {
    private String namespace;
    private String kubeConfig = "";
    private ApiClient apiClient = null;
    private String loglevel;
    private String arenaNamespace;
    private static String DefaultKubeConfig = "";
    private static String DefaultArenaNamespace = "arena-system";
    private static String DefaultLogLevel = "info";
    private static String DefaultNamespace = "default";

    public ArenaClient(String kubeConfig, String namespace, String loglevel, String arenaNamespace) throws ArenaException {
        if (kubeConfig == null || kubeConfig.length() == 0) {
            String defaultKubeconfigPath = System.getProperty("user.home") + "/.kube/config";
            File defaultKubeconfig = new File(defaultKubeconfigPath);
            if (defaultKubeconfig.exists()) {
                kubeConfig = defaultKubeconfigPath;
            }
        }
        this.kubeConfig = kubeConfig;
        if (namespace != null && namespace.length() != 0) {
            this.namespace = namespace;
        }
        if (loglevel != null && loglevel.length() != 0) {
            this.loglevel = loglevel;
        }
        if (arenaNamespace != null && arenaNamespace.length() != 0) {
            this.arenaNamespace = arenaNamespace;
        }
        try {
            this.apiClient = new Utils().createK8sClient(this.kubeConfig);
        } catch (IOException e) {
            throw new ArenaException(ArenaErrorEnum.UNKNOWN, "create arena client failed, " + e.getMessage());
        }
    }

    public ArenaClient() throws ArenaException {
        this(DefaultKubeConfig, DefaultNamespace, DefaultLogLevel, DefaultArenaNamespace);
    }

    public ArenaClient(String kubeConfig) throws ArenaException {
        this(kubeConfig, DefaultNamespace, DefaultLogLevel, DefaultArenaNamespace);
    }

    public ArenaClient(String kubeConfig, String namespace) throws ArenaException {
        this(kubeConfig, namespace, DefaultLogLevel, DefaultArenaNamespace);
    }

    public ArenaClient(String kubeConfig, String namespace, String loglevel) throws ArenaException {
        this(kubeConfig, namespace, loglevel, DefaultArenaNamespace);
    }

    public ApiClient getApiClient() {
        return this.apiClient;
    }

    public TrainingClient training() {
        return new TrainingClient(this.kubeConfig, this.namespace, this.loglevel, this.arenaNamespace);
    }

    public ServingClient serving() {
        return new ServingClient(this.kubeConfig, this.namespace, this.loglevel, this.arenaNamespace);
    }

    public NodeClient nodes() {
        return new NodeClient(this.kubeConfig, this.namespace, this.loglevel, this.arenaNamespace);
    }

}
