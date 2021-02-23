## Assign configuration files for pytorch jobs

You can pass the configuration files to containers when submiting jobs.

1. Prepare the configuration file to be mounted on the submitted machine.
	```shell
	# prepare your config-file
	➜ cat  /tmp/test-config.json
	{
		"key": "job-config"
	}
	```
2. Submit the job, and specify the configuration file to mount by `--config-file`.
	```shell
	# arena submit job by --config-file  ${host-config-file}:${container-config-file}
	# This parameter supports multiple use and mounting multiple configuration files
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
3. Get the details of the this job.
	```shell
	➜ arena get pytorch-config-file --type pytorchjob
	STATUS: RUNNING
	NAMESPACE: default
	PRIORITY: N/A
	TRAINING DURATION: 51s

	NAME                 STATUS   TRAINER     AGE  INSTANCE                      NODE
	pytorch-config-file  RUNNING  PYTORCHJOB  51s  pytorch-config-file-master-0  172.16.0.210
	```
4. Use kubectl to check file is in container or not:
    ```
    ➜ kubectl exec -ti pytorch-config-file-master-0 -- cat /etc/config/config.json
    {
        "key": "job-config"
    }
    ```