package com.github.kubeflow.arena.model.serving;

import com.alibaba.fastjson.JSON;

public class Endpoint {
    private String name;
    private int port;
    private int nodePort;

    public void setName(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }

    public void setNodePort(int nodePort) {
        this.nodePort = nodePort;
    }

    public int getNodePort() {
        return nodePort;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public int getPort() {
        return port;
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this, true);
    }
}
