# Submit a standalone tensorflow job

## Submit the job

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

    $ arena \
      submit \
      tfjob \
      --gpus=1 \
      --name=tf-standalone-test-with-git \
      --env=TEST_TMPDIR=code/tensorflow-sample-code/ \
      --sync-mode=git \
      --sync-source=https://github.com/happy2048/tensorflow-sample-code.git \
      --logdir=/training_logs \
      --image="registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu" \
      "'python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 5000'"

    configmap/tf-git-tfjob created
    configmap/tf-git-tfjob labeled
    tfjob.kubeflow.org/tf-git created
    INFO[0000] The Job tf-git has been submitted successfully
    INFO[0000] You can run `arena get tf-git --type tfjob` to check the job status

!!! note

    * if you can't pull the image "registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu", please replace it with "registry.cn-hongkong.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu"

descriptions:

* the source code will be downloaded and extracted to the directory ``code/`` of the working directory. The default working directory is ``/root``, you can also specify by using ``--workingDir``. Also, you may specify the branch you are pulling code from by addding ``--env GIT_SYNC_BRANCH=main`` to the paramasters while submitting the job.
* If you are using the private git repo, you can use the following command:

```

$ arena submit tf \
--name=tf-git \
--gpus=1 \
--image=tensorflow/tensorflow:1.5.0-devel-gpu \
--syncMode=git \
--syncSource=https://github.com/happy2048/tensorflow-sample-code.git \
--env=GIT_SYNC_USERNAME=yourname \
--env=GIT_SYNC_PASSWORD=yourpwd \
"python code/tensorflow-sample-code/tfjob/docker/mnist/main.py"

```

!!! note

    ``arena`` is using [git-sync](https://github.com/kubernetes/git-sync/blob/master/cmd/git-sync/main.go) to sync up source code. You can set the environment variables defined in git-sync project.


## List all tensorflow jobs

You can use ``arena list --type tfjob`` to list all tensorflow jobs:

    $ arena list --type tfjob
    NAME                         STATUS   TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
    tf-standalone-test-with-git  PENDING  TFJOB    3s        0               0               N/A

If you want to list all training jobs,you can use ``arena list``:

    $ arena list
    NAME                         STATUS     TRAINER     DURATION  GPU(Requested)  GPU(Allocated)  NODE
    tf-standalone-test-with-git  PENDING    TFJOB       5m        0               0               N/A
    pytorch-test                 FAILED     PYTORCHJOB  10m       1               N/A             192.168.1.101
    mpi-dist                     SUCCEEDED  MPIJOB      1m        0               N/A             192.168.1.100


## Get the tensorflow job detail information

If you want to get the job details, you can use ``arena get``:

    $ arena get tf-standalone-test-with-git
    Name:        tf-standalone-test-with-git
    Status:      PENDING
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    7m

    Instances:
    NAME                                 STATUS    AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                                 ------    ---  --------  --------------  ----
    tf-standalone-test-with-git-chief-0  Init:0/1  7m   true      0               N/A


## Get the tensorflow job logs

When the job status is running, use ``arena logs`` to get the job logs:

    $ arena logs tf-standalone-test-with-git --tail 10
    Accuracy at step 4920: 0.9828
    Accuracy at step 4930: 0.9823
    Accuracy at step 4940: 0.9827
    Accuracy at step 4950: 0.9824
    Accuracy at step 4960: 0.983
    Accuracy at step 4970: 0.979
    Accuracy at step 4980: 0.9821
    Accuracy at step 4990: 0.9823
    Adding run metadata for 4999
    Total Train-accuracy=0.9823

In this case,we only display the last 10 lines of the logs.


## Get the logviewer

More information about the training job in the logviewer:

    $ arena logviewer tf-standalone-test-with-git
    Your LogViewer will be available on:
    172.20.0.197:8080/tfjobs/ui/#/default/tf-standalone-test-with-git

![logviewer](1-tfjob-logviewer.jpg)


## Delete the job

When the job is completed, use ``arena delete`` to delete the job:

    $ arena delete tf-standalone-test-with-git 

Congratulations! You've run the first training job with ``arena`` successfully. 
