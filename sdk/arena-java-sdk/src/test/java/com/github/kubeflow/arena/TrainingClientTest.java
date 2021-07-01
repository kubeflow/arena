package com.github.kubeflow.arena;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.training.TrainingJobInfo;
import org.junit.Test;

import java.io.IOException;
import java.util.List;

public class TrainingClientTest {

    @Test
    public void testListTrainingJob() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        List<TrainingJobInfo> jobs = client.training().namespace("user1-ns").list(TrainingJobType.AllTrainingJob);
        for(TrainingJobInfo info : jobs) {
            System.out.println(JSON.toJSONString(info));

            for(com.github.kubeflow.arena.model.training.Instance instance : info.getInstances()) {
                System.out.println(JSON.toJSONString(instance));
            }

        }
    }

}
