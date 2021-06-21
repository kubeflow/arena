package com.github.kubeflow.arena;

import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.nodes.GPUExclusiveNode;
import com.github.kubeflow.arena.model.nodes.NodeSet;
import org.junit.Test;

import java.io.IOException;

public class NodeDetailsTest {

    @Test
    public void testNodeDetails() throws IOException,ArenaException  {
        // 1.create arena client
        System.out.println("start to test arena-java-sdk.");
        ArenaClient client = new ArenaClient();
        System.out.println("create ArenaClient succeed.");
        NodeSet nodeSet = client.nodes().all();
        System.out.println(nodeSet);
        GPUExclusiveNode[] exclusiveNodes = client.nodes().gpuExclusiveNodes();
        for(int i = 0;i < exclusiveNodes.length;i++) {
            System.out.println(exclusiveNodes[i]);
        }
    }

}
