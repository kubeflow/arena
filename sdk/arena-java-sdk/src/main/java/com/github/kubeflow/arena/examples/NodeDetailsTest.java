package com.github.kubeflow.arena.examples;

import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.model.nodes.*;
import com.github.kubeflow.arena.exceptions.ArenaException;

import java.io.IOException;

public class NodeDetailsTest {
    public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        // 1.create arena client
        System.out.println("start to test arena-java-sdk.");
        ArenaClient client = new ArenaClient("/Users/yangjunfeng/.kube/config-gpushare");
        System.out.println("create ArenaClient succeed.");
        NodeSet nodeSet = client.nodes().all();
        System.out.println(nodeSet);
        GPUExclusiveNode[] exclusiveNodes = client.nodes().gpuExclusiveNodes();
        for(int i = 0;i < exclusiveNodes.length;i++) {
            System.out.println(exclusiveNodes[i]);
        }
    }
}
