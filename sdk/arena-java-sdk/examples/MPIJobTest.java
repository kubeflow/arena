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

public class MPIJobTest {

    public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        // 1.create arena client
        System.out.println("start to test arena-java-sdk.");
        ArenaClient client = new ArenaClient();
        System.out.println("create ArenaClient succeed.");
        //Map<String, String> envs = new HashMap<String, String>();
        //envs.put("batch_size", "128");
        System.out.println("start to create mpi job.");
        // 2.create mpi job
        Task mpiTask = new MPITask()
                .withCount(1)
                .withGpu(1)
                .withTensorboard()
               // .withEnvs(envs)
                .withImage("registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5")
               // .withCommand("sh -c -- 'count=0;for i in $(seq 1 100);do sleep 10;echo count:$i;done'");
                .withCommand("sh -c -- 'mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10'");

        Training mpi = new MPITraining("mpi-dist").WithMPITask(mpiTask);
        if (client.getTraining("mpi-dist","mpijob") == null) {
            try {
                client.createTraining(mpi);
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
                throw  new ArenaException("time out for waiting mpi training job to be running.");
            }
            count++;
            mpi = client.getTraining("mpi-dist","mpijob");
            if (mpi.getStatus().equals("PENDING")) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            if (mpi.getTensorboardURL() != null) {
                System.out.println("tensorboard url is: " + mpi.getTensorboardURL());
            }
            /*
            if (mpi.getStatus().equals("RUNNING")) {
                TimeUnit.SECONDS.sleep(10);
                continue;
            }
            */
            System.out.println("get mpi training job information: ");
            System.out.println(mpi);
            System.out.println("start to get instance logs:");
            Logger logger = new Logger().follow().followTimeout(60);
            //Logger logger = new Logger();
            Instance[] instances = mpi.getInstances();
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
        String output = client.deleteTraining("mpi-dist","mpijob");
        System.out.println(output);
    }
}
