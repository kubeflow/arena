package com.github.kubeflow.arena.examples;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import java.util.concurrent.TimeUnit;
import com.github.kubeflow.arena.client.ArenaClient;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.enums.TrainingJobStatus;
import com.github.kubeflow.arena.enums.TrainingJobType;
import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.model.training.*;
import com.github.kubeflow.arena.model.common.*;

import java.io.IOException;

public class MPIJobTest {

    public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        // 1.create arena client
        System.out.println("start to test arena-java-sdk.");
        ArenaClient client = new ArenaClient();
        System.out.println("create ArenaClient succeed.");
        //Map<String, String> envs = new HashMap<String, String>();
        //envs.put("batch_size", "128");
        System.out.println("start to create mpi job.");
        String jobName = "mpi-dist";
        TrainingJobType jobType = TrainingJobType.MPITrainingJob;
        // 2.create mpi job
        TrainingJob job = new MPIJobBuilder()
                .name(jobName)
                .workers(1)
                .gpus(1)
                .enableTensorboard()
                .cpu("3000m")
                .memory("4Gi")
                .image("registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5")
                .command("sh -c -- 'mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10'")
                .build();
        if (client.training().namespace("default").get(jobName,jobType) == null) {
            try {
                client.training().namespace("default").submit(job);
                System.out.println("create mpi training job succeed.");
            } catch (Exception e) {
                e.printStackTrace();
                System.out.println("create mpi training job failed.");
            }
        }

        System.out.println("waiting mpi job to be running...");
        int count = 0;
        while (true) {
            if (count >= 30) {
                System.out.println("timeout for waiting mpi training job to be running.");
                throw  new ArenaException(ArenaErrorEnum.UNKNOWN,"time out for waiting mpi training job to be running.");
            }
            count++;
            TrainingJobInfo[] jobInfos = client.training().list(TrainingJobType.AllTrainingJob,true);
            for (int i = 0;i < jobInfos.length;i++) {
                System.out.print(jobInfos[i]);
            }
            TrainingJobInfo jobInfo = client.training().get(jobName,jobType);
            if (jobInfo.getStatus().equals(TrainingJobStatus.TrainingJobPending)) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            if (jobInfo.getTensorboard() != null && jobInfo.getTensorboard().length() != 0) {
                System.out.println("tensorboard url is: " + jobInfo.getTensorboard());
            }
            System.out.println("get mpi training job information: ");
            System.out.println(jobInfo);
            System.out.println("start to get instance logs:");
            Logger logger = new Logger().follow().followTimeout(60);
            //Logger logger = new Logger();
            Instance[] instances = jobInfo.getInstances();
            for(int i = 0;i < instances.length;i++) {
                System.out.println(instances[i]);
            }
            for(int i = 0;i < instances.length;i++) {
                if(instances[i].isChief() == false) {
                    continue;
                }
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
            break;
        }
        System.out.println("start to delete mpi training job:");
        String output = client.training().delete(jobName,jobType);
        System.out.println(output);
    }
}
