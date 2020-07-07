## 指定节点运行 pytorch 任务

1. 查询 k8s 集群信息
	```shell
	➜ kubectl get nodes
	NAME                        STATUS   ROLES    AGE     VERSION
	cn-huhehaote.172.16.0.205   Ready    master   4h19m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.206   Ready    master   4h18m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.207   Ready    master   4h17m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.208   Ready    <none>   4h13m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.209   Ready    <none>   4h13m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.210   Ready    <none>   4h13m   v1.16.9-aliyun.1
	```
2. 给不同的节点打上不同的标签
	```shell
	# 172.16.0.208 打上 gpu_node=ok
	➜ kubectl label nodes cn-huhehaote.172.16.0.208 gpu_node=ok
	node/cn-huhehaote.172.16.0.208 labeled
	# 172.16.0.209 打上 gpu_node=ok
	➜ kubectl label nodes cn-huhehaote.172.16.0.209 gpu_node=ok
	node/cn-huhehaote.172.16.0.209 labeled
	# 172.16.0.210 打上 ssd_node=ok
	➜ kubectl label nodes cn-huhehaote.172.16.0.210 ssd_node=ok
	node/cn-huhehaote.172.16.0.210 labeled
	```
3. 提交 pytorch 作业的时候，通过 `--selector` 选定 job 运行在哪个标签的节点上
	```shell
	➜ arena --loglevel info submit pytorch \
            --name=pytorch-selector \
            --gpus=1 \
            --workers=2 \
            --selector gpu_node=ok \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            "python /root/code/mnist-pytorch/mnist.py --backend gloo"
	configmap/pytorch-selector-pytorchjob created
	configmap/pytorch-selector-pytorchjob labeled
	pytorchjob.kubeflow.org/pytorch-selector created
	INFO[0000] The Job pytorch-selector has been submitted successfully
	INFO[0000] You can run `arena get pytorch-selector --type pytorchjob` to check the job status
	```
4. 查询 job 信息，可以看到作业 `pytorch-selector` 只运行在带有标签 `gpu_node=ok` 的节点 172.16.0.209 上
	```shell
	➜ arena get pytorch-selector
	STATUS: PENDING
	NAMESPACE: default
	PRIORITY: N/A
	TRAINING DURATION: 14s

	NAME              STATUS   TRAINER     AGE  INSTANCE                   NODE
	pytorch-selector  PENDING  PYTORCHJOB  14s  pytorch-selector-master-0  172.16.0.209
	pytorch-selector  PENDING  PYTORCHJOB  14s  pytorch-selector-worker-0  172.16.0.209
	```