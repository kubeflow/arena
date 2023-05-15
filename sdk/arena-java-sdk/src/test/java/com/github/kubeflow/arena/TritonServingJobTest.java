package com.github.kubeflow.arena;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.serving.ServingJob;
import com.github.kubeflow.arena.model.serving.TritonServingJobBuilder;

import org.junit.Test;

import java.io.IOException;

import java.util.HashMap;
import java.util.Map;

public class TritonServingJobTest {
    @Test
    public void testTensorflowServing() throws ArenaException, IOException {
        Map<String, String> labels = new HashMap<>();
        labels.put("key1", "value1");
        Map<String, String> datas = new HashMap<>();
        datas.put("model-pvc", "/data");
        Map<String, String> tempDirs = new HashMap<>();
        tempDirs.put("empty-0", "/opt/logs");
        Map<String, String> exprs = new HashMap<>();
        exprs.put("empty-0", "$(ARENA_POD_NAMESPACE)/$(ARENA_POD_NAME)");
        ServingJob job = new TritonServingJobBuilder()
                .name("triton-serv")
                .namespace("default-group")
                .version("alpha")
                .gpus(1)
                .replicas(1)
                .labels(labels)
                .image("tensorflow/serving:1.15.0-gpu")
                .datas(datas)
                .tempDirs(tempDirs)
                .tempDirSubpathExprs(exprs)
                .shell("bash")
                .build();

        ArenaClient client = new ArenaClient();
        String output = client.serving().namespace("default-group").submit(job);
        System.out.println(output);

        System.out.println("start to delete triton serving job:");
        String result = client.serving().namespace("default-group").delete("triton-serv", ServingJobType.TritonServingJob,"alpha");
        System.out.println(result);
    }
}
