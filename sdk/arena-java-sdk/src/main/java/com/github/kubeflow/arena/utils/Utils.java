package com.github.kubeflow.arena.utils;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.Configuration;
import io.kubernetes.client.util.KubeConfig;
import io.kubernetes.client.util.ClientBuilder;
import java.io.FileReader;
import java.io.IOException;

public class Utils {

    public ApiClient createK8sClient(String kubeconfig) throws IOException {
        ApiClient client;
        if (kubeconfig != null && kubeconfig.length() != 0) {
            client = ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(new FileReader(kubeconfig))).build();
        } else{
            client = ClientBuilder.cluster().build();
        }
        Configuration.setDefaultApiClient(client);
        return client;
    }

}