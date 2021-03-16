package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;

import java.util.Map;

public class ScaleInETJobBuilder extends ScaleETJobBuilder {

    public ScaleInETJobBuilder() {
        super(TrainingJobType.ETTrainingJob);
    }

    /**
     * following functions invoke ScaleETJobBuilder  functions
     *
     *
     * **/

    public ScaleInETJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public ScaleInETJobBuilder timeout(String timeout) {
        super.timeout(timeout);
        return this;
    }

    public ScaleInETJobBuilder retry(int retry) {
        super.retry(retry);
        return this;
    }

    public ScaleInETJobBuilder count(int count) {
        super.count(count);
        return this;
    }

    public ScaleInETJobBuilder script(String script) {
        super.script(script);
        return this;
    }

    public ScaleInETJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }
}
