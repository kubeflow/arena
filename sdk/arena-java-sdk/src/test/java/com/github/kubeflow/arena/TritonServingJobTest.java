package com.github.kubeflow.arena;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.common.Logger;
import com.github.kubeflow.arena.model.serving.CustomServingJobBuilder;
import com.github.kubeflow.arena.model.serving.Instance;
import com.github.kubeflow.arena.model.serving.ServingJob;
import com.github.kubeflow.arena.model.serving.ServingJobInfo;
import com.github.kubeflow.arena.model.serving.TritonServingJobBuilder;

import org.junit.Test;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
public class TritonServingJobTest {
    @Test
    public void testTensorflowServing() throws ArenaException, IOException {
        Map<String, String> labels = new HashMap<>();
        labels.put("key1", "value1");
        Map<String, String> datas = new HashMap<>();
        datas.put("model-pvc", "/data");
        Map<String, String> emptyDirs = new HashMap<>();
        emptyDirs.put("empty-0", "/opt/logs");
        Map<String, String> exprs = new HashMap<>();
        exprs.put("empty-0", "$(ARENA_POD_NAMESPACE)/$(ARENA_POD_NAME)");
        ServingJob job = new TritonServingJobBuilder()
                .name("triton-serv")
                .namespace("default-group")
                .gpus(1)
                .replicas(1)
                .labels(labels)
                .image("tensorflow/serving:1.15.0-gpu")
                .datas(datas)
                .emptyDirs(emptyDirs)
                .emptyDirSubpathExprs(exprs)
                .shell("bash")
                .build();

        ArenaClient client = new ArenaClient();
        client.serving().submit(job);
    }
}
