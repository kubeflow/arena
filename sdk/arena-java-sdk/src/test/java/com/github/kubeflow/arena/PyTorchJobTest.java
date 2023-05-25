package com.github.kubeflow.arena;

import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.training.PytorchJobBuilder;
import com.github.kubeflow.arena.model.training.TrainingJob;
import org.junit.Test;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

public class PyTorchJobTest {

    @Test
    public void testSubmit() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();

        Map<String, String> labels = new HashMap<>();
        labels.put("arena.kubeflow.org/console-user", "1234567890");

        String command = "python /root/code/mnist-pytorch/mnist.py --backend gloo";

        TrainingJob job = new PytorchJobBuilder()
                .ttlAfterFinished("30s")
                .name("test-pytorchjob-1")
                .gpus(1)
                .labels(labels)
                .syncMode("git")
                .syncSource("https://code.aliyun.com/370272561/mnist-pytorch.git")
                .logDir("/training_logs")
                .image("registry.cn-beijing.aliyuncs.com/ai-samples/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime")
                .imagePullPolicy("IfNotPresent")
                .command(command)
                .build();

        String result = client.training().namespace("default-group").submit(job);
        System.out.println(result);
    }

}
