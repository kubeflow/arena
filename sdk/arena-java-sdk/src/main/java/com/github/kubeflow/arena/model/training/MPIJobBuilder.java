package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class MPIJobBuilder extends JobBuilder {

    public MPIJobBuilder() {
        super(TrainingJobType.MPITrainingJob);
    }

    public MPIJobBuilder cpu(String c) {
        this.options.add(new StringField("--cpu",c));
        return this;
    }

    public MPIJobBuilder memory(String m) {
        this.options.add(new StringField("--memory",m));
        return this;
    }


    /**
     * following functions invoke JobBuilder functions
     *
     *
     * **/


    public MPIJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public MPIJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public MPIJobBuilder workers(int count) {
       super.workers(count);
        return this;
    }

    public MPIJobBuilder imagePullSecrets(ArrayList<String> secrets) {
       super.imagePullSecrets(secrets);
        return this;
    }

    public MPIJobBuilder gpus(int count) {
      super.gpus(count);
        return this;
    }

    public MPIJobBuilder envs(Map<String, String> envs) {
       super.envs(envs);
        return this;
    }

    public MPIJobBuilder nodeSelectors(Map<String, String> selectors) {
       super.nodeSelectors(selectors);
        return this;
    }

    public MPIJobBuilder tolerations(ArrayList<String> tolerations) {
       super.tolerations(tolerations);
        return this;
    }

    public MPIJobBuilder configFiles(Map<String, String> files) {
       super.configFiles(files);
        return this;
    }

    public MPIJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public MPIJobBuilder datas(Map<String, String> datas) {
       super.datas(datas);
        return this;
    }

    public MPIJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public MPIJobBuilder logDir(String dir) {
        super.logDir(dir);
        return this;
    }

    public MPIJobBuilder priority(String priority) {
        super.priority(priority);
        return  this;
    }

    public MPIJobBuilder enableRDMA() {
        super.enableRDMA();
        return  this;
    }

    public MPIJobBuilder syncImage(String image) {
        super.syncImage(image);
        return  this;
    }

    public MPIJobBuilder syncMode(String mode) {
       super.syncMode(mode);
        return  this;
    }

    public MPIJobBuilder syncSource(String source) {
       super.syncSource(source);
        return  this;
    }

    public MPIJobBuilder enableTensorboard() {
        super.enableTensorboard();
        return this;
    }

    public MPIJobBuilder tensorboardImage(String image) {
       super.tensorboardImage(image);
        return this;
    }

    public MPIJobBuilder workingDir(String dir) {
        super.workingDir(dir);
        return this;
    }

    public MPIJobBuilder retryCount(int count) {
        super.retryCount(count);
        return this;
    }

    public MPIJobBuilder enableCoscheduling() {
        super.enableCoscheduling();
        return this;
    }

    public MPIJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}

