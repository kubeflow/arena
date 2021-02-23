## 为 pytorch 训练任务提供配置文件
在提交训练任务时，我们可以指定该训练任务所需的配置文件。

1. 在提交的机器上准备好要挂载的配置文件
	```shell
	# prepare your config-file
	➜ cat  /tmp/test-config.json
	{
		"key": "job-config"
	}
	```
2. 提交作业，通过 `--config-file` 指定要挂载的配置文件
	```shell
	# arena submit job by --config-file  ${host-config-file}:${container-config-file}
	# 该参数支持多次使用，挂载多个配置文件
	➜ arena --loglevel info submit pytorch \
            --name=pytorch-config-file \
            --gpus=1 \
            --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard:1.5.1-cuda10.1-cudnn7-runtime \
            --sync-mode=git \
            --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
            --config-file /tmp/test-config.json:/etc/config/config.json \
            "python /root/code/mnist-pytorch/mnist.py --epochs 50 --backend gloo"
	configmap/pytorch-config-file-pytorchjob created
	configmap/pytorch-config-file-pytorchjob labeled
	configmap/pytorch-config-file-a9cbad1b8719778 created
	pytorchjob.kubeflow.org/pytorch-config-file created
	INFO[0000] The Job pytorch-config-file has been submitted successfully
	INFO[0000] You can run `arena get pytorch-config-file --type pytorchjob` to check the job status
	```
3. 查询作业详情
	```shell
	➜ arena get pytorch-config-file --type pytorchjob
	STATUS: RUNNING
	NAMESPACE: default
	PRIORITY: N/A
	TRAINING DURATION: 51s

	NAME                 STATUS   TRAINER     AGE  INSTANCE                      NODE
	pytorch-config-file  RUNNING  PYTORCHJOB  51s  pytorch-config-file-master-0  172.16.0.210
	```
4. 使用 kubectl 检测文件是否已经存放在了任务的实例中:
    ```
    ➜ kubectl exec -ti pytorch-config-file-master-0 -- cat /etc/config/config.json
    {
        "key": "job-config"
    }
    ```