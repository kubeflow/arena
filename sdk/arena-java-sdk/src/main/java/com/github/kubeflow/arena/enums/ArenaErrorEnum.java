package com.github.kubeflow.arena.enums;

public enum ArenaErrorEnum {

    VALIDATE_ARGS("validate_args_exception", "failed to validate arena cmd args"),

    UNKNOWN("unknown_exception", "unknown error when used arena"),

    TRAINING_SUBMIT("training_submit_exception", "failed to submit training job"),

    TRAINING_LIST("training_list_exception", "failed to list training jobs"),

    TRAINING_GET("training_get_exception", "failed to get training job"),

    TRAINING_LOGS("training_logs_exception", "failed to get training logs"),

    TRAINING_DELETE("training_delete_exception", "failed to delete training jobs"),

    TRAINING_PRUNE("prune_training_jobs", "failed to prune training jobs"),

    TRAINING_SCALE_IN("scale_in_training_jobs", "failed to scale in training jobs"),

    TRAINING_SCALE_OUT("scale_out_training_jobs", "failed to scale out training jobs"),

    SERVING_SUBMIT("serving_submit_exception", "failed to submit serving job"),

    SERVING_LIST("serving_list_exception", "failed to list serving jobs"),

    SERVING_GET("serving_get_exception", "failed to get serving job"),

    SERVING_LOGS("serving_logs_exception", "failed to get serving job logs"),

    SERVING_DELETE("serving_delete_exception", "failed to delete serving jobs"),

    SERVING_UPDATE("serving_update_exception", "failed to update serving jobs"),

    TOP_NODE("top_node_exception", "failed to get node information"),

    TOP_JOB("top_job_exception", "failed to top job"),

    SERVING_JOB_EXISTS("serving_job_exists", "serving job is existed"),

    SERVING_JOB_NOT_FOUND("serving_job_not_found", "serving job is not found"),

    TRAINING_JOB_EXISTS("training_job_exists", "training job is existed"),

    TRAINING_JOB_NOT_FOUND("training_job_not_found", "not found training job"),

    INVALID_TRAINING_JOB_TYPE("training_job_type_is_invalid", "training job is invalid"),

    EVALUATE_JOB_EXISTS("evaluate_job_exists", "evaluate job is existed"),

    EVALUATE_SUBMIT_FAILED("evaluate_submit_exception", "failed to submit evaluate job"),

    EVALUATE_GET_FAILED("evaluate_get_exception", "failed to get evaluate job"),

    EVALUATE_LIST_FAILED("evaluate_list_exception", "failed to list evaluate jobs"),

    EVALUATE_DELETE_FAILED("evaluate_delete_exception", "failed to delete evaluate job");


    public final String code;

    public final String message;

    ArenaErrorEnum(final String code, final String description) {
        this.code = code;
        this.message = description;
    }

    public String getCode() {
        return code;
    }

    public String getDescription() {
        return message;
    }
}

