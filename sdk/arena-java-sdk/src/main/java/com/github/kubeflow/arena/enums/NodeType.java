package com.github.kubeflow.arena.enums;

public enum NodeType {

    GPUShareNodeType("s", "GPUShare"),

    GPUExclusiveNodeType("e", "GPUExclusive"),

    GPUTopologyNodeType("t", "GPUTopology"),

    NormalNodeType("n", "Normal"),

    AllNodeType("", ""),

    UnknownNodeType("unknown", "unknown");

    private final String shortHand;
    private final String alias;

    NodeType(final String shortHand, final String alias) {
        this.shortHand = shortHand;
        this.alias = alias;
    }

    public String alias() {
        return this.alias;
    }

    public String shortHand() {
        return this.shortHand;
    }

    public static NodeType getByAlias(String alias) {
        if (alias == null || alias.length() == 0) {
            return AllNodeType;
        }
        for (NodeType value : NodeType.values()) {
            if (alias.toUpperCase().equals(value.alias().toUpperCase())) {
                return value;
            }
        }
        return UnknownNodeType;
    }
}
