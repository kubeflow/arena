package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.model.fields.BoolField;
import com.github.kubeflow.arena.model.fields.StringField;
import com.github.kubeflow.arena.model.fields.StringListField;
import com.github.kubeflow.arena.model.fields.StringMapField;

import java.util.ArrayList;
import java.util.Map;

public class CustomServingJobBuilder extends JobBuilder {

    public CustomServingJobBuilder() {
        super(ServingJobType.CustomServingJob);
    }


    /**
     * following functions invoke base class' functions
     *
     *
     * **/
    public CustomServingJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public CustomServingJobBuilder version(String version) {
        super.version(version);
        return this;
    }

    public CustomServingJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public CustomServingJobBuilder replicas(int count) {
        super.replicas(count);
        return this;
    }

    public CustomServingJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public CustomServingJobBuilder gpus(int count) {
       super.gpus(count);
        return this;
    }

    public CustomServingJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public CustomServingJobBuilder nodeSelectors(Map<String, String> selectors) {
       super.nodeSelectors(selectors);
        return this;
    }

    public CustomServingJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public CustomServingJobBuilder annotations(Map<String, String> annotions) {
        super.annotations(annotions);
        return this;
    }

    public CustomServingJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public CustomServingJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public CustomServingJobBuilder gpuMemory(int count) {
        super.gpuMemory(count);
        return this;
    }


    public CustomServingJobBuilder cpu(String c) {
        super.cpu(c);
        return  this;
    }

    public CustomServingJobBuilder memory(String m) {
        super.memory(m);
        return  this;
    }

    public CustomServingJobBuilder enableIstio() {
        super.enableIstio();
        return this;
    }

    public CustomServingJobBuilder port(int port) {
        super.port(port);
        return this;
    }

    public CustomServingJobBuilder restfulPort(int port) {
        super.restfulPort(port);
        return this;
    }

    public CustomServingJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}

