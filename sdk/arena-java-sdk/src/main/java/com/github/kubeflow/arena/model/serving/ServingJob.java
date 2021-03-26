package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;

import java.util.ArrayList;

public class ServingJob {
    private String name;
    private ServingJobType jobType;
    private String version;
    private ArrayList<String> args;
    private String command;

    public ServingJob(String name,ServingJobType jobType,String version,ArrayList<String> args,String command) {
        this.name = name;
        this.jobType = jobType;
        this.args = args;
        this.command = command;
    }

    public String name() {
        return this.name;
    }

    public ServingJobType getType() {
        return  this.jobType;
    }

    public ArrayList<String> getArgs() {
        return  this.args;
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
