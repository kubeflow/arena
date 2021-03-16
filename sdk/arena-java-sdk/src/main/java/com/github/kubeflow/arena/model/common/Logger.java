package com.github.kubeflow.arena.model.common;

public class Logger {
    private String instance;
    private boolean follow = false;
    private Integer followTimeout;
    private String container;
    private Integer sinceSeconds;
    private Integer tailLines ;
    private boolean timestamps = false;
    public Logger(){}

    public Logger follow() {
        this.follow = true;
        return this;
    }
    public boolean getFollow() {
        return  this.follow;
    }

    public Logger withInstance(String name) {
        this.instance = name;
        return this;
    }

    public Logger followTimeout(int timeout) {
        this.followTimeout = timeout;
        return this;
    }
    public Integer getFollowTimeout() {
        return this.followTimeout;
    }
    public String getInstance() {
        return this.instance;
    }

    public Logger container(String name) {
        this.container = name;
        return this;
    }

    public String getContainer() {
        return this.container;
    }

    public Logger tailLines(Integer lines) {
        this.tailLines = lines;
        return this;
    }

    public Integer getTailLines() {
        return this.tailLines;
    }

    public Logger sinceSeconds(Integer seconds) {
        this.sinceSeconds = seconds;
        return this;
    }
    public Integer getSinceSeconds() {
        return this.sinceSeconds;
    }

    public Logger timestamps() {
        this.timestamps = true;
        return this;
    }
    public boolean getTimestamps() {
        return this.timestamps;
    }

}
