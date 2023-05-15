package com.github.kubeflow.arena;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.serving.Instance;
import com.github.kubeflow.arena.model.serving.ServingJob;
import com.github.kubeflow.arena.model.serving.ServingJobInfo;
import com.github.kubeflow.arena.model.serving.TensorflowServingJobBuilder;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class ServingClientTest {

    @Before
    public void testSubmit() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();

        Map<String, String> datas = new HashMap<>();
        datas.put("model-pvc", "/data");

        ServingJob job = new TensorflowServingJobBuilder()
                .name("test-tfserving")
                .namespace("default-group")
                .modelName("transformer")
                .version("alpha")
                .replicas(1)
                .gpus(1)
                .image("tensorflow/serving:1.15.0-gpu")
                .datas(datas)
                .modelPath("/data/models/soul/new_opt_saved_model/transformer")
                .build();

        client.serving().namespace("default-group").submit(job);
    }

    @Test
    public void testListServingJob() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        List<ServingJobInfo> jobs = client.serving().list(ServingJobType.AllServingJob, true);
        for(ServingJobInfo info : jobs) {
            System.out.println(JSON.toJSONString(info));
            for(Instance instance : info.getInstances()) {
                System.out.println(JSON.toJSONString(instance));
            }
        }
    }

    @Test
    public void testUpdate() throws ArenaException, IOException {
        ServingJob job = new TensorflowServingJobBuilder()
                .name("test-tfserving")
                .namespace("default-group")
                .version("alpha")
                .replicas(2)
                .build();

        ArenaClient client = new ArenaClient();
        String result = client.serving().namespace("default-group").update(job);
        System.out.println(result);
    }

    @After
    public void testDelete() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        String result = client.serving().namespace("default-group")
                .delete("test-tfserving", ServingJobType.TFServingJob, "alpha");
        System.out.println(result);
    }

}
