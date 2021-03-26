package com.github.kubeflow.arena.enums;

public enum ServingJobType {

    TFServingJob("tf","Tensorflow"),

    TRTServingJob("trt","Tensorrt"),

    KFServingJob("kf","KFServing"),

    CustomServingJob("custom","Custom"),

    AllServingJob("",""),

    UnknownServingJob("unknown","unknown"),

    ;

    private final String shortHand;
    private final String alias;


    ServingJobType(final String shortHand,final String alias) {
        this.shortHand = shortHand;
        this.alias = alias;
    }

    public String alias() {
        return this.alias.toUpperCase();
    }

    public String shortHand() {
        return this.shortHand;
    }

    public static ServingJobType getByAlias(String alias) {
        if (alias == null || alias.length() == 0) {
            return AllServingJob;
        }
        for (ServingJobType value : ServingJobType.values()) {
            if (alias.toUpperCase().equals(value.alias().toUpperCase())) {
                return value;
            }
        }
        return UnknownServingJob;
    }
}
