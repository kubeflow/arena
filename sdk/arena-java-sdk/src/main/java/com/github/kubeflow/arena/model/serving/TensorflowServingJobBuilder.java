package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class TensorflowServingJobBuilder extends JobBuilder {

    public TensorflowServingJobBuilder() {
        super(ServingJobType.TFServingJob);
    }

    public JobBuilder modelName(String modelName) {
        this.options.add(new StringField("--model-name",String.valueOf(modelName)));
        return this;
    }

    public JobBuilder modelPath(String modelPath) {
        this.options.add(new StringField("--model-path",String.valueOf(modelPath)));
        return this;
    }

    public JobBuilder modelConfigFile(String modelConfigFile) {
        this.options.add(new StringField("--modelConfigFile",String.valueOf(modelConfigFile)));
        return this;
    }

    public JobBuilder versionPolicy(String policy) {
        this.options.add(new StringField("--version-policy",String.valueOf(policy)));
        return this;
    }

    public JobBuilder port(int port) {
        this.options.add(new StringField("--port",String.valueOf(port)));
        return this;
    }

    public JobBuilder restfulPort(int port) {
        this.options.add(new StringField("--restfulPort",String.valueOf(port)));
        return this;
    }

    /**
     * following functions invoke base class' functions
     *
     *
     * **/
    public TensorflowServingJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public TensorflowServingJobBuilder version(String version) {
        super.version(version);
        return this;
    }

    public TensorflowServingJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public TensorflowServingJobBuilder replicas(int count) {
        super.replicas(count);
        return this;
    }

    public TensorflowServingJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public TensorflowServingJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public TensorflowServingJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public TensorflowServingJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public TensorflowServingJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public TensorflowServingJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public TensorflowServingJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public TensorflowServingJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public TensorflowServingJobBuilder gpuMemory(int count) {
        super.gpuMemory(count);
        return this;
    }


    public TensorflowServingJobBuilder cpu(String c) {
        super.cpu(c);
        return  this;
    }

    public TensorflowServingJobBuilder memory(String m) {
        super.memory(m);
        return  this;
    }

    public TensorflowServingJobBuilder enableIstio() {
        super.enableIstio();
        return this;
    }

    public TensorflowServingJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
