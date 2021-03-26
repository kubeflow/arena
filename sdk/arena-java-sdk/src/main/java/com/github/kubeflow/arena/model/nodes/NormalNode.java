package com.github.kubeflow.arena.model.nodes;

import com.alibaba.fastjson.JSON;

public class NormalNode extends Node {
    public NormalNode(){
        super();
    }

    @Override
    public String toString() {
        return JSON.toJSONString(this,true);
    }
}
