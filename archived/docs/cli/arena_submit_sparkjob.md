## arena submit sparkjob 

Submit SparkJob as training job.

### Synopsis

Submit SparkJob as training job.

```
arena submit tfjob [flags]
```

### Options

```
    --image string        the docker image name of training job
    --jar string          jar path in image
    --main-class string   main class of your jar
    --name string         override name
    --workers int         the worker number to run the distributed training. (default 1)
```

### Options inherited from parent commands

```
      --arenaNamespace string   The namespace of arena system service, like TFJob (default "arena-system")
      --config string           Path to a kube config. Only required if out-of-cluster
      --loglevel string         Set the logging level. One of: debug|info|warn|error (default "info")
      --namespace string        the namespace of the job (default "default")
      --pprof                   enable cpu profile
      --trace                   enable trace
```

### SEE ALSO

* [arena submit](arena_submit.md)	 - Submit a job.

