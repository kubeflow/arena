package com.github.kubeflow.arena.model.evaluate;

import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.fields.*;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class EvaluateJobBuilder {

    protected String jobName;

    protected List<Field> options;

    protected String command;

    public EvaluateJobBuilder() {
        this.options = new ArrayList<>();
    }

    public EvaluateJob build() throws ArenaException {
        List<String> args = new ArrayList<>();
        for (int i = 0; i < this.options.size(); i++) {
            Field f = this.options.get(i);
            f.validate();
            for (int j = 0; j < f.options().size(); j++) {
                args.add(f.options().get(j));
            }
        }
        return new EvaluateJob(this.jobName, args, this.command);
    }

    public EvaluateJobBuilder name(String name) {
        this.jobName = name;
        this.options.add(new StringField("--name", name));
        return this;
    }

    public EvaluateJobBuilder namespace(String namespace) {
        this.options.add(new StringField("--namespace", namespace));
        return this;
    }

    public EvaluateJobBuilder image(String image) {
        this.options.add(new StringField("--image", image));
        return this;
    }

    public EvaluateJobBuilder imagePullSecrets(ArrayList<String> secrets) {
        this.options.add(new StringListField("--image-pull-secret", secrets));
        return this;
    }

    public EvaluateJobBuilder cpu(String c) {
        this.options.add(new StringField("--cpu", c));
        return this;
    }

    public EvaluateJobBuilder memory(String m) {
        this.options.add(new StringField("--memory", m));
        return this;
    }

    public EvaluateJobBuilder gpus(int count) {
        this.options.add(new StringField("--gpus", String.valueOf(count)));
        return this;
    }

    public EvaluateJobBuilder envs(Map<String, String> envs) {
        this.options.add(new StringMapField("--env", envs, "="));
        return this;
    }

    public EvaluateJobBuilder nodeSelectors(Map<String, String> selectors) {
        this.options.add(new StringMapField("--selector", selectors, "="));
        return this;
    }

    public EvaluateJobBuilder tolerations(ArrayList<String> tolerations) {
        this.options.add(new StringListField("--toleration", tolerations));
        return this;
    }

    public EvaluateJobBuilder annotations(Map<String, String> annotions) {
        this.options.add(new StringMapField("--annotation", annotions, "="));
        return this;
    }

    public EvaluateJobBuilder datas(Map<String, String> datas) {
        this.options.add(new StringMapField("--data", datas, ":"));
        return this;
    }

    public EvaluateJobBuilder dataDirs(Map<String, String> dataDirs) {
        this.options.add(new StringMapField("--data-dir", dataDirs, ":"));
        return this;
    }

    public EvaluateJobBuilder syncImage(String image) {
        this.options.add(new StringField("--sync-image", image));
        return this;
    }

    public EvaluateJobBuilder syncMode(String mode) {
        this.options.add(new StringField("--sync-mode", mode));
        return this;
    }

    public EvaluateJobBuilder syncSource(String source) {
        this.options.add(new StringField("--sync-source", source));
        return this;
    }

    public EvaluateJobBuilder datasetPath(String datasetPath) {
        this.options.add(new StringField("--dataset-path", datasetPath));
        return this;
    }

    public EvaluateJobBuilder metricsPath(String metricsPath) {
        this.options.add(new StringField("--metrics-path", metricsPath));
        return this;
    }

    public EvaluateJobBuilder modelName(String modelName) {
        this.options.add(new StringField("--model-name", modelName));
        return this;
    }

    public EvaluateJobBuilder modelPath(String modelPath) {
        this.options.add(new StringField("--model-path", modelPath));
        return this;
    }

    public EvaluateJobBuilder modelVersion(String modelVersion) {
        this.options.add(new StringField("--model-version", modelVersion));
        return this;
    }

    public EvaluateJobBuilder workingDir(String workingDir) {
        this.options.add(new StringField("--working-dir", workingDir));
        return this;
    }

    public EvaluateJobBuilder shell(String shell) {
        this.options.add(new StringField("--shell", String.valueOf(shell)));
        return this;
    }

    public EvaluateJobBuilder command(String command) {
        this.command = command;
        return this;
    }
}


