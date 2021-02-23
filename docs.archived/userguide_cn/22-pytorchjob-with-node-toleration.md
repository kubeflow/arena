## 支持 pytorch 任务运行在 k8s 污点上

1. 查看 k8s 集群节点信息
	```shell
	➜ kubectl get node
	NAME                        STATUS   ROLES    AGE     VERSION
	cn-huhehaote.172.16.0.205   Ready    master   5h13m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.206   Ready    master   5h12m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.207   Ready    master   5h11m   v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.208   Ready    <none>   5h7m    v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.209   Ready    <none>   5h7m    v1.16.9-aliyun.1
	cn-huhehaote.172.16.0.210   Ready    <none>   5h7m    v1.16.9-aliyun.1
	```
2. 给节点加上 taints
	```shell
	# taint --> gpu_node
	➜  kubectl taint nodes cn-huhehaote.172.16.0.208 gpu_node=invalid:NoSchedule
	node/cn-huhehaote.172.16.0.208 tainted
	➜  kubectl taint nodes cn-huhehaote.172.16.0.209 gpu_node=invalid:NoSchedule
	node/cn-huhehaote.172.16.0.209 tainted
	# taint --> ssd_node
	➜  kubectl taint nodes cn-huhehaote.172.16.0.210 ssd_node=invalid:NoSchedule
	node/cn-huhehaote.172.16.0.210 tainted
	```
3. 当我们加错节点的 taints 或者恢复节点的调度性，可以通过以下方式去除节点的 taints
	```shell
	➜ kubectl taint nodes cn-huhehaote.172.16.0.208 gpu_node-
	node/cn-huhehaote.172.16.0.208 untainted
	➜ kubectl taint nodes cn-huhehaote.172.16.0.209 gpu_node-
	node/cn-huhehaote.172.16.0.209 untainted
	➜ kubectl taint nodes cn-huhehaote.172.16.0.210 ssd_node-
	node/cn-huhehaote.172.16.0.210 untainted
	```
4. 提交作业，使用 `--toleration` 可以容忍的节点 key, 如：`--toleration=gpu_node`；该参数可以使用多次，带上不同 taint key
	```shell
	➜ arena --loglevel info submit pytorch \
            --name=pytorch-toleration \
            --gpus=1 \
            --workers=2 \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            --tensorboard \
            --logdir=/root/logs \
            --toleration gpu_node \
            "python /root/code/mnist-pytorch/mnist.py --epochs 50 --backend gloo --dir /root/logs"
	configmap/pytorch-toleration-pytorchjob created
	configmap/pytorch-toleration-pytorchjob labeled
	service/pytorch-toleration-tensorboard created
	deployment.apps/pytorch-toleration-tensorboard created
	pytorchjob.kubeflow.org/pytorch-toleration created
	INFO[0000] The Job pytorch-toleration has been submitted successfully
	INFO[0000] You can run `arena get pytorch-toleration --type pytorchjob` to check the job status
	```
5. 查看 job 是否被调度起来
	```shell
	arena get pytorch-toleration
	STATUS: RUNNING
	NAMESPACE: default
	PRIORITY: N/A
	TRAINING DURATION: 2m

	NAME                STATUS   TRAINER     AGE  INSTANCE                     NODE
	pytorch-toleration  RUNNING  PYTORCHJOB  2m   pytorch-toleration-master-0  172.16.0.209
	pytorch-toleration  RUNNING  PYTORCHJOB  2m   pytorch-toleration-worker-0  172.16.0.209

	Your tensorboard will be available on:
	http://172.16.0.205:32091
	```
6. `--toleration` 通过指定 all, 标记这个任务可以容忍所有节点
	```shell
	➜ arena --loglevel info submit pytorch \
            --name=pytorch-toleration-all \
            --gpus=1 \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            --toleration all \
            "python /root/code/mnist-pytorch/mnist.py --epochs 10 --backend gloo"
	configmap/pytorch-toleration-all-pytorchjob created
	configmap/pytorch-toleration-all-pytorchjob labeled
	pytorchjob.kubeflow.org/pytorch-toleration-all created
	INFO[0000] The Job pytorch-toleration-all has been submitted successfully
	INFO[0000] You can run `arena get pytorch-toleration-all --type pytorchjob` to check the job status
	```
7. 查看 job 是否被调度起来
	```shell
	➜ arena get pytorch-toleration-all
	STATUS: RUNNING
	NAMESPACE: default
	PRIORITY: N/A
	TRAINING DURATION: 33s

	NAME                    STATUS   TRAINER     AGE  INSTANCE                         NODE
	pytorch-toleration-all  RUNNING  PYTORCHJOB  33s  pytorch-toleration-all-master-0  172.16.0.210
	```