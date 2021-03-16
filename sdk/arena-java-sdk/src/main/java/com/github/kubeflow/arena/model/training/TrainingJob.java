package com.github.kubeflow.arena.model.training;
import com.github.kubeflow.arena.enums.TrainingJobType;

import java.util.ArrayList;
import org.apache.commons.lang3.StringUtils;

public class TrainingJob {
    private String name;
    private TrainingJobType jobType;
    private ArrayList<String> args;
    private String command;

    public TrainingJob(String name,TrainingJobType jobType,ArrayList<String> args,String command) {
        this.name = name;
        this.jobType = jobType;
        this.args = args;
        this.command = command;
    }

    public String name() {
        return this.name;
    }

    public TrainingJobType getType() {
        return  this.jobType;
    }

    public ArrayList<String> getArgs() {
        return  this.args;
    }

    public String getCommand() {
        if (this.command == null) {
            return "";
        }
        return this.command;
    }
}
