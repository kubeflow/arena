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

Arena spark job is based on [spark-on-k8s-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator).You can create spark job image with tool [docker-image-tool](https://spark.apache.org/docs/latest/running-on-kubernetes.html#docker-images).

3\. Install the spark operator.

We recommend using helm to install spark-operator, add the repo:

```
$ helm repo add spark-operator https://googlecloudplatform.github.io/spark-on-k8s-operator
```

Then install it.

```
$ helm install my-release spark-operator/spark-operator --namespace spark-operator --create-namespace
```

This will install the Kubernetes Operator for Apache Spark into the namespace spark-operator. The operator by default watches and handles SparkApplications in every namespaces. If you would like to limit the operator to watch and handle SparkApplications in a single namespace, e.g., default instead, add the following option to the helm install command:

```
$ helm install my-release spark-operator/spark-operator --namespace spark-operator --create-namespace --set sparkJobNamespace=default
```

Please refer [spark-on-k8s-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) to get more details about  how to install spark operator. 


4\. Submit a spark job,the following command shows how to use arena to submit a sparkjob.

```
$ arena submit spark \
   --name=sparktest \
   --image=registry.aliyuncs.com/acs/spark-pi:ack-2.4.5-latest \
   --main-class=org.apache.spark.examples.SparkPi \
   --jar=local:///opt/spark/examples/jars/spark-examples_2.11-2.4.5.jar

   configmap/sparktest-sparkjob created
   configmap/sparktest-sparkjob labeled
   sparkapplication.sparkoperator.k8s.io/sparktest created
   INFO[0001] The Job sparktest has been submitted successfully
   INFO[0001] You can run `arena get sparktest --type sparkjob` to check the job status
```
!!! note

    The Spark version is 2.4.5 which includes the kubernetes client version 4.6.3 and the supported kubernetes versions go all the way up to v1.17.0. If you want to get more details, please refer the https://stackoverflow.com/questions/61565751/why-am-i-not-able-to-run-sparkpi-example-on-a-kubernetes-k8s-cluster.

5\. Get spark job details 

```
$ arena get sparktest
Name:      sparktest
Status:    SUCCEEDED
Namespace: default
Priority:  N/A
Trainer:   SPARKJOB
Duration:  12s

Instances:
  NAME              STATUS     AGE  IS_CHIEF  GPU(Requested)  NODE
  ----              ------     ---  --------  --------------  ----
  sparktest-driver  Completed  12s  true      0               cn-beijing.192.168.8.52

```

6\. Get logs of the spark job.

```
$ arena logs sparktest  --tail 10
21/03/29 07:52:25 WARN ExecutorPodsWatchSnapshotSource: Kubernetes client has been closed (this is expected if the application is shutting down.)
21/03/29 07:52:25 INFO MapOutputTrackerMasterEndpoint: MapOutputTrackerMasterEndpoint stopped!
21/03/29 07:52:25 INFO MemoryStore: MemoryStore cleared
21/03/29 07:52:25 INFO BlockManager: BlockManager stopped
21/03/29 07:52:25 INFO BlockManagerMaster: BlockManagerMaster stopped
21/03/29 07:52:25 INFO OutputCommitCoordinator$OutputCommitCoordinatorEndpoint: OutputCommitCoordinator stopped!
21/03/29 07:52:25 INFO SparkContext: Successfully stopped SparkContext
21/03/29 07:52:25 INFO ShutdownHookManager: Shutdown hook called
21/03/29 07:52:25 INFO ShutdownHookManager: Deleting directory /tmp/spark-e8c052b4-0522-492a-ad77-e6734746c417
21/03/29 07:52:25 INFO ShutdownHookManager: Deleting directory /var/data/spark-349e799f-f21e-4433-afd2-61d3b08875e4/spark-b7824f43-5839-4566-92be-a93014283f31

```

7\. Delete the spark job when it finished.

```
# arena delete sparktest
sparkapplication.sparkoperator.k8s.io "sparktest" deleted
configmap "sparktest-sparkjob" deleted
INFO[0001] The training job sparktest has been deleted successfully
Congratulations! You've run the distributed spark job with ``arena`` successfully. 

```
