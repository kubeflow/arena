package com.github.kubeflow.arena;

import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.serving.ServingJob;
import com.github.kubeflow.arena.model.serving.TensorflowServingJobBuilder;
import org.junit.Test;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

public class TensorflowServingJobTest {

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
        ServingJob job = new TensorflowServingJobBuilder()
                .name("bert3")
                .namespace("default-group")
                .gpus(1)
                .replicas(1)
                .modelName("transformer")
                .labels(labels)
                .image("tensorflow/serving:1.15.0-gpu")
                .datas(datas)
                .tempDirs(tempDirs)
                .tempDirSubpathExprs(exprs)
                .shell("bash")
                .modelPath("/data/models/soul/saved_model/transformer")
                .build();

        ArenaClient client = new ArenaClient();
        client.serving().submit(job);
    }

    @Test
    public void testUpdate() throws ArenaException, IOException {
        ServingJob job = new TensorflowServingJobBuilder()
                .name("cartoongan")
                .namespace("default-group")
                .version("202202091103")
                .replicas(2)
                .build();

        ArenaClient client = new ArenaClient();
        client.serving().namespace("default-group").update(job);
    }

}
