## API for MPI Job

we will introduce a pipeline about mpi job operations, it contains: 
* create arena client
* create mpi job
* get information of mpi job
* get instances logs of mpi job
* delete mpi job

### 1.create arena client

if you want to use arena-java-sdk,you should create arena client firstly:
```
import com.aliyun.ack.arena.client.ArenaClient;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
         ArenaClient client = new ArenaClient(); 
     }
}
```
you can assign kubeconfig file and namespace,like:
```
ArenaClient client = new ArenaClient("~/.kube/config","default");
```
"~/.kube/config" is the kubeconfig file,"default" is the namespace name.

### 2.create mpi job

there is an example for creating mpi job:
```
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.MPITraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
            ArenaClient client = new ArenaClient();
            Map<String, String> envs = new HashMap<String, String>();
            envs.put("batch_size", "128");
            Task mpiTask = new MPITask()
                .withCount(2)
                .withGpu(1)
                .withTensorboard()
                .withEnvs(envs)
                .withImage("registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5")
                .withCommand("sh -c -- 'mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10'");

        Training mpi = new MPITraining("mpi-dist").WithMPITask(mpiTask);
        client.createTraining(mpi);
     }
}
```
it includes four steps: 

* (1) create arena client 
* (2) create mpiTask 
* (3) areate MPITraining 
* (4) use arena client to create job

creating mpiTask includes some options:

|  function name  |   description  | 
|:---|:--:|
| withCount(int)   | total of workers | 
| withGpu(int)   | how many gpus that app can use in a node  |  
|  withTensorboard()  | create tensorboard    |
|  withEnvs(Map<String, String>)  |  set envs for mpi job  | 
| withImage(String)   | set image   | 
| withCommand(String)   | set command   |
|  withCpu(String)  |  how many cpus that app can use in a node,like "200m","500m"  |
| withMemory(String)   | how many memory that app can use in a node,like "4Gi","300Mi"   |
| withLogDir(String)   | set the training logs dir |
| withDataDirs(Map<String, String> dataDirs)|the data dir|
|withDataSources(Map<String, String> dataSources)|specify the datasource to mount to the job|

### 3.get information of mpi job 

you can get the information of mpi job that you  have created:

```
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.MPITraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        ArenaClient client = new ArenaClient();
        Training mpi = client.getTraining("mpi-dist","mpijob");
        System.out.println(mpi)
     }
}
```
two parameters are required for getting the job information,they are job name and job type. "mpi-dist" is the job name in the example.

the Training object has some attributes,you can use following functions to access them:

| function name    | return type   |  description  | 
|:---|:--:| :---: |
| getNamespace()  | String | get the namespace of training   |    
|  getDuration()  | String   |  get the age of training|    
|  getTensorboardURL()  | String    | get the url of tensorboard |   
| getPriority()   | String   |  get the priority of training  |
| getStatus()   | String   | get the status of training |   
| getName()  | String   | get the name of training |   
| getCheifName()   | String   | get the chief pod |    
| getInstances()   | []Instance   |  get the instances of training  |
| getType()   | String   | get the type of training |    


### 4.get instances logs of mpi job

if you want to get the instances logs of mpi job,the following codes can help you:
```
import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.MPITraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.training.Instance;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
          // create arena client 
          ArenaClient client = new ArenaClient();
          // get mpi training
          Training mpi = client.getTraining("mpi-dist","mpijob");
          // get mpi training instances
          Instance[] instances = mpi.getInstances();
          Logger logger = new Logger().follow().followTimeout(60);
          Instance[] instances = mpi.getInstances();
          for(int i = 0;i < instances.length;i++) {
              InputStream  stdout = instances[i].getLog(logger);
              BufferedReader inReader = new BufferedReader(new InputStreamReader(stdout, Charset.defaultCharset()));
              System.out.println(instances[i].getName());
              String line = inReader.readLine();
              while (line != null) {
                  System.out.println(line);
                  line = inReader.readLine();
              }
              inReader.close();
              stdout.close();
          }
        for(int i = 0;i < instances.length;i++) {
            // get log of instance
            String output = instances[i].getLog(logger);
            System.out.println(output);
        }
     }
}

```
this code includes three steps:

* (1) create arena client 
* (2) create Logger 
* (3) get mpi training
* (4) print instances logs of mpi training

Instance Object has some attributes,you can use following functions to access:

|  function name  |  return value  |  description  |
|:---|:--:|:---:|
|  getOwner()  | String   | the training name which instance is belong to   |
| getNamespace()   | String   | get namespace   |
| getName()   |  String  | get instance name   |
|  getAge()  | String   |  get the age of instance  |
|  getStatus()  | String   | get the instance status   |
|  getNode()  | String   |  get the node name which instance is belong to   |
| isChief() | bool| is chief instance or not |
| getLog(Logger logger)| String |get the instance log|

if you want to get instance log,creating Logger object is required firstly.

"Logger" is an Object which gather options for getting instance log.

"Logger" includes some functions you can use:

| function name   |  description  | 
|:---|:--:|
| follow()   |  follow the instance logs,like 'tail -f'  |     
|  tail(int lines)  | get the last n rows of logs  | 
| followTimeout(int timeout)|set the timeout when following instance log|
| container(String name)| assign container,default is first container of instance|  

the following example displays how to get last 10 rows of logs:

```
// skip other code
....
    Logger logger = new Logger().tail(10);
``` 

### 5.delete mpi job

the following code displays how to delete a mpi job:

```
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.MPITraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        ArenaClient client = new ArenaClient();
        String output = client.deleteTraining("mpi-dist","mpijob");
        System.out.println(output)
     }
}
```

"mpi-dist" is the name of mpi job.


all api are completed,Please refer to com/aliyun/ack/arena/examples/MPIJobTest.java for the complete pipeline.
