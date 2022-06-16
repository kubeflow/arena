package com.github.kubeflow.arena;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.training.TFJobBuilder;
import com.github.kubeflow.arena.model.training.TrainingJob;
import com.github.kubeflow.arena.model.training.TrainingJobInfo;
import org.junit.Test;

import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class TrainingClientTest {

    @Test
    public void testSubmit() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();

        Map<String, String> labels = new HashMap<>();
        labels.put("arena.kubeflow.org/console-user", "1138391584567121");

        String command = "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 5000";

        TrainingJob job = new TFJobBuilder()
                .name("test-tfjob")
                .gpus(1)
                .labels(labels)
                .syncMode("git")
                .syncSource("https://github.com/happy2048/tensorflow-sample-code.git")
                .logDir("/training_logs")
                .image("registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu")
                .command(command)
                .build();

        client.training().namespace("default-group").submit(job);
    }

    @Test
    public void testListTrainingJob() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        List<TrainingJobInfo> jobs = client.training().list(TrainingJobType.AllTrainingJob);
        for(TrainingJobInfo info : jobs) {
            System.out.println(JSON.toJSONString(info));

            for(com.github.kubeflow.arena.model.training.Instance instance : info.getInstances()) {
                System.out.println(JSON.toJSONString(instance));
            }
        }
    }

}
