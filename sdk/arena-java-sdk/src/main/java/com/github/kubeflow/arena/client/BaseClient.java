package com.github.kubeflow.arena.client;

import java.util.ArrayList;
import java.util.List;

public class BaseClient {

    protected String namespace = "";
    protected String kubeConfig = "";
    protected String loglevel = "";
    protected String arenaSystemNamespace = "";
    protected static final String arenaBinary = "arena";

    protected BaseClient() {

    }

    protected BaseClient(String kubeConfig, String namespace, String loglevel, String arenaSystemNamespace) {
        this.namespace = namespace;
        this.kubeConfig = kubeConfig;
        this.loglevel = loglevel;
        this.arenaSystemNamespace = arenaSystemNamespace;
    }

    protected BaseClient namespace(String namespace) {
        return new BaseClient(this.kubeConfig, namespace, this.loglevel, this.arenaSystemNamespace);
    }

    protected List<String> generateCommands(String... subCommand) {
        List<String> cmds = new ArrayList<>();
        cmds.add(arenaBinary);
        for (int i = 0; i < subCommand.length; i++) {
            cmds.add(subCommand[i]);
        }
        if (this.namespace.length() != 0) {
            cmds.add("--namespace=" + this.namespace);
        }
        if (this.kubeConfig.length() != 0) {
            cmds.add("--config=" + this.kubeConfig);
        }
        if (this.loglevel.length() != 0) {
            cmds.add("--loglevel=" + this.loglevel);
        }
        if (this.arenaSystemNamespace.length() != 0) {
            cmds.add("--arena-namespace=" + this.arenaSystemNamespace);
        }
        return cmds;
    }
}
