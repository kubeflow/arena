
这个示例展示了如何使用 `Arena` 进行机器学习模型训练。该示例将从 git url 下载源代码。

1. 第一步是检查可用的GPU资源

```
arena top node
NAME IPADDRESS ROLE GPU(Total) GPU(Allocated)
i-j6c68vrtpvj708d9x6j0 192.168.1.116 master 0 0
i-j6c8ef8d9sqhsy950x7x 192.168.1.119 worker 1 0
i-j6c8ef8d9sqhsy950x7y 192.168.1.120 worker 1 0
i-j6c8ef8d9sqhsy950x7z 192.168.1.118 worker 1 0
i-j6ccue91mx9n2qav7qsm 192.168.1.115 master 0 0
i-j6ce09gzdig6cfcy1lwr 192.168.1.117 master 0 0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
0/3 (0%)
```

有 3 个包含 GPU 的可用节点用于运行训练作业。


2\.现在，我们可以通过 `arena` 提交一个训练作业，本示例从 github 下载源代码

```
#arena submit tf \
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
NAME AGE
tf-git-tfjob 0s
```

> 这会下载源代码，并将其解压缩到工作目录的 `code/` 目录。默认的工作目录是 `/root`，您也可以使用 `--workingDir` 加以指定。

> 如果您正在使用非公开 git 代码库，则可以使用以下命令：

```
#arena submit tf \
             --name=tf-git \
             --gpus=1 \
             --image=tensorflow/tensorflow:1.5.0-devel-gpu \
             --syncMode=git \
             --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
             --env=GIT_SYNC_USERNAME=yourname \
             --env=GIT_SYNC_PASSWORD=yourpwd \
             "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py"
```

注意：`arena` 使用 [git-sync](https://github.com/kubernetes/git-sync/blob/master/cmd/git-sync/main.go) 来同步源代码。您可以设置在 git-sync 项目中定义的环境变量。

3\.列出所有作业

```
#arena list
NAME STATUS TRAINER AGE NODE
tf-git RUNNING tfjob 0s 192.168.1.120
```

4\.检查作业所使用的GPU资源

```
#arena top job
NAME STATUS TRAINER AGE NODE GPU(Requests) GPU(Allocated)
tf-git RUNNING TFJOB 17s 192.168.1.120 1 1


Total Allocated GPUs of Training Job:
1

Total Requested GPUs of Training Job:
1
```

5\.检查集群所使用的GPU资源

```
#arena top node
NAME IPADDRESS ROLE GPU(Total) GPU(Allocated)
i-j6c68vrtpvj708d9x6j0 192.168.1.116 master 0 0
i-j6c8ef8d9sqhsy950x7x 192.168.1.119 worker 1 0
i-j6c8ef8d9sqhsy950x7y 192.168.1.120 worker 1 1
i-j6c8ef8d9sqhsy950x7z 192.168.1.118 worker 1 0
i-j6ccue91mx9n2qav7qsm 192.168.1.115 master 0 0
i-j6ce09gzdig6cfcy1lwr 192.168.1.117 master 0 0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
1/3 (33%)
```


6\.获取特定作业的详细信息

```
#arena get tf-git
NAME STATUS TRAINER AGE INSTANCE NODE
tf-git RUNNING TFJOB 5s tf-git-tfjob-worker-0 192.168.1.120
```

7\.检查日志

```
#arena logs tf-git
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

8\.日志查看器中有关训练作业的更多信息

```
#arena logviewer tf-git
Your LogViewer will be available on:
192.168.1.120:8080/tfjobs/ui/#/default/tf-git-tfjob
```

![](1-tfjob-logviewer.jpg)


恭喜！您已经成功使用 `arena` 完成了第一项训练作业。 
