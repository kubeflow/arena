package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.fields.Field;
import com.github.kubeflow.arena.model.fields.StringField;
import com.github.kubeflow.arena.model.fields.StringListField;

import java.util.ArrayList;
import java.util.Map;

public class VolcanoJobBuilder {
    protected String jobName;

    protected TrainingJobType jobType;

    protected ArrayList<Field> options;

    protected String command;
    public VolcanoJobBuilder() {
        this.jobType = TrainingJobType.VolcanoTrainingJob;
        this.options =  new ArrayList<Field>();
    }

    public VolcanoJobBuilder name(String name) {
        this.options.add(new StringField("--name",name));
        return this;
    }

    public VolcanoJobBuilder queue(String m) {
        this.options.add(new StringField("--queue",m));
        return this;
    }

    public VolcanoJobBuilder minAvailable(String name) {
        this.options.add(new StringField("--min-available",name));
        return this;
    }

    public VolcanoJobBuilder taskCPU(String cpu) {
        this.options.add(new StringField("--task-cpu",cpu));
        return this;
    }

    public VolcanoJobBuilder schedulerName(String name) {
        this.options.add(new StringField("--scheduler-name",name));
        return this;
    }

    public VolcanoJobBuilder taskImages(ArrayList<String> images) {
        this.options.add(new StringListField("--task-images",images));
        return this;
    }

    public VolcanoJobBuilder taskMemory(String memory) {
        this.options.add(new StringField("--task-memory",memory));
        return this;
    }

    public VolcanoJobBuilder taskName(String name) {
        this.options.add(new StringField("--task-name",name));
        return this;
    }

    public VolcanoJobBuilder taskPort(String port) {
        this.options.add(new StringField("--task-port",port));
        return this;
    }


    public VolcanoJobBuilder taskReplicas(int count) {
        this.options.add(new StringField("--task-replicas",String.valueOf(count)));
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

    public VolcanoJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
