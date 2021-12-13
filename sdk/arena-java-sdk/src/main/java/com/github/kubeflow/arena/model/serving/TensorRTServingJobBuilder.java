package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.model.fields.BoolField;
import com.github.kubeflow.arena.model.fields.StringField;

import java.util.ArrayList;
import java.util.Map;

/**
 * Deprecated. Please use TritonJobBuilder instead
 */
@Deprecated
public class TensorRTServingJobBuilder extends JobBuilder {

    public TensorRTServingJobBuilder() {
        super(ServingJobType.TRTServingJob);
    }

    public TensorRTServingJobBuilder modelStore(String store) {
        this.options.add(new StringField("--model-store",String.valueOf(store)));
        return this;
    }

    public TensorRTServingJobBuilder port(int port) {
        this.options.add(new StringField("--grpc-port",String.valueOf(port)));
        return this;
    }

    public TensorRTServingJobBuilder metricPort(int port) {
        this.options.add(new StringField("--metric-port",String.valueOf(port)));
        return this;
    }

    public TensorRTServingJobBuilder allowMetric() {
        this.options.add(new BoolField("--allow-metrics"));
        return this;
    }

    public TensorRTServingJobBuilder restfulPort(int port) {
        this.options.add(new StringField("--http-port",String.valueOf(port)));
        return this;
    }

    /**
     * following functions invoke base class' functions
     *
     *
     * **/
    public TensorRTServingJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public TensorRTServingJobBuilder version(String version) {
        super.version(version);
        return this;
    }

    public TensorRTServingJobBuilder image(String image) {
        super.image(image);
        return this;
    }

    public TensorRTServingJobBuilder replicas(int count) {
        super.replicas(count);
        return this;
    }

    public TensorRTServingJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        super.imagePullSecrets(secrets);
        return this;
    }

    public TensorRTServingJobBuilder gpus(int count) {
        super.gpus(count);
        return this;
    }

    public TensorRTServingJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }

    public TensorRTServingJobBuilder nodeSelectors(Map<String, String> selectors) {
        super.nodeSelectors(selectors);
        return this;
    }

    public TensorRTServingJobBuilder tolerations(ArrayList<String> tolerations) {
        super.tolerations(tolerations);
        return this;
    }

    public TensorRTServingJobBuilder annotations(Map<String, String> annotations) {
        super.annotations(annotations);
        return this;
    }

    public TensorRTServingJobBuilder datas(Map<String, String> datas) {
        super.datas(datas);
        return this;
    }

    public TensorRTServingJobBuilder dataDirs(Map<String, String> dataDirs) {
        super.dataDirs(dataDirs);
        return  this;
    }

    public TensorRTServingJobBuilder gpuMemory(int count) {
        super.gpuMemory(count);
        return this;
    }


    public TensorRTServingJobBuilder cpu(String c) {
        super.cpu(c);
        return  this;
    }

    public TensorRTServingJobBuilder memory(String m) {
        super.memory(m);
        return  this;
    }

    public TensorRTServingJobBuilder enableIstio() {
        super.enableIstio();
        return this;
    }

    public TensorRTServingJobBuilder shell(String shell) {
        super.shell(shell);
        return this;
    }

    public TensorRTServingJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}
