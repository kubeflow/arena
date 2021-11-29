package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class KFServingJobBuilder  extends JobBuilder {

    public  KFServingJobBuilder() {
        super(ServingJobType.KFServingJob);
    }

    public KFServingJobBuilder modelType(String modelType) {
        this.options.add(new StringField("--model-type",modelType));
        return  this;
    }

    public KFServingJobBuilder storageURI(String uri) {
        this.options.add(new StringField("--storage-uri",uri));
        return  this;
    }

    public KFServingJobBuilder canaryPercent(int percent) {
        this.options.add(new StringField("--canary-percent",String.valueOf(percent)));
        return  this;
    }

    public KFServingJobBuilder port(int port) {
        this.options.add(new StringField("--port", String.valueOf(port)));
        return this;
    }

    /**
     * following functions invoke base class' functions
     *
     *
     * **/
    public KFServingJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public KFServingJobBuilder namespace(String namespace) {
        super.namespace(namespace);
        return this;
    }

    public KFServingJobBuilder version(String version) {
        super.version(version);
        return this;
    }

    public KFServingJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public KFServingJobBuilder replicas(int count) {
        super.replicas(count);
        return this;
    }

    public KFServingJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public KFServingJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public KFServingJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public KFServingJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public KFServingJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public KFServingJobBuilder annotations(Map<String, String> annotations) {
        super.annotations(annotations);
        return this;
    }

    public KFServingJobBuilder labels(Map<String, String> labels) {
        super.labels(labels);
        return this;
    }

    public KFServingJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public KFServingJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public KFServingJobBuilder gpuMemory(int count) {
        super.gpuMemory(count);
        return this;
    }

    public KFServingJobBuilder cpu(String c) {
        super.cpu(c);
        return this;
    }

    public KFServingJobBuilder memory(String m) {
        super.memory(m);
        return this;
    }

    public KFServingJobBuilder enableIstio() {
        super.enableIstio();
        return this;
    }

    public KFServingJobBuilder shell(String shell) {
        super.shell(shell);
        return this;
    }

    public KFServingJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
