package com.github.kubeflow.arena;

import com.alibaba.fastjson.JSON;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.evaluate.EvaluateJob;
import com.github.kubeflow.arena.model.evaluate.EvaluateJobBuilder;
import com.github.kubeflow.arena.model.evaluate.EvaluateJobInfo;
import org.junit.Test;

import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class EvaluateClientTest {

    @Test
    public void testSubmit() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();

        Map<String, String> datas = new HashMap<>();
        datas.put("model-pvc", "/data");

        String command = null;

        EvaluateJob job = new EvaluateJobBuilder()
                .name("test-evaluate2")
                .namespace("default-group")
                .image("registry.cn-beijing.aliyuncs.com/hzx-kubeflow/evaluate_test_tensorflow1.15:1.0.3")
                .datas(datas)
                .modelName("bert1")
                .modelPath("registry.cn-beijing.aliyuncs.com/hzx-kubeflow/evaluate_test_tensorflow1.15:1.0.3")
                .modelVersion("1")
                .metricsPath("/data/evaluate/output")
                .datasetPath("/data/evaluate/dataset")
                .command(command)
                .build();

        String result = client.evaluate().namespace("default-group").submit(job);
        System.out.println(result);
    }

    @Test
    public void testGet() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        EvaluateJobInfo jobInfo = client.evaluate().namespace("default-group").get("test-evaluate2");
        System.out.println(JSON.toJSONString(jobInfo));
    }

    @Test
    public void testList() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        List<EvaluateJobInfo> jobs = client.evaluate().namespace("default-group").list();
        for(EvaluateJobInfo job : jobs) {
            System.out.println(JSON.toJSONString(job));
        }
    }

    @Test
    public void testDelete() throws ArenaException, IOException {
        ArenaClient client = new ArenaClient();
        String result = client.evaluate().namespace("default-group").delete("test-evaluate2");
        System.out.println(result);
    }

}
