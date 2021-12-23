package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;

import java.util.List;


public class ServingJob {
    private String name;
    private ServingJobType jobType;
    private String version;
    private List<String> args;
    private String command;

    public ServingJob(String name, ServingJobType jobType, String version, List<String> args, String command) {
        this.name = name;
        this.jobType = jobType;
        this.version = version;
        this.args = args;
        this.command = command;
    }

    public String name() {
        return this.name;
    }

    public ServingJobType getType() {
        return this.jobType;
    }

    public List<String> getArgs() {
        return this.args;
    }

    public String version() {
        return this.version;
    }

    public String getCommand() {
        if (this.command == null) {
            return "";
        }
        return this.command;
    }
}
