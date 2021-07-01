package com.github.kubeflow.arena.enums;

public enum TrainingJobStatus {

    TrainingJobPending("PENDING"),

    TrainingJobRunning("RUNNING"),

    TrainingJobSucceeded("SUCCEEDED"),

    TrainingJobFailed("FAILED"),

    TrainingJobUnknownStatus("");

    public final String otherName;

    TrainingJobStatus(final String otherName) {
        this.otherName = otherName;
    }

    public String alias() {
        return this.otherName;
    }

    public static TrainingJobStatus getByAlias(String alias) {
        for (TrainingJobStatus value : TrainingJobStatus.values()) {
            if (alias.equals(value.alias())) {
                return value;
            }
        }
        return TrainingJobUnknownStatus;
    }
}
