## Arena supports assigning  pytorch jobs to some k8s particular nodes

1. Get k8s cluster information:
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
2. Give a label to nodes,for example:
	```shell
	# 172.16.0.208 label gpu_node=ok
	➜ kubectl label nodes cn-huhehaote.172.16.0.208 gpu_node=ok
	node/cn-huhehaote.172.16.0.208 labeled
	# 172.16.0.209 label gpu_node=ok
	➜ kubectl label nodes cn-huhehaote.172.16.0.209 gpu_node=ok
	node/cn-huhehaote.172.16.0.209 labeled
	# 172.16.0.210 label ssd_node=ok
	➜ kubectl label nodes cn-huhehaote.172.16.0.210 ssd_node=ok
	node/cn-huhehaote.172.16.0.210 labeled
	```
3. When submitting a python job, you can use the `--selector` to decide which node the job runs on
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
4. Get the job details, you can see that the job only runs on this node with IP 172.16.0.209 and label `gpu_node=ok`.
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