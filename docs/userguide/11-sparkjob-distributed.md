
Arena supports and simplifies distributed spark job. 

### 1. To run a distributed spark job, you need to specify:
- The spark job image which contains the main class jar. (required)
- Main class of your jar. (required)
- Jar path in the container.(required)
- The number of executors.(default: 1)
- The resource cpu request of driver pod (default: 1)
- The resource memory request of driver pod (default: 500m)
- The resource cpu request of executor pod (default: 1)
- The resource memory request of executor pod (default: 500m)

### 2. How to create spark job image. 

Arena spark job is based on spark-on-k8s-operator(https://github.com/GoogleCloudPlatform/spark-on-k8s-operator).You can create spark job image with tool `docker-image-tool` (https://spark.apache.org/docs/latest/running-on-kubernetes.html#docker-images)

### 3. How to use Arena spark job

##### install spark operator 
```$xslt
# arena-system is the default namespace,if not exist please create it.
kubectl create -f arena/kubernetes-artifacts/spark-operator/spark-operator.yaml
```

##### create rbac of spark job 
The spark job need service account `spark` to create executors.
```$xslt
kubectl create -f arena/kubernetes-artifacts/spark-operator/spark-rbac.yaml
```
The default namespace is `default`. If you want to run spark job in other namespaces. You can change namespace in spark-rbac.yaml and create a new service account.
##### submit a spark job 
```$xslt
arena submit sparkjob --name=demo --image=registry.aliyuncs.com/acs/spark:v2.4.0 --main-class=org.apache.spark.examples.SparkPi --jar=local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar
```
The result is like below.
```$xslt
configmap/demo-sparkjob created
configmap/demo-sparkjob labeled
sparkapplication.sparkoperator.k8s.io/demo created
INFO[0005] The Job demo has been submitted successfully
INFO[0005] You can run `arena get demo --type sparkjob` to check the job status
```
##### get spark job status 
```$xslt
arena get --type=sparkjob demo
```
When the job succeed,you will see the result below.
```$xslt
STATUS: SUCCEEDED
NAMESPACE: default
TRAINING DURATION: 15s

NAME   STATUS     TRAINER   AGE  INSTANCE      NODE
demo1  SUCCEEDED  SPARKJOB  1h   demo1-driver  N/A
```

##### watch log of spark job
```$xslt
arena logs -f demo 
```
You will get the log of spark driver pod.
```$xslt
2019-05-08T08:25:21.904409561Z ++ id -u
2019-05-08T08:25:21.904639867Z + myuid=0
2019-05-08T08:25:21.904649704Z ++ id -g
2019-05-08T08:25:21.904901542Z + mygid=0
2019-05-08T08:25:21.904909072Z + set +e
2019-05-08T08:25:21.905241846Z ++ getent passwd 0
2019-05-08T08:25:21.905608733Z + uidentry=root:x:0:0:root:/root:/bin/ash
2019-05-08T08:25:21.905623028Z + set -e
2019-05-08T08:25:21.905629226Z + '[' -z root:x:0:0:root:/root:/bin/ash ']'
2019-05-08T08:25:21.905633894Z + SPARK_K8S_CMD=driver
2019-05-08T08:25:21.905757494Z + case "$SPARK_K8S_CMD" in
2019-05-08T08:25:21.90622059Z + shift 1
2019-05-08T08:25:21.906232126Z + SPARK_CLASSPATH=':/opt/spark/jars/*'
2019-05-08T08:25:21.906236316Z + env
2019-05-08T08:25:21.906239651Z + grep SPARK_JAVA_OPT_
2019-05-08T08:25:21.90624307Z + sort -t_ -k4 -n
2019-05-08T08:25:21.906585896Z + sed 's/[^=]*=\(.*\)/\1/g'
2019-05-08T08:25:21.906908601Z + readarray -t SPARK_EXECUTOR_JAVA_OPTS
2019-05-08T08:25:21.906917535Z + '[' -n '' ']'
2019-05-08T08:25:21.906999069Z + '[' -n '' ']'
2019-05-08T08:25:21.907003871Z + PYSPARK_ARGS=
2019-05-08T08:25:21.907006605Z + '[' -n '' ']'
2019-05-08T08:25:21.907008951Z + R_ARGS=
2019-05-08T08:25:21.907012105Z + '[' -n '' ']'
2019-05-08T08:25:21.907148385Z + '[' '' == 2 ']'
2019-05-08T08:25:21.907994286Z + '[' '' == 3 ']'
2019-05-08T08:25:21.908014459Z + case "$SPARK_K8S_CMD" in
2019-05-08T08:25:21.908018653Z + CMD=("$SPARK_HOME/bin/spark-submit" --conf "spark.driver.bindAddress=$SPARK_DRIVER_BIND_ADDRESS" --deploy-mode client "$@")
2019-05-08T08:25:21.908023924Z + exec /sbin/tini -s -- /opt/spark/bin/spark-submit --conf spark.driver.bindAddress=172.20.90.160 --deploy-mode client --properties-file /opt/spark/conf/spark.properties --class org.apache.spark.examples.SparkPi spark-internal
2019-05-08T08:25:23.326681135Z 2019-05-08 08:25:23 WARN  NativeCodeLoader:62 - Unable to load native-hadoop library for your platform... using builtin-java classes where applicable
2019-05-08T08:25:23.829843117Z 2019-05-08 08:25:23 INFO  SparkContext:54 - Running Spark version 2.4.0
2019-05-08T08:25:23.8529898Z 2019-05-08 08:25:23 INFO  SparkContext:54 - Submitted application: Spark Pi
2019-05-08T08:25:23.94670344Z 2019-05-08 08:25:23 INFO  SecurityManager:54 - Changing view acls to: root
2019-05-08T08:25:23.946735076Z 2019-05-08 08:25:23 INFO  SecurityManager:54 - Changing modify acls to: root
2019-05-08T08:25:23.946740267Z 2019-05-08 08:25:23 INFO  SecurityManager:54 - Changing view acls groups to: 
2019-05-08T08:25:23.946744543Z 2019-05-08 08:25:23 INFO  SecurityManager:54 - Changing modify acls groups to: 
2019-05-08T08:25:23.946748767Z 2019-05-08 08:25:23 INFO  SecurityManager:54 - SecurityManager: authentication disabled; ui acls disabled; users  with view permissions: Set(root); groups with view permissions: Set(); users  with modify permissions: Set(root); groups with modify permissions: Set()
2019-05-08T08:25:24.273960575Z 2019-05-08 08:25:24 INFO  Utils:54 - Successfully started service 'sparkDriver' on port 7078.
2019-05-08T08:25:24.307632934Z 2019-05-08 08:25:24 INFO  SparkEnv:54 - Registering MapOutputTracker
2019-05-08T08:25:24.339548141Z 2019-05-08 08:25:24 INFO  SparkEnv:54 - Registering BlockManagerMaster
2019-05-08T08:25:24.339577986Z 2019-05-08 08:25:24 INFO  BlockManagerMasterEndpoint:54 - Using org.apache.spark.storage.DefaultTopologyMapper for getting topology information
2019-05-08T08:25:24.340887925Z 2019-05-08 08:25:24 INFO  BlockManagerMasterEndpoint:54 - BlockManagerMasterEndpoint up
2019-05-08T08:25:24.359682519Z 2019-05-08 08:25:24 INFO  DiskBlockManager:54 - Created local directory at /var/data/spark-118b216d-2d39-4287-ad71-5b5d7c7195c9/blockmgr-5532fd8b-64b9-492c-b94d-308b55d60a71
2019-05-08T08:25:24.388529744Z 2019-05-08 08:25:24 INFO  MemoryStore:54 - MemoryStore started with capacity 110.0 MB
2019-05-08T08:25:24.413347888Z 2019-05-08 08:25:24 INFO  SparkEnv:54 - Registering OutputCommitCoordinator
2019-05-08T08:25:24.560654618Z 2019-05-08 08:25:24 INFO  log:192 - Logging initialized @2462ms
2019-05-08T08:25:24.654721075Z 2019-05-08 08:25:24 INFO  Server:351 - jetty-9.3.z-SNAPSHOT, build timestamp: unknown, git hash: unknown
2019-05-08T08:25:24.680943254Z 2019-05-08 08:25:24 INFO  Server:419 - Started @2586ms
2019-05-08T08:25:24.715867156Z 2019-05-08 08:25:24 INFO  AbstractConnector:278 - Started ServerConnector@7e97551f{HTTP/1.1,[http/1.1]}{0.0.0.0:4040}
2019-05-08T08:25:24.715897312Z 2019-05-08 08:25:24 INFO  Utils:54 - Successfully started service 'SparkUI' on port 4040.
2019-05-08T08:25:24.76123501Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@1450078a{/jobs,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.762173789Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@534ca02b{/jobs/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.763361524Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@29a23c3d{/jobs/job,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.764374535Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@6fe46b62{/jobs/job/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.764919809Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@591fd34d{/stages,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.765687152Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@61e45f87{/stages/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.766434602Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@7c9b78e3{/stages/stage,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.769934319Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@5491f68b{/stages/stage/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.769949155Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@736ac09a{/stages/pool,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.769966711Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@6ecd665{/stages/pool/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.77037559Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@45394b31{/storage,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.772696599Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@1ec7d8b3{/storage/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.772709487Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@3b0ca5e1{/storage/rdd,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.773014833Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@5bb3131b{/storage/rdd/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.77546416Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@54dcbb9f{/environment,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.775478151Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@74fef3f7{/environment/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.775882882Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@2a037324{/executors,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.780702953Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@69eb86b4{/executors/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.780717178Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@585ac855{/executors/threadDump,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.78072195Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@5bb8f9e2{/executors/threadDump/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.793805533Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@6a933be2{/static,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.808511998Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@378bd86d{/,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.808532751Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@2189e7a7{/api,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.808537695Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@644abb8f{/jobs/job/kill,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.80854206Z 2019-05-08 08:25:24 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@1a411233{/stages/stage/kill,null,AVAILABLE,@Spark}
2019-05-08T08:25:24.808546336Z 2019-05-08 08:25:24 INFO  SparkUI:54 - Bound SparkUI to 0.0.0.0, and started at http://demo1-1557303918993-driver-svc.default.svc:4040
2019-05-08T08:25:24.834767942Z 2019-05-08 08:25:24 INFO  SparkContext:54 - Added JAR file:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar at spark://demo1-1557303918993-driver-svc.default.svc:7078/jars/spark-examples_2.11-2.4.0.jar with timestamp 1557303924832
2019-05-08T08:25:26.274526541Z 2019-05-08 08:25:26 INFO  ExecutorPodsAllocator:54 - Going to request 1 executors from Kubernetes.
2019-05-08T08:25:26.455658752Z 2019-05-08 08:25:26 INFO  Utils:54 - Successfully started service 'org.apache.spark.network.netty.NettyBlockTransferService' on port 7079.
2019-05-08T08:25:26.47651031Z 2019-05-08 08:25:26 INFO  NettyBlockTransferService:54 - Server created on demo1-1557303918993-driver-svc.default.svc:7079
2019-05-08T08:25:26.476533172Z 2019-05-08 08:25:26 INFO  BlockManager:54 - Using org.apache.spark.storage.RandomBlockReplicationPolicy for block replication policy
2019-05-08T08:25:26.503099521Z 2019-05-08 08:25:26 INFO  BlockManagerMaster:54 - Registering BlockManager BlockManagerId(driver, demo1-1557303918993-driver-svc.default.svc, 7079, None)
2019-05-08T08:25:26.506168762Z 2019-05-08 08:25:26 INFO  BlockManagerMasterEndpoint:54 - Registering block manager demo1-1557303918993-driver-svc.default.svc:7079 with 110.0 MB RAM, BlockManagerId(driver, demo1-1557303918993-driver-svc.default.svc, 7079, None)
2019-05-08T08:25:26.529524775Z 2019-05-08 08:25:26 INFO  BlockManagerMaster:54 - Registered BlockManager BlockManagerId(driver, demo1-1557303918993-driver-svc.default.svc, 7079, None)
2019-05-08T08:25:26.529543725Z 2019-05-08 08:25:26 INFO  BlockManager:54 - Initialized BlockManager: BlockManagerId(driver, demo1-1557303918993-driver-svc.default.svc, 7079, None)
2019-05-08T08:25:26.661414752Z 2019-05-08 08:25:26 INFO  ContextHandler:781 - Started o.s.j.s.ServletContextHandler@4c777e7b{/metrics/json,null,AVAILABLE,@Spark}
2019-05-08T08:25:30.459756195Z 2019-05-08 08:25:30 INFO  KubernetesClusterSchedulerBackend$KubernetesDriverEndpoint:54 - Registered executor NettyRpcEndpointRef(spark-client://Executor) (172.20.90.161:52168) with ID 1
2019-05-08T08:25:30.534179215Z 2019-05-08 08:25:30 INFO  KubernetesClusterSchedulerBackend:54 - SchedulerBackend is ready for scheduling beginning after reached minRegisteredResourcesRatio: 0.8
2019-05-08T08:25:30.679510273Z 2019-05-08 08:25:30 INFO  BlockManagerMasterEndpoint:54 - Registering block manager 172.20.90.161:36718 with 110.0 MB RAM, BlockManagerId(1, 172.20.90.161, 36718, None)
2019-05-08T08:25:30.906713226Z 2019-05-08 08:25:30 INFO  SparkContext:54 - Starting job: reduce at SparkPi.scala:38
2019-05-08T08:25:30.93537711Z 2019-05-08 08:25:30 INFO  DAGScheduler:54 - Got job 0 (reduce at SparkPi.scala:38) with 2 output partitions
2019-05-08T08:25:30.936000643Z 2019-05-08 08:25:30 INFO  DAGScheduler:54 - Final stage: ResultStage 0 (reduce at SparkPi.scala:38)
2019-05-08T08:25:30.936506781Z 2019-05-08 08:25:30 INFO  DAGScheduler:54 - Parents of final stage: List()
2019-05-08T08:25:30.938152322Z 2019-05-08 08:25:30 INFO  DAGScheduler:54 - Missing parents: List()
2019-05-08T08:25:30.958509715Z 2019-05-08 08:25:30 INFO  DAGScheduler:54 - Submitting ResultStage 0 (MapPartitionsRDD[1] at map at SparkPi.scala:34), which has no missing parents
2019-05-08T08:25:31.128459296Z 2019-05-08 08:25:31 INFO  MemoryStore:54 - Block broadcast_0 stored as values in memory (estimated size 1936.0 B, free 110.0 MB)
2019-05-08T08:25:31.172704042Z 2019-05-08 08:25:31 INFO  MemoryStore:54 - Block broadcast_0_piece0 stored as bytes in memory (estimated size 1256.0 B, free 110.0 MB)
2019-05-08T08:25:31.178025215Z 2019-05-08 08:25:31 INFO  BlockManagerInfo:54 - Added broadcast_0_piece0 in memory on demo1-1557303918993-driver-svc.default.svc:7079 (size: 1256.0 B, free: 110.0 MB)
2019-05-08T08:25:31.182000364Z 2019-05-08 08:25:31 INFO  SparkContext:54 - Created broadcast 0 from broadcast at DAGScheduler.scala:1161
2019-05-08T08:25:31.202640906Z 2019-05-08 08:25:31 INFO  DAGScheduler:54 - Submitting 2 missing tasks from ResultStage 0 (MapPartitionsRDD[1] at map at SparkPi.scala:34) (first 15 tasks are for partitions Vector(0, 1))
2019-05-08T08:25:31.203502967Z 2019-05-08 08:25:31 INFO  TaskSchedulerImpl:54 - Adding task set 0.0 with 2 tasks
2019-05-08T08:25:31.245126257Z 2019-05-08 08:25:31 INFO  TaskSetManager:54 - Starting task 0.0 in stage 0.0 (TID 0, 172.20.90.161, executor 1, partition 0, PROCESS_LOCAL, 7878 bytes)
2019-05-08T08:25:31.805815672Z 2019-05-08 08:25:31 INFO  BlockManagerInfo:54 - Added broadcast_0_piece0 in memory on 172.20.90.161:36718 (size: 1256.0 B, free: 110.0 MB)
2019-05-08T08:25:31.946492966Z 2019-05-08 08:25:31 INFO  TaskSetManager:54 - Starting task 1.0 in stage 0.0 (TID 1, 172.20.90.161, executor 1, partition 1, PROCESS_LOCAL, 7878 bytes)
2019-05-08T08:25:31.957903365Z 2019-05-08 08:25:31 INFO  TaskSetManager:54 - Finished task 0.0 in stage 0.0 (TID 0) in 727 ms on 172.20.90.161 (executor 1) (1/2)
2019-05-08T08:25:31.99308236Z 2019-05-08 08:25:31 INFO  TaskSetManager:54 - Finished task 1.0 in stage 0.0 (TID 1) in 47 ms on 172.20.90.161 (executor 1) (2/2)
2019-05-08T08:25:31.994764897Z 2019-05-08 08:25:31 INFO  TaskSchedulerImpl:54 - Removed TaskSet 0.0, whose tasks have all completed, from pool 
2019-05-08T08:25:31.995390219Z 2019-05-08 08:25:31 INFO  DAGScheduler:54 - ResultStage 0 (reduce at SparkPi.scala:38) finished in 0.998 s
2019-05-08T08:25:32.003622135Z 2019-05-08 08:25:32 INFO  DAGScheduler:54 - Job 0 finished: reduce at SparkPi.scala:38, took 1.094511 s
2019-05-08T08:25:32.005407995Z Pi is roughly 3.1436157180785904
2019-05-08T08:25:32.011499948Z 2019-05-08 08:25:32 INFO  AbstractConnector:318 - Stopped Spark@7e97551f{HTTP/1.1,[http/1.1]}{0.0.0.0:4040}
2019-05-08T08:25:32.014105609Z 2019-05-08 08:25:32 INFO  SparkUI:54 - Stopped Spark web UI at http://demo1-1557303918993-driver-svc.default.svc:4040
2019-05-08T08:25:32.01861939Z 2019-05-08 08:25:32 INFO  KubernetesClusterSchedulerBackend:54 - Shutting down all executors
2019-05-08T08:25:32.019973046Z 2019-05-08 08:25:32 INFO  KubernetesClusterSchedulerBackend$KubernetesDriverEndpoint:54 - Asking each executor to shut down
2019-05-08T08:25:32.025136562Z 2019-05-08 08:25:32 WARN  ExecutorPodsWatchSnapshotSource:87 - Kubernetes client has been closed (this is expected if the application is shutting down.)
2019-05-08T08:25:32.087137746Z 2019-05-08 08:25:32 INFO  MapOutputTrackerMasterEndpoint:54 - MapOutputTrackerMasterEndpoint stopped!
2019-05-08T08:25:32.097659039Z 2019-05-08 08:25:32 INFO  MemoryStore:54 - MemoryStore cleared
2019-05-08T08:25:32.098360561Z 2019-05-08 08:25:32 INFO  BlockManager:54 - BlockManager stopped
2019-05-08T08:25:32.104432515Z 2019-05-08 08:25:32 INFO  BlockManagerMaster:54 - BlockManagerMaster stopped
2019-05-08T08:25:32.10761075Z 2019-05-08 08:25:32 INFO  OutputCommitCoordinator$OutputCommitCoordinatorEndpoint:54 - OutputCommitCoordinator stopped!
2019-05-08T08:25:32.114734944Z 2019-05-08 08:25:32 INFO  SparkContext:54 - Successfully stopped SparkContext
2019-05-08T08:25:32.117170277Z 2019-05-08 08:25:32 INFO  ShutdownHookManager:54 - Shutdown hook called
2019-05-08T08:25:32.118273045Z 2019-05-08 08:25:32 INFO  ShutdownHookManager:54 - Deleting directory /tmp/spark-bdb4e416-5ab7-420c-905e-ef43c30fb187
2019-05-08T08:25:32.120019227Z 2019-05-08 08:25:32 INFO  ShutdownHookManager:54 - Deleting directory /var/data/spark-118b216d-2d39-4287-ad71-5b5d7c7195c9/spark-06dbab1f-13aa-474c-a1db-8845e14627bf
```

##### delete spark job 
```$xslt
arena delete --type=sparkjob demo 
```
You will found the spark job is deleted.
```$xslt
sparkapplication.sparkoperator.k8s.io "demo1" deleted
time="2019-05-08T17:27:06+08:00" level=info msg="The Job demo1 has been deleted successfully"
configmap "demo1-sparkjob" deleted
```

Congratulations! You've run the distributed spark job with `arena` successfully. 