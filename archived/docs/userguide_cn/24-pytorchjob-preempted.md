## pytorch 任务支持抢占

1. 利用下列 yaml 创建 PriorityClass 对象，这里定义了两个优先级 `critical` 和 `medium`:
	```shell
	# critical 和 medium 声明
	➜ cat priorityClass.yaml
	apiVersion: scheduling.k8s.io/v1beta1
	description: Used for the critical app
	kind: PriorityClass
	metadata:
	  name: critical
	value: 1100000

	---

	apiVersion: scheduling.k8s.io/v1beta1
	description: Used for the medium app
	kind: PriorityClass
	metadata:
	  name: medium
	value: 1000000

	# 创建 critical 和 medium 两个 priority 对象
	➜ kubectl create -f priorityClass.yaml
	priorityclass.scheduling.k8s.io/critical created
	priorityclass.scheduling.k8s.io/medium created
	```
2. 检查一下集群资源使用情况。总共有 3 个节点，每个节点有 4 张 gpu 显卡。
	```shell
	➜ arena top node
	NAME                       IPADDRESS     ROLE    STATUS  GPU(Total)  GPU(Allocated)
	cn-huhehaote.172.16.0.205  172.16.0.205  master  ready   0           0
	cn-huhehaote.172.16.0.206  172.16.0.206  master  ready   0           0
	cn-huhehaote.172.16.0.207  172.16.0.207  master  ready   0           0
	cn-huhehaote.172.16.0.208  172.16.0.208  <none>  ready   4           0
	cn-huhehaote.172.16.0.209  172.16.0.209  <none>  ready   4           0
	cn-huhehaote.172.16.0.210  172.16.0.210  <none>  ready   4           0
	-----------------------------------------------------------------------------------------
	Allocated/Total GPUs In Cluster:
	0/12 (0%)
	```
3. 提交一个 3 节点 4 卡的 medium 的 gpu 作业，占满资源，为了验证效果，我们可以将训练的 epoch 调大一点，延长训练时间，方便实验查看
	```shell
	➜ arena --loglevel info submit pytorch \
		--name=pytorch-priority-medium \
		--gpus=4 \
		--workers=3 \
		--image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
		--sync-mode=git \
		--sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
		--priority=medium \
		"python /root/code/mnist-pytorch/mnist.py --backend gloo --epochs 200"
	configmap/pytorch-priority-medium-pytorchjob created
	configmap/pytorch-priority-medium-pytorchjob labeled
	pytorchjob.kubeflow.org/pytorch-priority-medium created
	INFO[0000] The Job pytorch-priority-medium has been submitted successfully
	INFO[0000] You can run `arena get pytorch-priority-medium --type pytorchjob` to check the job status
	```
4. 查看 `medium` 任务运行状态，已经全部运行起来
	```shell
	➜ arena get pytorch-priority-medium
	STATUS: RUNNING
	NAMESPACE: default
	PRIORITY: medium
	TRAINING DURATION: 3m

	NAME                     STATUS   TRAINER     AGE  INSTANCE                          NODE
	pytorch-priority-medium  RUNNING  PYTORCHJOB  3m   pytorch-priority-medium-master-0  172.16.0.208
	pytorch-priority-medium  RUNNING  PYTORCHJOB  3m   pytorch-priority-medium-worker-0  172.16.0.210
	pytorch-priority-medium  RUNNING  PYTORCHJOB  3m   pytorch-priority-medium-worker-1  172.16.0.209
	```
5. 查看节点 gpu 卡使用情况，已经全部被占用
	```shell
	➜ arena top node
	NAME                       IPADDRESS     ROLE    STATUS  GPU(Total)  GPU(Allocated)
	cn-huhehaote.172.16.0.205  172.16.0.205  master  ready   0           0
	cn-huhehaote.172.16.0.206  172.16.0.206  master  ready   0           0
	cn-huhehaote.172.16.0.207  172.16.0.207  master  ready   0           0
	cn-huhehaote.172.16.0.208  172.16.0.208  <none>  ready   4           4
	cn-huhehaote.172.16.0.209  172.16.0.209  <none>  ready   4           4
	cn-huhehaote.172.16.0.210  172.16.0.210  <none>  ready   4           4
	-----------------------------------------------------------------------------------------
	Allocated/Total GPUs In Cluster:
	12/12 (100%)
	```
6. 提交一个单机 1 卡的 `critical` 的 gpu 作业，发起抢占
	```shell
	➜ arena --loglevel info submit pytorch \
		--name=pytorch-priority-critical \
		--gpus=1 \
		--image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
		--sync-mode=git \
		--sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
		--priority=critical \
		"python /root/code/mnist-pytorch/mnist.py --backend gloo --epochs 50"
	configmap/pytorch-priority-critical-pytorchjob created
	configmap/pytorch-priority-critical-pytorchjob labeled
	pytorchjob.kubeflow.org/pytorch-priority-critical created
	INFO[0000] The Job pytorch-priority-critical has been submitted successfully
	INFO[0000] You can run `arena get pytorch-priority-critical --type pytorchjob` to check the job status
	```
7. 查看 `critical` 任务运行状态，等待运行起来，完成抢占
	```shell
	➜  arena get pytorch-priority-critical
	arena get pytorch-priority-critical
	STATUS: RUNNING
	NAMESPACE: default
	PRIORITY: critical
	TRAINING DURATION: 22s

	NAME                       STATUS   TRAINER     AGE  INSTANCE                            NODE
	pytorch-priority-critical  RUNNING  PYTORCHJOB  22s  pytorch-priority-critical-master-0  172.16.0.208
	```
8. 查看 `medium` 的作业状态，已经变成 `FAILED` 了, 有一个 instance 由于抢占被删除了
	```shell
	➜  arena get pytorch-priority-medium
	STATUS: FAILED
	NAMESPACE: default
	PRIORITY: medium
	TRAINING DURATION: 1m

	NAME                     STATUS  TRAINER     AGE  INSTANCE                          NODE
	pytorch-priority-medium  FAILED  PYTORCHJOB  2m   pytorch-priority-medium-master-0  172.16.0.210
	pytorch-priority-medium  FAILED  PYTORCHJOB  2m   pytorch-priority-medium-worker-0  172.16.0.209
	```
9. 检查 `medium` 作业的 event, 可以看到他的 `pytorch-priority-medium-worker-1` 被驱逐了，而被驱逐的原因是 `critical` 作业的实例 `pytorch-priority-critical-master-0` 也在申请这个节点的资源，而节点已经没有额外的 gpu 资源，因此低优先级的 `medium` 作业被高优的任务抢占
	```shell
	➜ kubectl get events --field-selector involvedObject.name=pytorch-priority-medium-worker-1
	```
	![](24-pytorchjob-preempted.png) 