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

        ServingJob job = new TensorflowServingJobBuilder()
                .name("bert3")
                .namespace("default-group")
                .gpus(1)
                .replicas(1)
                .modelName("transformer")
                .labels(labels)
                .image("tensorflow/serving:1.15.0-gpu")
                .datas(datas)
                .shell("bash")
                .modelPath("/data/models/soul/saved_model/transformer")
                .build();

        ArenaClient client = new ArenaClient();
        client.serving().submit(job);
    }

}
