# Assign configuration files for jobs

you can pass the configuration files to containers when submiting jobs.

this feature only support follow jobs:

* tfjob
* mpijob

and requirements are:

* helm version >= 2.14.1 and not support helm v3
  
## 1.usage

you can use `--config-file <host_path_file>:<container_path_file>` to assign a configuration file to container.and there is some rules:

* if assignd <host_path_file> and not assign <container_path_file>,we see <container_path_file> is the same as <host_path_file>
* <container_path_file> must be a file with absolute path
*  you can use `--config-file` more than one in a command,eg: "--config-file /tmp/test1.conf:/etc/config/test1.conf --config-file /tmp/test2.conf:/etc/config/test2.conf"


## 2.sample


firstly,we create a test file which name is "test-config.json",its' path is "/tmp/test-config.json". we want push this file to containers of a tfjob (or mpijob) and the path in container is "/etc/config/config.json". 
```
# cat /tmp/test-config.json
{
    "key": "job-config"

}
```
secondly,use follow command to create tfjob:
```
# arena submit tfjob \
    --name=tf \
    --gpus=1              \
    --workers=1              \
    --workerImage=cheyang/tf-mnist-distributed:gpu \
    --psImage=cheyang/tf-mnist-distributed:cpu \
    --ps=1              \
    --tensorboard \
    --config-file /tmp/test-config.json:/etc/config/config.json \
      "python /app/main.py"
```
wait a minute,get the job status:
```
# arena get tf
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 16s

NAME  STATUS   TRAINER  AGE  INSTANCE     NODE
tf    RUNNING  TFJOB    16s  tf-ps-0      192.168.7.18
tf    RUNNING  TFJOB    16s  tf-worker-0  192.168.7.16

Your tensorboard will be available on:
http://192.168.7.10:31825
```
use kubectl to check file is in container or not:
```
# kubectl exec -ti tf-ps-0 -- cat /etc/config/config.json
{
    "key": "job-config"

}
# kubectl exec -ti tf-worker-0 -- cat /etc/config/config.json
{
    "key": "job-config"

}

```

as you see,the file is in the containers.
