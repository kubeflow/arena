## 指定 pytorch 任务结束后 pod 的清理策略

1. 提交一个作业，指定 `--clean-task-policy` 为 `All`, 作业结束后（成功或者失败），将会删除所有 instance (pod)；默认为 `None`, 会保留所有 pod
	```shell
	➜ arena --loglevel info submit pytorch \
		--name=pytorch-clean-policy \
		--gpus=1 \
		--image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
		--sync-mode=git \
		--sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
		--clean-task-policy=All \
		"python /root/code/mnist-pytorch/mnist.py --backend gloo"
	configmap/pytorch-clean-policy-pytorchjob created
	configmap/pytorch-clean-policy-pytorchjob labeled
	pytorchjob.kubeflow.org/pytorch-clean-policy created
	INFO[0000] The Job pytorch-clean-policy has been submitted successfully
	INFO[0000] You can run `arena get pytorch-clean-policy --type pytorchjob` to check the job status
	```

2. 查看作业详情, 任务结束后，实例 `pytorch-clean-policy-master-0` 被删除
	```shell
    # RUNNING
    ➜ arena get pytorch-clean-policy
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 18s
    
    NAME                  STATUS   TRAINER     AGE  INSTANCE                       NODE
    pytorch-clean-policy  RUNNING  PYTORCHJOB  18s  pytorch-clean-policy-master-0  172.16.0.209
    
    # FINISHED
    ➜ arena get pytorch-clean-policy
    STATUS: SUCCEEDED
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 37s
 
    NAME  STATUS  TRAINER  AGE  INSTANCE  NODE
	```