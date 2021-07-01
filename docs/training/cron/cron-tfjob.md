# Submit a cron tensorflow job

## Submit the cron tfjob

Here is an example how you can use ``Arena`` for the machine learning training. It will download the source code from git url.

1\. the first step is to check the available resources:

    $ arena top node
    
    NAME                       IPADDRESS      ROLE    STATUS  GPU(Total)  GPU(Allocated)
    cn-hongkong.192.168.2.107  47.242.51.160  <none>  Ready   0           0
    cn-hongkong.192.168.2.108  192.168.2.108  <none>  Ready   1           0
    cn-hongkong.192.168.2.109  192.168.2.109  <none>  Ready   1           0
    cn-hongkong.192.168.2.110  192.168.2.110  <none>  Ready   1           0
    ------------------------------------------------------------------------------------
    Allocated/Total GPUs In Cluster:
    0/3 (0.0%)

There are 3 available nodes with GPU for running training jobs.

2\. Now we can submit a training job with ``arena``, it will download the source code from github:

    $ arena cron \
      tfjob \
      --schedule="0 0 22 * * ?" \
      --concurrency-policy="Allow" \
      --deadline="2021-10-01T13:00:12Z" \
      --history-limit=10 \
      --gpus=1 \
      --name=cron-tfjob \
      --env=TEST_TMPDIR=code/tensorflow-sample-code/ \
      --sync-mode=git \
      --sync-source=https://github.com/happy2048/tensorflow-sample-code.git \
      --logdir=/training_logs \
      --image="registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu" \
      "'python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 5000'"
    
    configmap/cron-tfjob-tfjob created
    configmap/cron-tfjob-tfjob labeled
    cron.apps.kubedl.io/cron-tfjob created
    INFO[0003] The cron tfjob cron-tfjob has been submitted successfully
    INFO[0003] You can run `arena cron get cron-tfjob` to check the cron status

!!! note

    * if you can't pull the image "registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu", please replace it with "registry.cn-hongkong.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu"

descriptions:

* the ``--schedule``  specifies the schedule of tfjob, see https://en.wikipedia.org/wiki/Cron.
* the ``--concurrency-policy `` specifies how to treat concurrent executions of a tfjob, valid values are:

    * "Allow" (default): allows CronJobs to run concurrently;
    * "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
    * "Replace": cancels currently running job and replaces it with a new one
* the ``--deadline`` is optional, which specifies  the timestamp that a cron job can keep scheduling util then
* the ``--history-limit`` is optional which specifies  the number of finished job history to retain.



## List all cron tfjobs

You can use ``arena cron list -A`` to list all tensorflow jobs:

    $ arena cron list -A
    NAMESPACE  NAME         TYPE   SCHEDULE      SUSPEND  DEADLINE              CONCURRENCYPOLICY
    default    cron-tfjob   TFJob  0 0 22 * * ?  false    2021-10-01T21:00:12Z  Allow



## Get the cron tfjob detail information

When the cron tfjob is submit,  you can use ``arena cron get`` to get the cron tfjob detail information.

```shell
$ arena cron get cron-tfjob
Name:               cron-tfjob
Namespace:          default
Type:               TFJob
Schedule:           0 0 22 * * ?
Suspend:            false
ConcurrencyPolicy:  Allow
CreationTimestamp:  2021-06-25T10:44:17Z
LastScheduleTime:
Deadline:           2021-10-01T21:00:12Z

History:
NAME  STATUS  TYPE  CREATETIME  FINISHTIME
----  ------  ----  ----------  ----------
```



## Suspend the cron tfjob

When you want to stop the cron tfjob, you can use ``arena cron suspend`` to suspend the cron tfjob schedule.


```shell
$ arena cron suspend cron-tfjob
cron cron-tfjob suspend success
```



## Resume the cron tfjob

When you want to resume the stopped cron tfjob, you can use ``arena cron resume`` to do it.

```shell
$ arena cron resume cron-tfjob
cron cron-tfjob resume success
```




## Delete the cron tfjob

When the job is completed, use ``arena cron delete`` to delete the job:

    $ arena cron delete cron-tfjob
    cron cron-tfjob has deleted
    configmap "cron-tfjob-tfjob" deleted

Congratulations! You've run the first training job with ``arena`` successfully. 

