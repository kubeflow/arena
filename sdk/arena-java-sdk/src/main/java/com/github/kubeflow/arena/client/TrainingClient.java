package com.github.kubeflow.arena.client;

import com.alibaba.fastjson.JSONObject;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.exceptions.ExitCodeException;
import com.github.kubeflow.arena.model.training.TrainingJob;
import com.github.kubeflow.arena.model.training.TrainingJobInfo;
import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.utils.Command;

import java.io.IOException;
import java.util.List;

public class TrainingClient extends BaseClient {

    public TrainingClient(String kubeConfig, String namespace, String loglevel, String arenaSystemNamespace) {
        super(kubeConfig, namespace, loglevel, arenaSystemNamespace);
    }

    public TrainingClient namespace(String namespace) {
        return new TrainingClient(this.kubeConfig, namespace, this.loglevel, this.arenaSystemNamespace);
    }

    public String submit(TrainingJob job) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("submit");
        cmds.add(job.getType().alias());
        for (int i = 0; i < job.getArgs().size(); i++) {
            cmds.add(job.getArgs().get(i));
        }
        cmds.add(job.getCommand());
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            if (e.getMessage().contains(String.format("the job %s is already exist, please delete it first.", job.name()))) {
                throw new ArenaException(ArenaErrorEnum.TRAINING_JOB_EXISTS, e.getMessage());
            } else {
                throw new ArenaException(ArenaErrorEnum.TRAINING_SUBMIT, e.getMessage());
            }
        }
    }

    public List<TrainingJobInfo> list(TrainingJobType jobType) throws ArenaException, IOException {
        return list(jobType, false);
    }

    public List<TrainingJobInfo> list(TrainingJobType jobType, Boolean allNamespaces) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("list");
        if (!jobType.equals(TrainingJobType.AllTrainingJob) && !jobType.equals(TrainingJobType.UnknownTrainingJob)) {
            cmds.add("--type=" + jobType.alias());
        }

        if (allNamespaces != null && allNamespaces) {
            cmds.add("-A");
        }

        cmds.add("-o");
        cmds.add("json");
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return JSONObject.parseArray(output, TrainingJobInfo.class);
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.TRAINING_LIST, e.getMessage());
        }
    }

    public TrainingJobInfo get(String jobName, TrainingJobType jobType) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("get");
        if (!jobType.equals(TrainingJobType.AllTrainingJob) && !jobType.equals(TrainingJobType.UnknownTrainingJob)) {
            cmds.add("--type=" + jobType.alias());
        }
        cmds.add(jobName);
        cmds.add("-o");
        cmds.add("json");
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        TrainingJobInfo jobInfo = new TrainingJobInfo();
        try {
            String output = Command.execCommand(arenaCommand);
            jobInfo = jobInfo.complete(output);
            return jobInfo;
        } catch (ExitCodeException e) {
            if (e.getMessage().contains(String.format("Not found training job %s in namespace %s,please use 'arena submit' to create it.", jobName, this.namespace))) {
                return null;
            } else {
                throw new ArenaException(ArenaErrorEnum.TRAINING_GET, e.getMessage());
            }
        }
    }

    public String delete(String jobName, TrainingJobType jobType) throws IOException, ArenaException {
        List<String> cmds = this.generateCommands("delete");
        if (!jobType.equals(TrainingJobType.AllTrainingJob) && !jobType.equals(TrainingJobType.UnknownTrainingJob)) {
            cmds.add("--type=" + jobType.alias());
        }
        cmds.add(jobName);
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.TRAINING_DELETE, e.getMessage());
        }
    }

    public String prune(String duration, Boolean allNamespaces) throws IOException, ArenaException {
        List<String> cmds = this.generateCommands("prune");
        cmds.add("--since=" + duration);
        if (allNamespaces != null && allNamespaces) {
            cmds.add("-A");
        }
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.TRAINING_PRUNE, e.getMessage());
        }
    }

    public String scaleIn(TrainingJob job) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("scalein");
        cmds.add(job.getType().alias());
        for (int i = 0; i < job.getArgs().size(); i++) {
            cmds.add(job.getArgs().get(i));
        }
        cmds.add(job.getCommand());
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.TRAINING_SCALE_IN, e.getMessage());
        }
    }

    public String scaleOut(TrainingJob job) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("scaleout");
        cmds.add(job.getType().alias());
        for (int i = 0; i < job.getArgs().size(); i++) {
            cmds.add(job.getArgs().get(i));
        }
        cmds.add(job.getCommand());
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.TRAINING_SCALE_OUT, e.getMessage());
        }
    }

}

