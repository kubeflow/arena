## Specify the clean-up policy of pod after finishing for pytorch job

1. Submit a job, and specify `--clean-task-policy` as `All`. After the job finished (`SUCCEEDED` or `FAILED`), all instances (pods) will be deleted; the default is `None`, and all pods will be retained.
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

2. Get the job details. After the job is finished, the instance `python-clean-policy-master-0` has been deleted.
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