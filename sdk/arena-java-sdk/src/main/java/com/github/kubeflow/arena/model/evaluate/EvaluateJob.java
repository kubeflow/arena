package com.github.kubeflow.arena.model.evaluate;

import java.util.List;

public class EvaluateJob {

    private String name;
    private List<String> args;
    private String command;

    public EvaluateJob(String name, List<String> args, String command) {
        this.name = name;
        this.args = args;
        this.command = command;
    }

    public String name() {
        return this.name;
    }

    public List<String> getArgs() {
        return this.args;
    }

    public String getCommand() {
        if (this.command == null) {
            return "";
        }
        return this.command;
    }

}
