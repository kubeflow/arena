package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.fields.Field;
import com.github.kubeflow.arena.model.fields.StringField;
import com.github.kubeflow.arena.model.fields.StringMapField;

import java.util.ArrayList;
import java.util.Map;

public abstract class ScaleETJobBuilder {

    protected String jobName;

    protected TrainingJobType jobType;

    protected ArrayList<Field> options;

    protected String command = "";

    public  ScaleETJobBuilder(TrainingJobType jobType) {
        this.jobType = jobType;
    }

    public ScaleETJobBuilder name(String name) {
        this.jobName = name;
        this.options.add(new StringField("--name",name));
        return this;
    }

    public ScaleETJobBuilder timeout(String timeout) {
        this.options.add(new StringField("--timeout",String.valueOf(timeout)));
        return this;
    }

    public ScaleETJobBuilder retry(int retry) {
        this.options.add(new StringField("--retry",String.valueOf(retry)));
        return this;
    }

    public ScaleETJobBuilder count(int count) {
        this.options.add(new StringField("--count",String.valueOf(count)));
        return this;
    }

    public ScaleETJobBuilder script(String script) {
        this.options.add(new StringField("--script",String.valueOf(script)));
        return this;
    }

    public ScaleETJobBuilder envs(Map<String, String> envs) {
        this.options.add(new StringMapField("--env",envs,"="));
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
