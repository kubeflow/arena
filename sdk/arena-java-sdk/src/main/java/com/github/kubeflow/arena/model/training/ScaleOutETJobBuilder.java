package com.github.kubeflow.arena.model.training;

import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.fields.Field;
import com.github.kubeflow.arena.model.fields.StringField;
import com.github.kubeflow.arena.model.fields.StringMapField;

import java.util.ArrayList;
import java.util.Map;

public class ScaleOutETJobBuilder extends ScaleETJobBuilder {

    public ScaleOutETJobBuilder() {
        super(TrainingJobType.ETTrainingJob);
    }

    /**
     * following functions invoke ScaleETJobBuilder  functions
     *
     *
     * **/

    public ScaleOutETJobBuilder name(String name) {
        super.name(name);
        return this;
    }

    public ScaleOutETJobBuilder timeout(String timeout) {
       super.timeout(timeout);
        return this;
    }

    public ScaleOutETJobBuilder retry(int retry) {
        super.retry(retry);
        return this;
    }

    public ScaleOutETJobBuilder count(int count) {
        super.count(count);
        return this;
    }

    public ScaleOutETJobBuilder script(String script) {
        super.script(script);
        return this;
    }

    public ScaleOutETJobBuilder envs(Map<String, String> envs) {
        super.envs(envs);
        return this;
    }
}
