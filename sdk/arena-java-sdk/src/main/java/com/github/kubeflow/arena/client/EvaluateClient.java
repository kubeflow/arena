package com.github.kubeflow.arena.client;

import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONObject;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.exceptions.ExitCodeException;
import com.github.kubeflow.arena.model.evaluate.EvaluateJob;
import com.github.kubeflow.arena.model.evaluate.EvaluateJobInfo;
import com.github.kubeflow.arena.utils.Command;

import java.io.IOException;
import java.util.List;

public class EvaluateClient extends BaseClient {

    public EvaluateClient(String kubeConfig, String namespace, String loglevel, String arenaSystemNamespace) {
        super(kubeConfig, namespace, loglevel, arenaSystemNamespace);
    }

    public EvaluateClient namespace(String namespace) {
        return new EvaluateClient(this.kubeConfig, namespace, this.loglevel, this.arenaSystemNamespace);
    }

    public String submit(EvaluateJob job) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("evaluate", "model");

        for (int i = 0; i < job.getArgs().size(); i++) {
            cmds.add(job.getArgs().get(i));
        }
        cmds.add(job.getCommand());
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            return Command.execCommand(arenaCommand);
        } catch (ExitCodeException e) {
            if (e.getMessage().contains(String.format("the job %s is already exist, please delete it first.", job.name()))) {
                throw new ArenaException(ArenaErrorEnum.EVALUATE_JOB_EXISTS, e.getMessage());
            } else {
                throw new ArenaException(ArenaErrorEnum.EVALUATE_SUBMIT_FAILED, e.getMessage());
            }
        }
    }

    public List<EvaluateJobInfo> list() throws ArenaException, IOException {
        return list(false);
    }

    public List<EvaluateJobInfo> list(Boolean allNamespaces) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("evaluate", "list");
        if (allNamespaces != null && allNamespaces) {
            cmds.add("-A");
        }

        cmds.add("-o");
        cmds.add("json");
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return JSONObject.parseArray(output, EvaluateJobInfo.class);
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.EVALUATE_LIST_FAILED, e.getMessage());
        }
    }

    public EvaluateJobInfo get(String jobName) throws ArenaException, IOException {
        List<String> cmds = this.generateCommands("evaluate", "get");
        cmds.add(jobName);
        cmds.add("-o");
        cmds.add("json");
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return JSON.parseObject(output, EvaluateJobInfo.class);
        } catch (ExitCodeException e) {
            if (e.getMessage().contains(String.format("Not found evaluate job %s, please check it with `arena serve list | grep %s`", jobName, jobName))) {
                return null;
            } else {
                throw new ArenaException(ArenaErrorEnum.EVALUATE_GET_FAILED, e.getMessage());
            }
        }
    }

    public String delete(String jobName) throws IOException, ArenaException {
        List<String> cmds = this.generateCommands("evaluate", "delete");
        cmds.add(jobName);
        String[] arenaCommand = cmds.toArray(new String[cmds.size()]);
        try {
            String output = Command.execCommand(arenaCommand);
            return output;
        } catch (ExitCodeException e) {
            throw new ArenaException(ArenaErrorEnum.EVALUATE_DELETE_FAILED, e.getMessage());
        }
    }

}
