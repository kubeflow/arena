package com.github.kubeflow.arena.model.serving;

import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;

import java.util.ArrayList;
import java.util.Map;

import com.github.kubeflow.arena.model.fields.*;

public abstract class JobBuilder {

    protected String namespace;

    protected String jobName;

    protected ServingJobType jobType;

    protected ArrayList<Field> options;

    protected String command;

    protected String version = "";

    public JobBuilder(ServingJobType jobType) {
        this.jobType = jobType;
        this.options = new ArrayList<>();
    }

    public ServingJob build() throws ArenaException {
        ArrayList<String> args = new ArrayList<>();
        for (int i = 0; i < this.options.size(); i++) {
            Field f = this.options.get(i);
            f.validate();
            for (int j = 0; j < f.options().size(); j++) {
                args.add(f.options().get(j));
            }
        }
        return new ServingJob(this.jobName, this.jobType, this.version, args, this.command);
    }

    public JobBuilder name(String name) {
        this.jobName = name;
        this.options.add(new StringField("--name", name));
        return this;
    }

    public JobBuilder namespace(String namespace) {
        this.namespace = namespace;
        this.options.add(new StringField("--namespace", namespace));
        return this;
    }

    public JobBuilder version(String version) {
        this.version = version;
        this.options.add(new StringField("--version", version));
        return this;
    }

    public JobBuilder image(String image) {
        this.options.add(new StringField("--image", image));
        return this;
    }

    public JobBuilder replicas(int count) {
        this.options.add(new StringField("--replicas", String.valueOf(count)));
        return this;
    }

    public JobBuilder imagePullSecrets(ArrayList<String> secrets) {
        this.options.add(new StringListField("--image-pull-secret", secrets));
        return this;
    }

    public JobBuilder gpus(int count) {
        this.options.add(new StringField("--gpus", String.valueOf(count)));
        return this;
    }

    public JobBuilder envs(Map<String, String> envs) {
        this.options.add(new StringMapField("--env", envs, "="));
        return this;
    }

    public JobBuilder nodeSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--selector", selectors, "="));
        return this;
    }

    public JobBuilder tolerations(ArrayList<String> tolerations) {
        this.options.add(new StringListField("--toleration", tolerations));
        return this;
    }

    public JobBuilder annotations(Map<String, String> annotations) {
        this.options.add(new StringMapField("--annotation", annotations, "="));
        return this;
    }

    public JobBuilder labels(Map<String, String> labels) {
        this.options.add(new StringMapField("--label", labels, "="));
        return this;
    }

    public JobBuilder devices(Map<String, String> devices) {
        this.options.add(new StringMapField("--device", devices, "="));
        return this;
    }

    public JobBuilder datas(Map<String, String> datas) {
        this.options.add(new StringMapField("--data", datas, ":"));
        return this;
    }

    public JobBuilder dataSubpathExprs(Map<String, String> exprs) {
        this.options.add(new StringMapField("--data-subpath-expr", exprs, ":"));
        return this;
    }

    public JobBuilder tempDirs(Map<String, String> tempDirs) {
        this.options.add(new StringMapField("--temp-dir", tempDirs, ":"));
        return this;
    }

    public JobBuilder tempDirSubpathExprs(Map<String, String> exprs) {
        this.options.add(new StringMapField("--temp-dir-subpath-expr", exprs, ":"));
        return this;
    }

    public JobBuilder dataDirs(Map<String, String> dataDirs) {
        this.options.add(new StringMapField("--data-dir", dataDirs, ":"));
        return this;
    }

    public JobBuilder gpuMemory(int count) {
        this.options.add(new StringField("--gpumemory", String.valueOf(count)));
        return this;
    }

    public JobBuilder cpu(String c) {
        this.options.add(new StringField("--cpu", c));
        return this;
    }

    public JobBuilder memory(String m) {
        this.options.add(new StringField("--memory", m));
        return this;
    }

    public JobBuilder enableIstio() {
        this.options.add(new BoolField("--enable-istio"));
        return this;
    }

    public JobBuilder shell(String shell) {
        this.options.add(new StringField("--shell", String.valueOf(shell)));
        return this;
    }

    public JobBuilder command(String command) {
        this.command = command;
        return this;
    }
}


