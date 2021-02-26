# Attach a Running Training Job

Sometimes, you may need to enter the containers of the training job and execute some commands,like the 'kubectl exec' command of kubectl tool, Arena also can cover this situation.

!!! warning

    `arena attach` command is only valid when the training job status is running 

`arena attach` command is described as below:

```
$ arena attach -h
Attach a training job and execute some commands

Usage:
  arena attach JOB [-i INSTANCE] [-c CONTAINER] [flags]

Flags:
  -c, --container string   Container name. If omitted, the first container in the instance will be chosen
  -h, --help               help for attach
  -i, --instance string    Job instance name
  -T, --type string        The training type to get, the possible option is tf(Tensorflow),mpi(MPI),py(Pytorch),horovod(Horovod),volcano(Volcano),et(ElasticTraining),spark(Spark). (optional)

Global Flags:
      --arena-namespace string   The namespace of arena system service, like tf-operator (default "arena-system")
      --config string            Path to a kube config. Only required if out-of-cluster
      --loglevel string          Set the logging level. One of: debug|info|warn|error (default "info")
  -n, --namespace string         the namespace of the job
      --pprof                    enable cpu profile
      --trace                    enable trace
```

1\. Make sure the training job is running.

```
$ arena list
NAME                           STATUS     TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
dawnbench-1x1-v6               RUNNING    MPIJOB   15m       2               2               192.168.1.137
tf-distributed-test            FAILED     TFJOB    11m       0               N/A             N/A
tf-git                         SUCCEEDED  TFJOB    14m       0               N/A             N/A
mpi-test                       SUCCEEDED  MPIJOB   12h       0               N/A             N/A
elastic-training               SUCCEEDED  ETJOB    40d       0               N/A             N/A
horovod-resnet50-v2-4x8-fluid  SUCCEEDED  MPIJOB   1h        0               N/A             N/A
horovod-resnet50-v2-4x8-nfs    SUCCEEDED  MPIJOB   2h        0               N/A             N/A
```

As you see, the training job dawnbench-1x1-v6 is running and get the training job details.

```
$ arena get dawnbench-1x1-v6
Name:      dawnbench-1x1-v6
Status:    RUNNING
Namespace: default
Priority:  N/A
Trainer:   MPIJOB
Duration:  18m

Instances:
  NAME                             STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
  ----                             ------   ---  --------  --------------  ----
  dawnbench-1x1-v6-launcher-7hshj  Running  18m  true      0               cn-beijing.192.168.1.137
  dawnbench-1x1-v6-worker-0        Running  18m  false     1               cn-beijing.192.168.1.137
  dawnbench-1x1-v6-worker-1        Running  18m  false     1               cn-beijing.192.168.1.138 
```

2\. Attach the training job.

```
$ arena attach dawnbench-1x1-v6 
Hello! Arena attach the container mpi of instance dawnbench-1x1-v6-launcher-7hshj
#
```
Then execute the command `ls` in container: 

```
# ls
README.md	 cmd.sh			 launch-example.sh  perseus-tf-vm-demo.ipynb   start.sh
benchmarks	 config-fp16-tf.sh.orig  login.sh	    run_dist_example.sh
clean_caches.sh  hurun_dist_example.sh	 perseus-tf-env.sh  run_local_2gpu_example.sh
```

!!! note

    * you can use option '-i' to specify the instance you want to attach
    * you can use option '-c' to specify the container you want to attach of instance

3\. If the container of training job can not execute 'sh' command, but it can execute 'bash', you can attach the container of the training job by using following command:

```
$ arena attach <JOB_NAME> bash
```

4\. If you don't need to attach the container and only need to execute one command in container, you can execute a command like:

```
$ arena attach <JOB_NAME> -- <COMMAND>
```

for example:

```
$ arena attach dawnbench-1x1-v6 -- mkdir /tmpdir
Hello! Arena attach the container mpi of instance dawnbench-1x1-v6-launcher-7hshj
```
