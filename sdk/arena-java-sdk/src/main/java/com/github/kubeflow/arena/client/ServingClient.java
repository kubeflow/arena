package com.github.kubeflow.arena.client;

import com.alibaba.fastjson.JSONObject;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.exceptions.ExitCodeException;
import com.github.kubeflow.arena.model.serving.ServingJob;
import com.github.kubeflow.arena.model.serving.ServingJobInfo;
import com.github.kubeflow.arena.utils.Command;

import java.io.IOException;
import java.util.List;

public class ServingClient extends BaseClient {

    public ServingClient(String kubeConfig, String namespace, String loglevel, String arenaSystemNamespace) {
        super(kubeConfig, namespace, loglevel, arenaSystemNamespace);
    }

    public ServingClient namespace(String namespace) {
        return new ServingClient(this.kubeConfig, namespace, this.loglevel, this.arenaSystemNamespace);
    }

    public String submit(ServingJob job) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("serve", job.getType().shortHand());
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
                throw new ArenaException(ArenaErrorEnum.SERVING_JOB_EXISTS, e.getMessage());
            } else {
                throw new ArenaException(ArenaErrorEnum.SERVING_SUBMIT, e.getMessage());
            }
        }
    }

    public List<ServingJobInfo> list(ServingJobType jobType) throws ArenaException, IOException {
        return this.list(jobType, null);
    }

    public List<ServingJobInfo> list(ServingJobType jobType, Boolean allNamespaces) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("serve", "list");
        if (!jobType.equals(ServingJobType.AllServingJob) && !jobType.equals(ServingJobType.UnknownServingJob)) {
            cmds.add("--type=" + jobType.shortHand());
        }

        if (allNamespaces != null && allNamespaces) {
            cmds.add("-A");
        }

        cmds.add("-o");
        cmds.add("json");
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return JSONObject.parseArray(output, ServingJobInfo.class);
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.SERVING_LIST, e.getMessage());
        }
    }

    public ServingJobInfo get(String jobName, ServingJobType jobType, String jobVersion) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("serve", "get");
        if (!jobType.equals(ServingJobType.AllServingJob) && !jobType.equals(ServingJobType.UnknownServingJob)) {
            cmds.add("--type=" + jobType.shortHand());
        }
        if (jobVersion != null && jobVersion.length() != 0) {
            cmds.add("--version=" + jobVersion);
        }
        cmds.add(jobName);
        cmds.add("-o");
        cmds.add("json");
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        ServingJobInfo jobInfo = new ServingJobInfo();
        try {
            String output = Command.execCommand(arenaCommand);
            jobInfo = jobInfo.complete(output);
            return jobInfo;
        } catch (ExitCodeException e) {
            if (e.getMessage().contains(String.format("Not found serving job %s, please check it with `arena serve list | grep %s`", jobName, jobName))) {
                return null;
            } else {
                throw new ArenaException(ArenaErrorEnum.SERVING_GET, e.getMessage());
            }
        }
    }

    public String delete(String jobName, ServingJobType jobType, String jobVersion) throws IOException, ArenaException {
        List<String> cmds = this.generateCommands("serve", "delete");
        if (!jobType.equals(ServingJobType.AllServingJob) && !jobType.equals(ServingJobType.UnknownServingJob)) {
            cmds.add("--type=" + jobType.shortHand());
        }
        if (jobVersion != null && jobVersion.length() != 0) {
            cmds.add("--version=" + jobVersion);
        }
        cmds.add(jobName);
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.SERVING_DELETE, e.getMessage());
        }
    }

}
