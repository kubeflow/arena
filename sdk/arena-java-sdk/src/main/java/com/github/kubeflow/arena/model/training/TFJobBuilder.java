package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.model.fields.BoolField;
import com.github.kubeflow.arena.model.fields.StringField;
import com.github.kubeflow.arena.model.fields.StringMapField;

import java.util.ArrayList;
import java.util.Map;

public class TFJobBuilder extends JobBuilder {

    public TFJobBuilder() {
        super(TrainingJobType.TFTrainingJob);
    }

    public TFJobBuilder workerCount(int count) {
        this.options.add(new StringField("--workers",String.valueOf(count)));
        return this;
    }

    public TFJobBuilder workerSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--worker-selector",selectors,"="));
        return this;
    }

    public TFJobBuilder workerPort(int port) {
        this.options.add(new StringField("--worker-port",String.valueOf(port)));
        return this;
    }

    public TFJobBuilder workerMemory(String memory) {
        this.options.add(new StringField("--worker-memory",memory));
        return this;
    }

    public TFJobBuilder workerMemoryLimit(String memory) {
        this.options.add(new StringField("--worker-memory-limit",memory));
        return this;
    }

    public TFJobBuilder workerImage(String image) {
        this.options.add(new StringField("--worker-image",image));
        return this;
    }

    public TFJobBuilder workerCPU(String cpu) {
        this.options.add(new StringField("--worker-cpu",cpu));
        return this;
    }

    public TFJobBuilder workerCPULimit(String cpu) {
        this.options.add(new StringField("--worker-cpu-limit",cpu));
        return this;
    }

    public TFJobBuilder psSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--ps-selector",selectors,"="));
        return this;
    }

    public TFJobBuilder psPort(int port) {
        this.options.add(new StringField("--ps-port",String.valueOf(port)));
        return this;
    }

    public TFJobBuilder psMemory(String memory) {
        this.options.add(new StringField("--ps-memory",memory));
        return this;
    }

    public TFJobBuilder psMemoryLimit(String memory) {
        this.options.add(new StringField("--ps-memory-limit",memory));
        return this;
    }

    public TFJobBuilder psImage(String image) {
        this.options.add(new StringField("--ps-image",image));
        return this;
    }

    public TFJobBuilder psCPU(String cpu) {
        this.options.add(new StringField("--ps-cpu",cpu));
        return this;
    }

    public TFJobBuilder psCPULimit(String cpu) {
        this.options.add(new StringField("--ps-cpu-limit",cpu));
        return this;
    }

    public TFJobBuilder psCount(int count) {
        this.options.add(new StringField("--ps",String.valueOf(count)));
        return this;
    }

    public TFJobBuilder evaluatorSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--evaluator-selector",selectors,"="));
        return this;
    }

    public TFJobBuilder evaluatorMemory(String memory) {
        this.options.add(new StringField("--evaluator-memory",memory));
        return  this;
    }

    public TFJobBuilder evaluatorMemoryLimit(String memory) {
        this.options.add(new StringField("--evaluator-memory-limit",memory));
        return  this;
    }

    public TFJobBuilder evaluatorCPU(String cpu) {
        this.options.add(new StringField("--evaluator-cpu",cpu));
        return this;
    }

    public TFJobBuilder evaluatorCPULimit(String cpu) {
        this.options.add(new StringField("--evaluator-cpu-limit",cpu));
        return this;
    }

    public TFJobBuilder enableEvaluator() {
        this.options.add(new BoolField("--evaluator"));
        return this;
    }

    public TFJobBuilder chiefSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--chief-selector",selectors,"="));
        return this;
    }

    public TFJobBuilder chiefPort(int port) {
        this.options.add(new StringField("--chief-port",String.valueOf(port)));
        return this;
    }

    public TFJobBuilder chiefMemory(String memory) {
        this.options.add(new StringField("--chief-memory",memory));
        return this;
    }

    public TFJobBuilder chiefMemoryLimit(String memory) {
        this.options.add(new StringField("--chief-memory-limit",memory));
        return this;
    }

    public TFJobBuilder chiefCPU(String cpu) {
        this.options.add(new StringField("--chief-cpu",cpu));
        return this;
    }

    public TFJobBuilder chiefCPULimit(String cpu) {
        this.options.add(new StringField("--chief-cpu-limit",cpu));
        return this;
    }

    public TFJobBuilder enableChief() {
        this.options.add(new BoolField("--chief"));
        return this;
    }

    /**
     * following functions invoke JobBuilder functions
     *
     *
     * **/


    public TFJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public TFJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public TFJobBuilder workers(int count) {
        super.workers(count);
        return this;
    }

    public TFJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public TFJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public TFJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public TFJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public TFJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public TFJobBuilder configFiles(Map<String, String> files) {
        super.configFiles(files);
        return this;
    }

    public TFJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public TFJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public TFJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public TFJobBuilder logDir(String dir) {
        super.logDir(dir);
        return this;
    }

    public TFJobBuilder priority(String priority) {
        super.priority(priority);
        return  this;
    }

    public TFJobBuilder enableRDMA() {
        super.enableRDMA();
        return  this;
    }

    public TFJobBuilder syncImage(String image) {
        super.syncImage(image);
        return  this;
    }

    public TFJobBuilder syncMode(String mode) {
        super.syncMode(mode);
        return  this;
    }

    public TFJobBuilder syncSource(String source) {
        super.syncSource(source);
        return  this;
    }

    public TFJobBuilder enableTensorboard() {
        super.enableTensorboard();
        return this;
    }

    public TFJobBuilder tensorboardImage(String image) {
        super.tensorboardImage(image);
        return this;
    }

    public TFJobBuilder workingDir(String dir) {
        super.workingDir(dir);
        return this;
    }

    public TFJobBuilder retryCount(int count) {
        super.retryCount(count);
        return this;
    }

    public TFJobBuilder enableCoscheduling() {
        super.enableCoscheduling();
        return this;
    }

    public TFJobBuilder shell(String shell) {
        super.shell(shell);
        return this;
    }

    public TFJobBuilder command(String command) {
        this.command = command;
        return this;
    }

}
