#  为训练任务提供配置文件

在提交训练任务时，我们可以指定该训练任务所需的配置文件。

目前，该功能只支持如下两种任务:

* tfjob
* mpijob

## 1.用法

当提交训练任务时，通过 `--config-file <host_path_file>:<container_path_file>` 为训练任务指定配置文件，该选项有一些规则：

* 如果指定了 <host_path_file> 并且没有指定 <container_path_file>，我们认为 <container_path_file> 的值与 <host_path_file>相同
* <container_path_file> 必须是一个文件，并且是绝对路径
*  在一个提交命令中，可以多次指定 `--config-file` ，例如: "--config-file /tmp/test1.conf:/etc/config/test1.conf --config-file /tmp/test2.conf:/etc/config/test2.conf"


## 2.例子


首先，我们创建一个名为"test-config.json"的文件,它的路径为"/tmp/test-config.json"。我们希望把这个文件放到训练任务的实例中，并存放路径为"/etc/config/config.json"。这个文件内容如下：
```
# cat /tmp/test-config.json
{
    "key": "job-config"

}
```
接着，使用如下命令提交一个tfjob:
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
等一段时间,查看任务状态:
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
使用kubectl检测文件是否已经存放在了任务的实例中:
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

可以看到，文件已经在实例中存在了。
