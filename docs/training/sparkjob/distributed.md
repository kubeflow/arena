# Submit a distributed spark job

Arena supports and simplifies distributed spark job.

1\. To run a distributed spark job, you need to specify:

- The spark job image which contains the main class jar. (required)
- Main class of your jar. (required)
- Jar path in the container.(required)
- The number of executors.(default: 1)
- The resource cpu request of driver pod (default: 1)
- The resource memory request of driver pod (default: 500m)
- The resource cpu request of executor pod (default: 1)
- The resource memory request of executor pod (default: 500m)

2\. How to create spark job image.

Arena spark job is based on [spark-operator](https://github.com/kubeflow/spark-operator).You can create spark job image with tool [docker-image-tool](https://spark.apache.org/docs/latest/running-on-kubernetes.html#docker-images).

3\. Install the spark operator.

Please refer [Spark operator documentation](https://www.kubeflow.org/docs/components/spark-operator) to get more details about how to install spark operator.

4\. Submit a spark job,the following command shows how to use arena to submit a sparkjob.

```bash
$ arena submit sparkjob \
   --name=spark-pi \
   --image=spark:3.5.0 \
   --jar=local:///opt/spark/examples/jars/spark-examples_2.12-3.5.0.jar \
   --main-class=org.apache.spark.examples.SparkPi \
   --spark-version=3.5.0 \
   --driver-cpu-request=1 \
   --driver-memory-request=500m \
   --driver-service-account=spark-operator-spark \
   --replicas=1 \
   --executor-cpu-request=1 \
   --executor-memory-request=500m
sparkapplication.sparkoperator.k8s.io/spark-pi created
INFO[0002] The Job spark-pi has been submitted successfully 
INFO[0002] You can run `arena get spark-pi --type sparkjob -n default` to check the job status
```

5\. Get spark job details

```bash
$ bin/arena get spark-pi     
Name:        spark-pi
Status:      SUCCEEDED
Namespace:   default
Priority:    N/A
Trainer:     SPARKJOB
Duration:    9s
CreateTime:  2024-07-30 10:56:25
EndTime:     2024-07-30 10:56:34

Instances:
  NAME             STATUS     AGE  IS_CHIEF  GPU(Requested)  NODE
  ----             ------     ---  --------  --------------  ----
  spark-pi-driver  Completed  9s   true      0               cn-hongkong.172.21.122.2
```

6\. Get logs of the spark job.

```bash
$ arena logs spark-pi  --tail 10
24/07/30 02:56:33 WARN ExecutorPodsWatchSnapshotSource: Kubernetes client has been closed.
24/07/30 02:56:33 INFO MapOutputTrackerMasterEndpoint: MapOutputTrackerMasterEndpoint stopped!
24/07/30 02:56:33 INFO MemoryStore: MemoryStore cleared
24/07/30 02:56:33 INFO BlockManager: BlockManager stopped
24/07/30 02:56:33 INFO BlockManagerMaster: BlockManagerMaster stopped
24/07/30 02:56:33 INFO OutputCommitCoordinator$OutputCommitCoordinatorEndpoint: OutputCommitCoordinator stopped!
24/07/30 02:56:33 INFO SparkContext: Successfully stopped SparkContext
24/07/30 02:56:33 INFO ShutdownHookManager: Shutdown hook called
24/07/30 02:56:33 INFO ShutdownHookManager: Deleting directory /tmp/spark-271a7ffc-3be4-4f5d-b6dc-082f4f046cb1
24/07/30 02:56:33 INFO ShutdownHookManager: Deleting directory /var/data/spark-1b0e1d8b-4dce-43aa-8e5c-20ea82490648/spark-0b1ae353-731c-4bc5-b4fe-0f21c92adfb0
```

7\. Delete the spark job when it finished.

```bash
$ arena delete spark-pi
INFO[0002] The training job spark-pi has been deleted successfully
```
