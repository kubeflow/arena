This recipe suggest how to update the tensorflow serving after it has deployed.

1\. Deploy a tensorflow serving job follow the [submit a tensorflow serving job which uses gpus](gpu.md).

2\. Update the serving.

arena support update some config of tensorlfow serving after it has deployed.

```shell
$ arena serve update tensorflow --help
Update a tensorflow serving job and its associated instances

Usage:
  arena serve update tensorflow [flags]

Flags:
      --command string                  the command will inject to container's command.
      --cpu string                      the request cpu of each replica to run the serve.
  -e, --env stringArray                 the environment variables
      --gpumemory int                   the limit GPU memory of each replica to run the serve.
      --gpus int                        the limit GPU count of each replica to run the serve.
  -h, --help                            help for tensorflow
      --image string                    the docker image name of serving job
      --memory string                   the request memory of each replica to run the serve.
      --model-config-file string        corresponding with --model_config_file in tensorflow serving
      --model-name string               the model name for serving, ignored if --model-config-file flag is set
      --model-path string               the model path for serving in the container, ignored if --model-config-file flag is set, otherwise required
      --monitoring-config-file string   corresponding with --monitoring_config_file in tensorflow serving
      --name string                     the serving name
      --replicas int                    the replicas number of the serve job.
      --version string                  the serving version

Global Flags:
      --arena-namespace string   The namespace of arena system service, like tf-operator (default "arena-system")
      --config string            Path to a kube config. Only required if out-of-cluster
      --loglevel string          Set the logging level. One of: debug|info|warn|error (default "info")
  -n, --namespace string         the namespace of the job
      --pprof                    enable cpu profile
      --trace                    enable trace
```

for example, if you want to scale the replicas, you can use

```shell
$ arena serve update tensorflow --name=mymnist1 --replicas=2 
```

and if you want to update the model path, you can do like this command.

```shell
$ arean serve update tensorflow --name=mymnist1 --model-name=/tfmodel/new_mnist
```

After you execute the command, the tensorflow serving will do rolling update with the support of kubernetes deployment.

