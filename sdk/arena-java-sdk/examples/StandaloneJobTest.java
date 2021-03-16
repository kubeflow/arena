
import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import java.util.concurrent.TimeUnit;
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.Instance;
import com.aliyun.ack.arena.model.common.*;

import java.io.IOException;
import java.util.HashMap;

import java.util.Map;

public class StandaloneJobTest {

    public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        // 1.create arena client
        System.out.println("start to test arena-java-sdk.");
        ArenaClient client = new ArenaClient();
        System.out.println("create ArenaClient succeed.");
        System.out.println("start to create tf standalone job.");
        // 2.create standalone tfjob
        Task standaloneTask = new StandaloneTask()
                .withGpu(1)
                .withTensorboard()
                .withImage("registry.cn-hangzhou.aliyuncs.com/happy365/tensorflow:1.5.0-devel-gpu")
               // .withCommand("sh -c -- 'count=0;for i in $(seq 1 100);do sleep 10;echo count:$i;done'");
                .withCommand("sh -c -- 'python tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 10000 --data_dir=tensorflow-sample-code/data'");

        Training standalone = new StandaloneTraining("tf-standalone-test").WithStandaloneTask(standaloneTask);
        if (client.getTraining("tf-standalone-test","tfjob") == null) {
            try {
                client.createTraining(standalone);
                System.out.println("create tf standalone training job succeed.");
            } catch (Exception e) {
                e.printStackTrace();
                System.out.println("create tf standalone training job failed.");
            }
        }

        System.out.println("waiting tf standalone job to be running...");
        int count = 0;
        while (true) {
            if (count >= 30) {
                System.out.println("timeout for waiting tf standalone training job to be running.");
                throw  new ArenaException("time out for waiting tf standalone training job to be running.");
            }
            count++;
            standalone = client.getTraining("tf-standalone-test","tfjob");
            if (standalone == null) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            if (standalone.getStatus().equals("PENDING")) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            if (standalone.getTensorboardURL() != null) {
                System.out.println("tensorboard url is: " + standalone.getTensorboardURL());
            }
            /*
            if (mpi.getStatus().equals("RUNNING")) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            */
            System.out.println("get tf standalone training job information: ");
            System.out.println(standalone);
            System.out.println("start to get instance logs:");
            Logger logger = new Logger().follow().followTimeout(600);
            //Logger logger = new Logger();
            Instance[] instances = standalone.getInstances();
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
        System.out.println("start to delete tf standalone training job:");
        String output = client.deleteTraining("tf-standalone-test","tfjob");
        System.out.println(output);
    }
}
