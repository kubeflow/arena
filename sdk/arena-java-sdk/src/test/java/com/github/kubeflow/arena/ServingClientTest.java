package com.github.kubeflow.arena;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.serving.Instance;
import com.github.kubeflow.arena.model.serving.ServingJobInfo;
import org.junit.Test;

import java.io.IOException;
import java.util.List;

public class ServingClientTest {

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

}
