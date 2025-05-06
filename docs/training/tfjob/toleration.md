# Submit Tensorflow Job With Specified Node Tolerations

Arena supports submitting a job and the job tolerates k8s taints nodes(Currently only support mpi job and tf job), the following steps can help you how to use this feature.

1\. query k8s cluster information.

```
$ kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.3.225   Ready    master   2d23h   v1.12.6-aliyun.1
cn-beijing.192.168.3.226   Ready    master   2d23h   v1.12.6-aliyun.1
cn-beijing.192.168.3.227   Ready    master   2d23h   v1.12.6-aliyun.1
cn-beijing.192.168.3.228   Ready    <none>   2d22h   v1.12.6-aliyun.1
cn-beijing.192.168.3.229   Ready    <none>   2d22h   v1.12.6-aliyun.1
cn-beijing.192.168.3.230   Ready    <none>   2d22h   v1.12.6-aliyun.1
```

2\. give some taints for k8s nodes,for example: give taint `gpu_node=invalid:NoSchedule` to node `cn-beijing.192.168.3.228` and node `cn-beijing.192.168.3.229`,give taint  `ssd_node=invalid:NoSchedule` to node `cn-beijing.192.168.3.230`,now all k8s pods can't schedule to these nodes.
```
$ kubectl taint nodes cn-beijing.192.168.3.228 gpu_node=invalid:NoSchedule                                                                            
node/cn-beijing.192.168.3.228 tainted
$ kubectl taint nodes cn-beijing.192.168.3.229 gpu_node=invalid:NoSchedule                                                                            
node/cn-beijing.192.168.3.229 tainted
$ kubectl taint nodes cn-beijing.192.168.3.230 ssd_node=invalid:NoSchedule                                                                            
node/cn-beijing.192.168.3.230 tainted
```

3\. when submitting a job with option ``--toleration``, you can tolerate some nodes which exists taints.
```
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
```

4\. query the job details.
```
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
tf-standalone-test-with-git-chief-0  Running   7m   true      0               192.168.3.230
```

the instances of job are running  on node cn-beijing.192.168.3.230(ip is 192.168.3.230,taint is ssd_node=invalid).

!!! note

    * you can use ``--toleration`` multiple times,for example: you can use  "--toleration gpu_node --toleration ssd_node" when submitting a job,it represents that the job tolerates nodes which own taint "gpu_node=invalid" and taint "ssd_node=invalid".

    * you can use "--toleration all" to tolerate all node taints.
