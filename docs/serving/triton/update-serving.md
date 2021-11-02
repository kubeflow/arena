This recipe suggest how to update the triton serving after it has deployed.

1\. Deploy a triton serving job follow the [submit a nvidia triton serving job which use gpus](triton/serving.md).

2\. Update the serving.

arena support update some config of triton serving after it has deployed.

```shell
$ arena serve update triton --help
Update a triton serving job and its associated instances

Usage:
  arena serve update triton [flags]

Flags:
      --allow-metrics             open metrics (default true)
      --command string            the command will inject to container's command.
      --cpu string                the request cpu of each replica to run the serve.
  -e, --env stringArray           the environment variables
      --gpumemory int             the limit GPU memory of each replica to run the serve.
      --gpus int                  the limit GPU count of each replica to run the serve.
  -h, --help                      help for triton
      --image string              the docker image name of serving job
      --memory string             the request memory of each replica to run the serve.
      --model-repository string   the path of triton model path
      --name string               the serving name
      --replicas int              the replicas number of the serve job.
      --version string            the serving version

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
$ arena serve update triton --name=test-triton --replicas=2 
```

and if you want to update the model path, you can do like this command.

```shell
$ arean serve update triton --name=test-triton --model-repository=/mnt/models/ai/triton/model_repository
```

After you execute the command, the triton serving will do rolling update with the support of kubernetes deployment.


