## arena serve distributed

Submit distributed server job to deploy and serve machine learning models.

### Synopsis

Submit distributed server job to deploy and serve machine learning models.

```
  arena serve distributed [flags]
```

### Options

```
  -a, --annotation stringArray                      specify the annotations, usage: "--annotation=key=value" or "--annotation key=value"
      --command string                              specify the container command
      --config-file stringArray                     giving configuration files when serving model, usage:"--config-file <host_path_file>:<container_path_file>"
  -d, --data stringArray                            specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>
      --data-dir stringArray                        specify the trained models datasource on host to mount for serving, like <host_path>:<mount_point_on_job>
      --data-subpath-expr stringArray               specify the datasource subpath to mount to the job by expression, like <name_of_datasource>:<mount_subpath_expr>
      --device stringArray                          the chip vendors and count that used for resources, such as amd.com/gpu=1 gpu.intel.com/i915=1.
      --enable-istio                                enable Istio for serving or not (disable Istio by default)
  -e, --env stringArray                             the environment variables, usage: "--env envName=envValue"
      --env-from-secret stringArray                 the environment variables using Secret data, usage: "--env-from-secret envName=secretName"
      --expose-service                              expose service using Istio gateway for external access or not (not expose by default)
  -h, --help                                        help for distributed
      --image string                                the docker image name of serving job
      --image-pull-policy string                    the policy to pull the image, and the default policy is IfNotPresent (default "IfNotPresent")
      --image-pull-secret stringArray               giving names of imagePullSecret when you want to use a private registry, usage:"--image-pull-secret <name1>"
      --init-backend string                         specity the init backend for distributed serving job. Currently only support ray. support: ray
  -l, --label stringArray                           specify the labels
      --liveness-probe-action string                the liveness probe action, support httpGet,exec,grpc,tcpSocket
      --liveness-probe-action-option stringArray    the liveness probe action option, usage: --liveness-probe-action-option="path: /healthz" or --liveness-probe-action-option="command=cat /tmp/healthy"
      --liveness-probe-option stringArray           the liveness probe option, usage: --liveness-probe-option="initialDelaySeconds: 3" or --liveness-probe-option="periodSeconds: 3"
      --master-command string                       the command to run for the master pod
      --master-cpu string                           the cpu resource to use for the master pod, like 1 for 1 core
      --master-gpucore int                          the limit GPU core of master pod to run the serve
      --master-gpumemory int                        the limit GPU memory of master pod to run the serve
      --master-gpus int                             the gpu resource to use for the master pod, like 1 for 1 gpu
      --master-memory string                        the memory resource to use for the master pod, like 1Gi
      --masters int                                 the number of the master pods (p.s. only support 1 master currently) (default 1)
      --max-surge string                            the maximum number of pods that can be created over the desired number of pods
      --max-unavailable string                      the maximum number of Pods that can be unavailable during the update process
      --metrics-port int                            the port of metrics, default is 0 represents that don't create service listening on this port
      --model-name string                           model name
      --model-version string                        model version
      --name string                                 the serving name
      --port int                                    the port of gRPC listening port, default is 0 represents that don't create service listening on this port
      --readiness-probe-action string               the readiness probe action, support httpGet,exec,grpc,tcpSocket
      --readiness-probe-action-option stringArray   the readiness probe action option, usage: --readiness-probe-action-option="path: /healthz" or --readiness-probe-action-option="command=cat /tmp/healthy"
      --readiness-probe-option stringArray          the readiness probe option, usage: --readiness-probe-option="initialDelaySeconds: 3" or --readiness-probe-option="periodSeconds: 3"
      --replicas int                                the replicas number of the serve job. (default 1)
      --restful-port int                            the port of RESTful listening port, default is 0 represents that don't create service listening on this port
      --selector stringArray                        assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" 
      --share-memory string                         the request share memory of each replica to run the serve.
      --shell string                                specify the linux shell, usage: bash or sh (default "sh")
      --startup-probe-action string                 the startup probe action, support httpGet,exec,grpc,tcpSocket
      --startup-probe-action-option stringArray     the startup probe action option, usage: --startup-probe-action-option="path: /healthz" or --startup-probe-action-option="command=cat /tmp/healthy"
      --startup-probe-option stringArray            the startup probe option, usage: --startup-probe-option="initialDelaySeconds: 3" or --startup-probe-option="periodSeconds: 3"
      --temp-dir stringArray                        specify the deployment empty dir, like <empty_dir_name>:<mount_point_on_pod>
      --temp-dir-subpath-expr stringArray           specify the datasource subpath to mount to the pod by expression, like <empty_dir_name>:<mount_subpath_expr>
      --toleration stringArray                      tolerate some k8s nodes with taints,usage: "--toleration key=value:effect,operator" or "--toleration all" 
      --version string                              the serving version
      --worker-command string                       the command to run of each worker pods
      --worker-cpu string                           the cpu resource to use for each worker pods, like 1 for 1 core
      --worker-gpucore int                          the limit GPU core of each worker pods to run the serve
      --worker-gpumemory int                        the limit GPU memory of each worker pods to run the serve
      --worker-gpus int                             the gpu resource to use for each worker pods, like 1 for 1 gpu
      --worker-memory string                        the memory resource to use for the worker pods, like 1Gi
      --workers int                                 the number of the worker pods
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

* [arena serve](arena_serve.md)	 - Serve a job.