package com.github.kubeflow.arena.examples;

import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.TrainingJobStatus;
import com.github.kubeflow.arena.enums.ServingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.common.Logger;
import com.github.kubeflow.arena.model.serving.Instance;
import com.github.kubeflow.arena.model.serving.CustomServingJobBuilder;
import com.github.kubeflow.arena.model.serving.ServingJob;
import com.github.kubeflow.arena.model.serving.ServingJobInfo;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import java.util.concurrent.TimeUnit;

public class CustomServingJobTest {

    public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        // 1.create arena client
        System.out.println("start to test arena-java-sdk.");
        ArenaClient client = new ArenaClient();
        System.out.println("create ArenaClient succeed.");
        //Map<String, String> envs = new HashMap<String, String>();
        //envs.put("batch_size", "128");
        System.out.println("start to create custom serving job.");
        String jobName = "custom-serving-test";
        String jobVersion = "alpha";
        ServingJobType jobType = ServingJobType.CustomServingJob;
        // 2.create mpi job
        ServingJob job = new CustomServingJobBuilder()
                .name(jobName)
                .version(jobVersion)
                .gpus(1)
                .replicas(1)
                .restfulPort(5000)
                .image("happy365/fast-style-transfer:latest")
                .command("python app.py")
                .build();
        if (client.serving().namespace("default").get(jobName,jobType,jobVersion) == null) {
            try {
                client.serving().namespace("default").submit(job);
                System.out.println("create custom serving job succeed.");
            } catch (Exception e) {
                e.printStackTrace();
                System.out.println("create custom serving job failed.");
            }
        }

        System.out.println("waiting custom serving job to be running...");
        int count = 0;
        while (true) {
            if (count >= 30) {
                System.out.println("timeout for waiting custom serving job to be running.");
                throw  new ArenaException(ArenaErrorEnum.UNKNOWN,"time out for waiting custom serving job to be running.");
            }
            count++;
            ServingJobInfo[] jobInfos = client.serving().list(ServingJobType.AllServingJob);
            for (int i = 0;i < jobInfos.length;i++) {
                System.out.println(jobInfos[i]);
            }
            ServingJobInfo jobInfo = client.serving().get(jobName,jobType,jobVersion);
            if (jobInfo.getAvailableInstances() != jobInfo.getDesiredInstances()) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            System.out.println("get custom serving job information: ");
            System.out.println(jobInfo);
            System.out.println("start to get instance logs:");
            Logger logger = new Logger();
            //Logger logger = new Logger();
            Instance[] instances = jobInfo.getInstances();
            for(int i = 0;i < instances.length;i++) {
                System.out.println(instances[i]);
            }
            for(int i = 0;i < instances.length;i++) {
                InputStream stdout = null;
                BufferedReader inReader = null;
                try {
                    stdout = instances[i].getLog(logger);
                    inReader = new BufferedReader(new InputStreamReader(stdout, Charset.defaultCharset()));
                    System.out.println(instances[i].getName());
                    String line = inReader.readLine();
                    while (line != null) {
                        System.out.println(line);
                        line = inReader.readLine();
                    }
                }catch (ArenaException e) {
                    System.out.println(e.getMessage());

                }catch (IOException e) {
                    System.out.println(e.getMessage());
                }finally {
                    if (inReader != null) {
                        inReader.close();
                    }
                    if (stdout != null) {
                        stdout.close();
                    }
                }
            }
            for(int i = 0; i < jobInfo.getEndpoints().length;i++) {
                System.out.println(jobInfo.getEndpoints()[i].getName());
                System.out.println(jobInfo.getEndpoints()[i].getNodePort());
                System.out.println(jobInfo.getEndpoints()[i].getPort());
            }
            break;
        }
        System.out.println("start to delete custom serving job:");
        String output = client.serving().delete(jobName,jobType,jobVersion);
        System.out.println(output);
    }
}