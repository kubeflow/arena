package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class HorovodJobBuilder extends JobBuilder {

    public HorovodJobBuilder() {
        super(TrainingJobType.HorovodTrainingJob);
    }


    public HorovodJobBuilder sshPort(int port) {
        this.options.add(new StringField("--ssh-port",String.valueOf(port)));
        return this;
    }

    public HorovodJobBuilder cpu(String c) {
        this.options.add(new StringField("--cpu",c));
        return this;
    }

    public HorovodJobBuilder memory(String m) {
        this.options.add(new StringField("--memory",m));
        return this;
    }

    /**
     * following functions invoke JobBuilder functions
     *
     *
     * **/


    public HorovodJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public HorovodJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public HorovodJobBuilder workers(int count) {
        super.workers(count);
        return this;
    }

    public HorovodJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public HorovodJobBuilder imagePullPolicy(String policy) {
        super.imagePullPolicy(policy);
        return this;
    }

    public HorovodJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public HorovodJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public HorovodJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public HorovodJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public HorovodJobBuilder configFiles(Map<String, String> files) {
        super.configFiles(files);
        return this;
    }

    public HorovodJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public HorovodJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public HorovodJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return this;
    }

    public HorovodJobBuilder logDir(String dir) {
        super.logDir(dir);
        return this;
    }

    public HorovodJobBuilder priority(String priority) {
        super.priority(priority);
        return this;
    }

    public HorovodJobBuilder enableRDMA() {
        super.enableRDMA();
        return this;
    }

    public HorovodJobBuilder syncImage(String image) {
        super.syncImage(image);
        return this;
    }

    public HorovodJobBuilder syncMode(String mode) {
        super.syncMode(mode);
        return this;
    }

    public HorovodJobBuilder syncSource(String source) {
        super.syncSource(source);
        return this;
    }

    public HorovodJobBuilder enableTensorboard() {
        super.enableTensorboard();
        return this;
    }

    public HorovodJobBuilder tensorboardImage(String image) {
        super.tensorboardImage(image);
        return this;
    }

    public HorovodJobBuilder workingDir(String dir) {
        super.workingDir(dir);
        return this;
    }

    public HorovodJobBuilder retryCount(int count) {
        super.retryCount(count);
        return this;
    }

    public HorovodJobBuilder enableCoscheduling() {
        super.enableCoscheduling();
        return this;
    }

    public HorovodJobBuilder shell(String shell) {
        super.shell(shell);
        return this;
    }

    public HorovodJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
