package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class PytorchJobBuilder extends JobBuilder {

    public PytorchJobBuilder() {
        super(TrainingJobType.PytorchTrainingJob);
    }

    public PytorchJobBuilder cpu(String c) {
        this.options.add(new StringField("--cpu", c));
        return this;
    }

    public PytorchJobBuilder memory(String m) {
        this.options.add(new StringField("--memory", m));
        return this;
    }

    public PytorchJobBuilder cleanPodPolicy(String policy) {
        this.options.add(new StringField("--clean-task-policy", policy));
        return this;
    }

    // Specifies the duration since startTime during which the job can remain active before it is terminated(e.g. '5s', '1m', '2h22m').
    public PytorchJobBuilder runningTimeout(String t) {
        this.options.add(new StringField("--running-timeout", t));
        return this;
    }

    // Defines the TTL for cleaning up finished PytorchJobs(e.g. '5s', '1m', '2h22m'). Defaults to infinite."
    public PytorchJobBuilder ttlAfterFinished(String t) {
        this.options.add(new StringField("--ttl-after-finished", t));
        return this;
    }

    /**
     * following functions invoke JobBuilder functions
     **/


    public PytorchJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public PytorchJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public PytorchJobBuilder workers(int count) {
        super.workers(count);
        return this;
    }

    public PytorchJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public PytorchJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public PytorchJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public PytorchJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public PytorchJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public PytorchJobBuilder configFiles(Map<String, String> files) {
        super.configFiles(files);
        return this;
    }

    public PytorchJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public PytorchJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public PytorchJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return this;
    }

    public PytorchJobBuilder logDir(String dir) {
        super.logDir(dir);
        return this;
    }

    public PytorchJobBuilder priority(String priority) {
        super.priority(priority);
        return this;
    }

    public PytorchJobBuilder enableRDMA() {
        super.enableRDMA();
        return this;
    }

    public PytorchJobBuilder syncImage(String image) {
        super.syncImage(image);
        return this;
    }

    public PytorchJobBuilder syncMode(String mode) {
        super.syncMode(mode);
        return this;
    }

    public PytorchJobBuilder syncSource(String source) {
        super.syncSource(source);
        return this;
    }

    public PytorchJobBuilder enableTensorboard() {
        super.enableTensorboard();
        return this;
    }

    public PytorchJobBuilder tensorboardImage(String image) {
        super.tensorboardImage(image);
        return this;
    }

    public PytorchJobBuilder workingDir(String dir) {
        super.workingDir(dir);
        return this;
    }

    public PytorchJobBuilder retryCount(int count) {
        super.retryCount(count);
        return this;
    }

    public PytorchJobBuilder enableCoscheduling() {
        super.enableCoscheduling();
        return this;
    }

    public PytorchJobBuilder shell(String shell) {
        super.shell(shell);
        return this;
    }

    public PytorchJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
