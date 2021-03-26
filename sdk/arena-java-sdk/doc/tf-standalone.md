## API for Standalone TFJob

we will introduce a pipeline about standalone tfjob operations, it contains: 
* create arena client
* create  standalone tfjob
* get information of standalone tfjob
* get instances logs of standalone tfjob
* delete standalone tfjob

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

### 2.create standalone tfjob

there is an example for creating standalone tfjob:
```
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.StandaloneTraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
            ArenaClient client = new ArenaClient();
            Task standaloneTask = new StandaloneTask()
                .withGpu(1)
                .withTensorboard()
                .withImage("registry.cn-hangzhou.aliyuncs.com/happy365/tensorflow:1.5.0-devel-gpu")
               // .withCommand("sh -c -- 'count=0;for i in $(seq 1 100);do sleep 10;echo count:$i;done'");
                .withCommand("sh -c -- 'python tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 10000 --data_dir=tensorflow-sample-code/data'");
        Training standalone = new StandaloneTraining("tf-standalone-test").WithStandaloneTask(standaloneTask);
        client.createTraining(standalone);
     }
}
```
it includes four steps: 

* (1) create arena client 
* (2) create StandaloneTask 
* (3) areate StandaloneTraining
* (4) use arena client to create job

creating standaloneTask includes some options:

|  function name  |   description  | 
|:---|:--:| 
| withGpu(int)   | how many gpus that app can use in a node  |  
|  withTensorboard()  | create tensorboard    |
|  withEnvs(Map<String, String>)  |  set envs for standalone tfjob  |
|  withCpu(String)  |  how many cpus that app can use in a node,like "200m","500m"  |
| withMemory(String)   | how many memory that app can use in a node,like "4Gi","300Mi"   |
| withImage(String)   | set image   | 
| withCommand(String)   | set command   |
| withDataDirs(Map<String, String> dataDirs)|the data dir|
|withDataSources(Map<String, String> dataSources)|specify the datasource to mount to the job|
| withLogDir(String)   | set the training logs dir |

### 3.get information of standalone tfjob 

you can get the information of standalone tfjob that you  have created:

```
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.StandaloneTraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        ArenaClient client = new ArenaClient();
        Training standalone = client.getTraining("tf-standalone-dist","tfjob");
        System.out.println(standalone)
     }
}
```
two parameters are required for getting the job information,they are job name and job type. "tf-standalone-dist" is the job name in the example.

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


### 4.get instances logs of standalone tfjob

if you want to get the instances logs of standalone tfjob,the following codes can help you:
```
import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.StandaloneTraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.training.Instance;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
          // create arena client 
          ArenaClient client = new ArenaClient();
          // get standalone training
          Training standalone = client.getTraining("tf-standalone-dist","tfjob");
          // create Logger object
          // get standalone training instances
          Instance[] instances = standalone.getInstances();
          Logger logger = new Logger().follow().followTimeout(600);
          Instance[] instances = standalone.getInstances();
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
* (3) get standalone training
* (4) print instances logs of standalone tfjob traning

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

### 5.delete standalone tfjob

the following code displays how to delete a standalone tfjob:

```
import com.aliyun.ack.arena.client.ArenaClient;
import com.aliyun.ack.arena.exceptions.ArenaException;
import com.aliyun.ack.arena.model.training.StandaloneTraining;
import com.aliyun.ack.arena.model.training.Training;
import com.aliyun.ack.arena.model.common.*;
public class MPIJobTest {
     public static void main(String[] args) throws IOException,ArenaException,InterruptedException {
        ArenaClient client = new ArenaClient();
        String output = client.deleteTraining("tf-standalone-dist","tfjob");
        System.out.println(output)
     }
}
```

"tf-standalone-dist" is the name of standalone tfjob.


all api are completed,Please refer to com/aliyun/ack/arena/examples/StandaloneJobTest.java for the complete pipeline.
