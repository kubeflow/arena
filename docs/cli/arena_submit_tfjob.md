## arena submit tfjob

Submit a TFJob as training job.

### Synopsis

Submit a TFJob as training job.

```
arena submit tfjob [flags]
```

### Options

```
  -a, --annotation strings           the annotations
      --chief                        enable chief, which is required for estimator.
      --chief-cpu string             the cpu resource to use for the Chief, like 1 for 1 core.
      --chief-cpu-limit string       the cpu resource limit to use for the Chief, like 1 for 1 core.      
      --chief-memory string          the memory resource to use for the Chief, like 1Gi.
      --chief-memory-limit string    the memory resource limit to use for the Chief, like 1Gi.      
      --chief-port int               the port of the chief.
      --chief-selector strings       assigning jobs with "Chief" role to some k8s particular nodes(this option would cover --selector), usage: "--chief-selector=key=value"
      --clean-task-policy string     How to clean tasks after Training is done, only support Running, None. (default "Running")
      --config-file strings          giving configuration files when submiting jobs,usage:"--config-file <host_path_file>:<container_path_file>"
  -d, --data strings                 specify the datasource to mount to the job, like <name_of_datasource>:<mount_point_on_job>
      --data-dir strings             the data dir. If you specify /data, it means mounting hostpath /data into container path /data
  -e, --env strings                  the environment variables
      --evaluator                    enable evaluator, which is optional for estimator.
      --evaluator-cpu string         the cpu resource to use for the evaluator, like 1 for 1 core.
      --evaluator-cpu-limit string   the cpu resource limit to use for the evaluator, like 1 for 1 core.      
      --evaluator-memory string      the memory resource to use for the evaluator, like 1Gi.
      --evaluator-memory-limit string the memory resource limit to use for the evaluator, like 1Gi.
      --evaluator-selector strings   assigning jobs with "Evaluator" role to some k8s particular nodes(this option would cover --selector), usage: "--evaluator-selector=key=value"
      --gang                         enable gang scheduling
      --gpus int                     the GPU count of each worker to run the training.
  -h, --help                         help for tfjob
      --image string                 the docker image name of training job
      --image-pull-secret strings    giving names of imagePullSecret when you want to use a private registry, usage:"--image-pull-secret <name1>"
      --logdir string                the training logs dir, default is /training_logs (default "/training_logs")
      --name string                  override name
  -p, --priority string              priority class name
      --ps int                       the number of the parameter servers.
      --ps-cpu string                the cpu resource to use for the parameter servers, like 1 for 1 core.
      --ps-cpu-limit string          the cpu-limit resource to use for the parameter servers, like 1 for 1 core.
      --ps-gpus int                  the gpu resource to use for the parameter servers, like 1 for 1 gpu.
      --ps-image string              the docker image for tensorflow workers
      --ps-memory string             the memory resource to use for the parameter servers, like 1Gi.
      --ps-memory-limit string       the memory limit resource to use for the parameter servers, like 1Gi.
      --ps-port int                  the port of the parameter server.
      --ps-selector strings          assigning jobs with "PS" role to some k8s particular nodes(this option would cover --selector), usage: "--ps-selector=key=value"
      --rdma                         enable RDMA
      --retry int                    retry times.
      --role-sequence string         specify the tfjob role sequence,like: "Worker,PS,Chief,Evaluator" or "w,p,c,e"
      --running-timeout duration     Specifies the duration since startTime during which the job can remain active before it is terminated(e.g. '5s', '1m', '2h22m').
      --selector strings             assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" 
      --sync-image string            the docker image of syncImage
      --sync-mode string             syncMode: support rsync, hdfs, git
      --sync-source string           sync-source: for rsync, it's like 10.88.29.56::backup/data/logoRecoTrain.zip; for git, it's like https://github.com/kubeflow/tf-operator.git
      --tensorboard                  enable tensorboard
      --tensorboard-image string     the docker image for tensorboard (default "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.12.0-devel")
      --toleration strings           tolerate some k8s nodes with taints,usage: "--toleration taint-key" or "--toleration all" 
      --ttl-after-finished duration  Defines the TTL for cleaning up finished TFJobs(e.g. '5s', '1m', '2h22m'). Defaults to infinite.
      --worker-cpu string            the cpu resource to use for the worker, like 1 for 1 core.
      --worker-cpu-limit string      the cpu resource limit to use for the worker, like 1 for 1 core.      
      --worker-image string          the docker image for tensorflow workers
      --worker-memory string         the memory resource to use for the worker, like 1Gi.
      --worker-memory-limit string   the memory resource limit to use for the worker, like 1Gi.
      --worker-port int              the port of the worker.
      --worker-selector strings      assigning jobs with "Worker" role to some k8s particular nodes(this option would cover --selector), usage: "--worker-selector=key=value"
      --workers int                  the worker number to run the distributed training. (default 1)
      --working-dir string           working directory to extract the code. If using syncMode, the $workingDir/code contains the code (default "/root")
```

### Options inherited from parent commands

```
      --arena-namespace string   The namespace of arena system service, like tf-operator (default "arena-system")
      --config string            Path to a kube config. Only required if out-of-cluster
      --loglevel string          Set the logging level. One of: debug|info|warn|error (default "info")
  -n, --namespace string         the namespace of the job
      --pprof                    enable cpu profile
      --trace                    enable trace
```

### SEE ALSO

* [arena submit](arena_submit.md)	 - Submit a training job.

###### Auto generated by spf13/cobra on 5-Mar-2021
