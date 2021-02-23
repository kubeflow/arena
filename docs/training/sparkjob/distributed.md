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

3\. How to use Arena spark job


step1: install spark operator(``arena-system`` is the default namespace,if not exist please create it).

    $ kubectl create -f arena/kubernetes-artifacts/spark-operator/spark-operator.yaml

step2: create rbac of spark job, the spark job need service account ``spark`` to create executors.

    $ kubectl create -f arena/kubernetes-artifacts/spark-operator/spark-rbac.yaml

The default namespace is ``default``. If you want to run spark job in other namespaces. You can change namespace in spark-rbac.yaml and create a new service account.

4\. Submit a spark job,the following command shows how to use arena to submit a sparkjob.

    $ arena submit sparkjob \
        --name=demo \
        --image=registry.aliyuncs.com/acs/spark:v2.4.0 \
        --main-class=org.apache.spark.examples.SparkPi \
        --jar=local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar

    configmap/demo-sparkjob created
    configmap/demo-sparkjob labeled
    sparkapplication.sparkoperator.k8s.io/demo created
    INFO[0005] The Job demo has been submitted successfully
    INFO[0005] You can run `arena get demo --type sparkjob` to check the job status

5\. Get spark job details 

    $ arena get --type=sparkjob demo

    STATUS: SUCCEEDED
    NAMESPACE: default
    TRAINING DURATION: 15s

    NAME   STATUS     TRAINER   AGE  INSTANCE      NODE
    demo1  SUCCEEDED  SPARKJOB  1h   demo1-driver  N/A

6\. Get logs of the spark job.

    $ arena logs -f demo 
    ...
    2019-05-08T08:25:32.104432515Z 2019-05-08 08:25:32 INFO  BlockManagerMaster:54 - BlockManagerMaster stopped
    2019-05-08T08:25:32.10761075Z 2019-05-08 08:25:32 INFO  OutputCommitCoordinator$OutputCommitCoordinatorEndpoint:54 - OutputCommitCoordinator stopped!
    2019-05-08T08:25:32.114734944Z 2019-05-08 08:25:32 INFO  SparkContext:54 - Successfully stopped SparkContext
    2019-05-08T08:25:32.117170277Z 2019-05-08 08:25:32 INFO  ShutdownHookManager:54 - Shutdown hook called
    2019-05-08T08:25:32.118273045Z 2019-05-08 08:25:32 INFO  ShutdownHookManager:54 - Deleting directory /tmp/spark-bdb4e416-5ab7-420c-905e-ef43c30fb187
    2019-05-08T08:25:32.120019227Z 2019-05-08 08:25:32 INFO  ShutdownHookManager:54 - Deleting directory /var/data/spark-118b216d-2d39-4287-ad71-5b5d7c7195c9/spark-06dbab1f-13aa-474c-a1db-8845e14627bf


7\. Delete the spark job when it finished.

    $ arena delete --type=sparkjob demo 

    sparkapplication.sparkoperator.k8s.io "demo1" deleted
    time="2019-05-08T17:27:06+08:00" level=info msg="The Job demo1 has been deleted successfully"
    configmap "demo1-sparkjob" deleted

Congratulations! You've run the distributed spark job with ``arena`` successfully. 
