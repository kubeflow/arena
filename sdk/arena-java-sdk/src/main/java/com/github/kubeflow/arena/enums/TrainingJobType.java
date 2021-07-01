package com.github.kubeflow.arena.enums;

public enum TrainingJobType {

    TFTrainingJob("tfjob"),

    MPITrainingJob("mpijob"),

    PytorchTrainingJob("pytorchjob"),

    HorovodTrainingJob("horovodjob"),

    VolcanoTrainingJob("volcanojob"),

    ETTrainingJob("etjob"),

    SparkTrainingJob("sparkjob"),

    AllTrainingJob(""),

    UnknownTrainingJob("unknown");

    private final String otherName;


    TrainingJobType(final String otherName) {
        this.otherName = otherName;
    }

    public String alias() {
        return this.otherName;
    }

    public static TrainingJobType getByAlias(String alias) {
        if (alias == null || alias.length() == 0) {
            return AllTrainingJob;
        }
        for (TrainingJobType value : TrainingJobType.values()) {
            if (alias.equals(value.alias())) {
                return value;
            }
        }
        return UnknownTrainingJob;
    }
}
