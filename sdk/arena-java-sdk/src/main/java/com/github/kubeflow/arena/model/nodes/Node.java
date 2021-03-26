package com.github.kubeflow.arena.model.nodes;

import com.github.kubeflow.arena.enums.NodeType;
public abstract class Node {
    protected String name;
    protected String ip;
    protected String status;
    protected String role;
    protected NodeType type;
    protected String description;

    public String getName() {
        return name;
    }
    public void setName(String name) {
        this.name = name;
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public NodeType getType() {
        return type;
    }

    public void setType(String type) {
        this.type = NodeType.getByAlias(type);
    }

    public String getRole() {
        return role;
    }

    public void setRole(String role) {
        this.role = role;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getDescription() {
        return this.description;
    }
}
