package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.model.fields.BoolField;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

public class TritonServingJobBuilder extends JobBuilder {

    public TritonServingJobBuilder() {
        super(ServingJobType.TritonServingJob);
    }

    public TritonServingJobBuilder modelRepository(String modelRepository) {
        this.options.add(new StringField("--model-repository",String.valueOf(modelRepository)));
        return this;
    }

    public TritonServingJobBuilder httpPort(int port) {
        this.options.add(new StringField("--http-port",String.valueOf(port)));
        return this;
    }

    public TritonServingJobBuilder grpcPort(int port) {
        this.options.add(new StringField("--grpc-port",String.valueOf(port)));
        return this;
    }

    public TritonServingJobBuilder allowMetrics() {
        this.options.add(new BoolField("--allow-metrics"));
        return this;
    }

    public TritonServingJobBuilder metricsPort(int port) {
        this.options.add(new StringField("--metrics-port",String.valueOf(port)));
        return this;
    }

    /**
     * following functions invoke base class' functions
     *
     *
     * **/
    public TritonServingJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public TritonServingJobBuilder namespace(String namespace) {
        super.namespace(namespace);
        return this;
    }

    public TritonServingJobBuilder version(String version) {
        super.version(version);
        return this;
    }

    public TritonServingJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public TritonServingJobBuilder replicas(int count) {
        super.replicas(count);
        return this;
    }

    public TritonServingJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public TritonServingJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public TritonServingJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public TritonServingJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public TritonServingJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public TritonServingJobBuilder annotations(Map<String, String> annotations) {
        super.annotations(annotations);
        return this;
    }

    public TritonServingJobBuilder labels(Map<String, String> labels) {
        super.labels(labels);
        return this;
    }

    public TritonServingJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public TritonServingJobBuilder dataSubpathExprs(Map<String, String> exprs) {
        super.dataSubpathExprs(exprs);
        return this;
    }
    public TritonServingJobBuilder emptyDirs(Map<String, String> emptyDirs) {
        super.emptyDirs(emptyDirs);
        return this;
    }

    public TritonServingJobBuilder emptyDirSubpathExprs(Map<String, String> exprs) {
        super.emptyDirSubpathExprs(exprs);
        return this;
    }
    public TritonServingJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public TritonServingJobBuilder gpuMemory(int count) {
        super.gpuMemory(count);
        return this;
    }

    public TritonServingJobBuilder cpu(String c) {
        super.cpu(c);
        return this;
    }

    public TritonServingJobBuilder memory(String m) {
        super.memory(m);
        return this;
    }

    public TritonServingJobBuilder enableIstio() {
        super.enableIstio();
        return this;
    }

    public TritonServingJobBuilder shell(String shell) {
        super.shell(shell);
        return this;
    }

    public TritonServingJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
