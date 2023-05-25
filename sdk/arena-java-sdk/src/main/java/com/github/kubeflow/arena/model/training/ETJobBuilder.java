package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.model.fields.BoolField;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class ETJobBuilder extends JobBuilder {
    public ETJobBuilder() {
        super(TrainingJobType.ETTrainingJob);
    }

    public ETJobBuilder minWorkers(int workers) {
        this.options.add(new StringField("--min-workers",String.valueOf(workers)));
        return this;
    }

    public ETJobBuilder maxWorkers(int workers) {
        this.options.add(new StringField("--max-workers",String.valueOf(workers)));
        return this;
    }

    public ETJobBuilder cpu(String c) {
        this.options.add(new StringField("--cpu",c));
        return this;
    }

    public ETJobBuilder memory(String m) {
        this.options.add(new StringField("--memory",m));
        return this;
    }

    public ETJobBuilder enableSpotInstance() {
        this.options.add(new BoolField("--spot-instance"));
        return this;
    }

    public ETJobBuilder maxWaitTime(int time) {
        this.options.add(new StringField("--max-wait-time", String.valueOf(time)));
        return this;
    }


    /**
     * following functions invoke JobBuilder functions
     *
     *
     * **/


    public ETJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public ETJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public ETJobBuilder workers(int count) {
        super.workers(count);
        return this;
    }

    public ETJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public ETJobBuilder imagePullPolicy(String policy) {
        super.imagePullPolicy(policy);
        return this;
    }

    public ETJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public ETJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public ETJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public ETJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public ETJobBuilder configFiles(Map<String, String> files) {
        super.configFiles(files);
        return this;
    }

    public ETJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public ETJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public ETJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public ETJobBuilder logDir(String dir) {
        super.logDir(dir);
        return this;
    }

    public ETJobBuilder priority(String priority) {
        super.priority(priority);
        return  this;
    }

    public ETJobBuilder enableRDMA() {
        super.enableRDMA();
        return  this;
    }

    public ETJobBuilder syncImage(String image) {
        super.syncImage(image);
        return  this;
    }

    public ETJobBuilder syncMode(String mode) {
        super.syncMode(mode);
        return  this;
    }

    public ETJobBuilder syncSource(String source) {
        super.syncSource(source);
        return  this;
    }

    public ETJobBuilder enableTensorboard() {
        super.enableTensorboard();
        return this;
    }

    public ETJobBuilder tensorboardImage(String image) {
        super.tensorboardImage(image);
        return this;
    }

    public ETJobBuilder workingDir(String dir) {
        super.workingDir(dir);
        return this;
    }

    public ETJobBuilder retryCount(int count) {
        super.retryCount(count);
        return this;
    }

    public ETJobBuilder enableCoscheduling() {
        super.enableCoscheduling();
        return this;
    }

    public ETJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
