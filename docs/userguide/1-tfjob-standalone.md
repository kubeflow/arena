
Here is an example how you can use `Arena` for the machine learning training. It will download the source code from git url.

1. the first step is to check the available resources

```
arena top node
NAME                                IPADDRESS      ROLE    GPU(Total)  GPU(Allocated)
i-j6c68vrtpvj708d9x6j0  192.168.1.116  master  0           0
i-j6c8ef8d9sqhsy950x7x  192.168.1.119  worker  1           0
i-j6c8ef8d9sqhsy950x7y  192.168.1.120  worker  1           0
i-j6c8ef8d9sqhsy950x7z  192.168.1.118  worker  1           0
i-j6ccue91mx9n2qav7qsm  192.168.1.115  master  0           0
i-j6ce09gzdig6cfcy1lwr  192.168.1.117  master  0           0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
0/3 (0%)
```

There are 3 available nodes with GPU for running training jobs.


2\. Now we can submit a training job with `arena`, it will download the source code from github

```
# arena submit tf \
             --name=tf-git \
             --gpus=1 \
             --image=tensorflow/tensorflow:1.5.0-devel-gpu \
             --syncMode=git \
             --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
             "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 10000"
NAME:   tf-git
LAST DEPLOYED: Mon Jul 23 07:51:52 2018
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1alpha2/TFJob
NAME          AGE
tf-git-tfjob  0s
```

> the source code will be downloaded and extracted to the directory `code/` of the working directory. The default working directory is `/root`, you can also specify by using `--workingDir`.

> If you are using the private git repo, you can use the following command:

```
# arena submit tf \
             --name=tf-git \
             --gpus=1 \
             --image=tensorflow/tensorflow:1.5.0-devel-gpu \
             --syncMode=git \
             --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
             --env=GIT_SYNC_USERNAME=yourname \
             --env=GIT_SYNC_PASSWORD=yourpwd \
             "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py"
```

Notice: `arena` is using [git-sync](https://github.com/kubernetes/git-sync/blob/master/cmd/git-sync/main.go) to sync up source code. You can set the environment varaibles defined in git-sync project.

3\. List all the jobs

```
# arena list
NAME    STATUS   TRAINER  AGE  NODE
tf-git  RUNNING  tfjob    0s   192.168.1.120
```

4\. Check the resource usage of the job

```
# arena top job
NAME    STATUS   TRAINER  AGE  NODE           GPU(Requests)  GPU(Allocated)
tf-git  RUNNING  TFJOB    17s  192.168.1.120  1              1


Total Allocated GPUs of Training Job:
1

Total Requested GPUs of Training Job:
1
```

5\. Check the resource usage of the cluster

```
# arena top node
NAME                    IPADDRESS      ROLE    GPU(Total)  GPU(Allocated)
i-j6c68vrtpvj708d9x6j0  192.168.1.116  master  0           0
i-j6c8ef8d9sqhsy950x7x  192.168.1.119  worker  1           0
i-j6c8ef8d9sqhsy950x7y  192.168.1.120  worker  1           1
i-j6c8ef8d9sqhsy950x7z  192.168.1.118  worker  1           0
i-j6ccue91mx9n2qav7qsm  192.168.1.115  master  0           0
i-j6ce09gzdig6cfcy1lwr  192.168.1.117  master  0           0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
1/3 (33%)
```


6\. Get the details of the specific job

```
# arena get tf-git
NAME    STATUS   TRAINER  AGE  INSTANCE               NODE
tf-git  RUNNING  TFJOB    5s   tf-git-tfjob-worker-0  192.168.1.120
```

7\. Check logs

```
# arena logs tf-git
2018-07-22T23:56:20.841129509Z WARNING:tensorflow:From code/tensorflow-sample-code/tfjob/docker/mnist/main.py:119: softmax_cross_entropy_with_logits (from tensorflow.python.ops.nn_ops) is deprecated and will be removed in a future version.
2018-07-22T23:56:20.841211064Z Instructions for updating:
2018-07-22T23:56:20.841217002Z
2018-07-22T23:56:20.841221287Z Future major versions of TensorFlow will allow gradients to flow
2018-07-22T23:56:20.841225581Z into the labels input on backprop by default.
2018-07-22T23:56:20.841229492Z
...
2018-07-22T23:57:11.842929868Z Accuracy at step 920: 0.967
2018-07-22T23:57:11.842933859Z Accuracy at step 930: 0.9646
2018-07-22T23:57:11.842937832Z Accuracy at step 940: 0.967
2018-07-22T23:57:11.842941362Z Accuracy at step 950: 0.9674
2018-07-22T23:57:11.842945487Z Accuracy at step 960: 0.9693
2018-07-22T23:57:11.842949067Z Accuracy at step 970: 0.9687
2018-07-22T23:57:11.842952818Z Accuracy at step 980: 0.9688
2018-07-22T23:57:11.842956775Z Accuracy at step 990: 0.9649
2018-07-22T23:57:11.842961076Z Adding run metadata for 999
```

8\. More information about the training job in the logviewer

```
# arena logviewer tf-git
Your LogViewer will be available on:
192.168.1.120:8080/tfjobs/ui/#/default/tf-git-tfjob
```

![](1-tfjob-logviewer.jpg)


Congratulations! You've run the first training job with `arena` successfully. 