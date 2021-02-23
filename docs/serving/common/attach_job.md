# Attach a serving Job

Sometimes, you may need to enter the containers of the serving job and execute some commands,like the 'kubectl exec' command of kubectl tool, Arena also can cover this situation.

!!! warning

    `arena attach` command is only valid when the serving job is available 

`arena attach` command is described as below:

```
$ arena serve attach -h
Attach a serving job and execute some commands

Usage:
  arena serve attach JOB [-i INSTANCE] [-c CONTAINER] [flags]

Flags:
  -c, --container string   Container name. If omitted, the first container in the instance will be chosen
  -h, --help               help for attach
  -i, --instance string    Job instance name
  -T, --type string        The serving type, the possible option is [tf(Tensorflow),trt(Tensorrt),custom(Custom),kf(KFServing)]. (optional)
  -v, --version string     Set the serving job version

Global Flags:
      --arena-namespace string   The namespace of arena system service, like tf-operator (default "arena-system")
      --config string            Path to a kube config. Only required if out-of-cluster
      --loglevel string          Set the logging level. One of: debug|info|warn|error (default "info")
  -n, --namespace string         the namespace of the job
      --pprof                    enable cpu profile
      --trace                    enable trace
```

1\. Make sure the serving job is available.

```
$ arena serve ls
NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ADDRESS      PORTS
fast-style-transfer  Custom  alpha    1        1          172.28.8.37  RESTFUL:32761->5000
```

As you see, the serving job fast-style-transfer is available(DESIRED == AVAILABLE) and get the serving job details.

```
$ arena serve get fast-style-transfer
Name:       fast-style-transfer
Namespace:  default
Type:       Custom
Version:    alpha
Desired:    1
Available:  1
Age:        17d
Address:    172.28.8.37
Port:       RESTFUL:32761->5000
GPUs:       1

Instances:
  NAME                                                       STATUS   AGE  READY  RESTARTS  GPUs  NODE
  ----                                                       ------   ---  -----  --------  ----  ----
  fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4  Running  2d   1/1    0         1     cn-beijing.192.168.8.3 
```

2\. Attach the serving job.

```
$  arena serve attach fast-style-transfer
Hello! Arena attach the container custom-serving of instance fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4
#
```
Then execute the command `ls` in container: 

```
# ls
README.md  docs.md	evaluate.pyc  floyd.yml		      images	src	  transform_video.py
app.py	   evaluate.py	examples      floyd_requirements.txt  setup.sh	style.py
```

!!! note

    * you can use option '-i' to specify the instance you want to attach
    * you can use option '-c' to specify the container you want to attach of instance

3\. If the container of serving job can not execute 'sh' command, but it can execute 'bash', you can attach the container of the serving job by using following command:

```
$ arena serve attach <JOB_NAME> bash
```

4\. If you don't need to attach the container and only need to execute one command in container, you can execute a command like:

```
$ arena serve attach <JOB_NAME> -- <COMMAND>
```

for example:

```
$ arena serve attach fast-style-transfer -- mkdir /tmpdir
Hello! Arena attach the container custom-serving of instance fast-style-transfer-alpha-custom-serving-856dbcdbcb-j2vv4
```
