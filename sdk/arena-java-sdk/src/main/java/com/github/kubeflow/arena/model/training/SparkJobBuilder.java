package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.fields.Field;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class SparkJobBuilder {

    protected String jobName;

    protected TrainingJobType jobType;

    protected ArrayList<Field> options;

    protected String command;

    public SparkJobBuilder() {
        this.jobType = TrainingJobType.SparkTrainingJob;
        this.options =  new ArrayList<Field>();
    }

    public SparkJobBuilder name(String name) {
        this.options.add(new StringField("--name",name));
        return this;
    }

    public SparkJobBuilder image(String image) {
        this.options.add(new StringField("--image",image));
        return this;
    }

    public SparkJobBuilder replicas(int replicas) {
        this.options.add(new StringField("--replicas",String.valueOf(replicas)));
        return this;
    }

    public SparkJobBuilder mainClass(String mainClass) {
        this.options.add(new StringField("--main-class",mainClass));
        return this;
    }

    public SparkJobBuilder jar(String jar) {
        this.options.add(new StringField("--jar",jar));
        return this;
    }

    public SparkJobBuilder driverCPU(int count) {
        this.options.add(new StringField("--driver-cpu-request",String.valueOf(count)));
        return this;
    }

    public SparkJobBuilder driverMemory(int count) {
        this.options.add(new StringField("--driver-memory-request",String.valueOf(count)));
        return this;
    }


    public SparkJobBuilder executorMemory(int count) {
        this.options.add(new StringField("--executor-memory-request",String.valueOf(count)));
        return this;
    }

    public SparkJobBuilder executorCPU(int count) {
        this.options.add(new StringField("--executor-cpu-request",String.valueOf(count)));
        return this;
    }

    public SparkJobBuilder command(String command) {
        this.command = command;
        return this;
    }


    public TrainingJob build() throws ArenaException {
        ArrayList<String> args = new ArrayList<String>();
        for (int i = 0;i < this.options.size();i++) {
            Field f = this.options.get(i);
            f.validate();
            for(int j = 0;j < f.options().size();j++) {
                args.add(f.options().get(j));
            }
        }
        return new TrainingJob(this.jobName,this.jobType,args,this.command);
    }
}
