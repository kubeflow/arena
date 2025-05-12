package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import com.github.kubeflow.arena.model.fields.*;

public abstract class JobBuilder {

    protected String jobName;

    protected TrainingJobType jobType;

    protected ArrayList<Field> options;

    protected String command;

    public JobBuilder(TrainingJobType jobType){
        this.jobType = jobType;
        this.options =  new ArrayList<Field>();
    }

    public TrainingJob build() throws ArenaException {
        List<String> args = new ArrayList<>();
        for (int i = 0;i < this.options.size();i++) {
            Field f = this.options.get(i);
            f.validate();
            for(int j = 0;j < f.options().size();j++) {
                args.add(f.options().get(j));
            }
        }
        return new TrainingJob(this.jobName,this.jobType,args,this.command);
    }

    public JobBuilder name(String name) {
        this.jobName = name;
        this.options.add(new StringField("--name",name));
        return this;
    }

    public JobBuilder image(String image) {
        this.options.add(new StringField("--image",image));
        return this;
    }

    public JobBuilder workers(int count) {
        this.options.add(new StringField("--workers",String.valueOf(count)));
        return this;
    }

    public JobBuilder imagePullSecrets(ArrayList<String> secrets) {
        this.options.add(new StringListField("--image-pull-secret",secrets));
        return this;
    }

    public JobBuilder imagePullPolicy(String policy) {
        this.options.add(new StringField("--image-pull-policy", policy));
        return this;
    }

    public JobBuilder gpus(int count) {
        this.options.add(new StringField("--gpus",String.valueOf(count)));
        return this;
    }

    public JobBuilder envs(Map<String, String> envs) {
        this.options.add(new StringMapField("--env",envs,"="));
        return this;
    }

    public JobBuilder nodeSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--selector",selectors,"="));
        return this;
    }

    public JobBuilder tolerations(ArrayList<String> tolerations) {
        this.options.add(new StringListField("--toleration",tolerations));
        return this;
    }

    public JobBuilder configFiles(Map<String, String> files) {
        this.options.add(new StringMapField("--config-file",files,":"));
        return this;
    }

    public JobBuilder labels(Map<String, String> labels) {
        this.options.add(new StringMapField("--label", labels,"="));
        return this;
    }

    public JobBuilder devices(Map<String, String> devices) {
        this.options.add(new StringMapField("--device", devices,"="));
        return this;
    }

    public JobBuilder annotations(Map<String, String> annotations) {
        this.options.add(new StringMapField("--annotation",annotations,"="));
        return this;
    }

    public JobBuilder datas(Map<String, String> datas) {
        this.options.add(new StringMapField("--data",datas,":"));
        return this;
    }

    public JobBuilder dataDirs(Map<String, String> dataDirs) {
        this.options.add(new StringMapField("--data-dir",dataDirs,":"));
        return  this;
    }

    public JobBuilder logDir(String dir) {
        this.options.add(new StringField("--logdir",dir));
        return this;
    }

    public JobBuilder priority(String priority) {
        this.options.add(new StringField("--priority",priority));
        return  this;
    }

    public JobBuilder enableRDMA() {
        this.options.add(new BoolField("--rdma"));
        return  this;
    }

    public JobBuilder syncImage(String image) {
        this.options.add(new StringField("--sync-image",image));
        return  this;
    }

    public JobBuilder syncMode(String mode) {
        this.options.add(new StringField("--sync-mode",mode));
        return  this;
    }

    public JobBuilder syncSource(String source) {
        this.options.add(new StringField("--sync-source",source));
        return  this;
    }

    public JobBuilder enableTensorboard() {
        this.options.add(new BoolField("--tensorboard"));
        return this;
    }

    public JobBuilder tensorboardImage(String image) {
        this.options.add(new StringField("--tensorboard-image",image));
        return this;
    }

    public JobBuilder workingDir(String dir) {
        this.options.add(new StringField("--working-dir",dir));
        return this;
    }

    public JobBuilder retryCount(int count) {
        this.options.add(new StringField("--retry",String.valueOf(count)));
        return this;
    }

    public JobBuilder enableCoscheduling() {
       this.options.add(new BoolField("--gang"));
       return this;
    }

    public JobBuilder shell(String shell) {
        this.options.add(new StringField("--shell",String.valueOf(shell)));
        return this;
    }

    public JobBuilder command(String command) {
        this.command = command;
        return this;
    }
}


