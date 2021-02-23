# Training job with image pull secret

You can use a private registry and set image pull secrets for training jobs(include tensorboard images). Assume the following images are in your private registry.

    # pytorch
    registry.cn-beijing.aliyuncs.com/ai-samples/pytorch-with-tensorboard-secret:1.5.1-cuda10.1-cudnn7-runtime

    # tf
    registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu

    # mpi
    registry.cn-beijing.aliyuncs.com/ai-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5

    # tensorboard (--tensorboard-image)

    registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.12.0-devel


## Create Image Pull Secrets

Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) with kubectl, it's a [imagePullSecrets](https://kubernetes.io/docs/concepts/containers/images/
) in following case.

    $ kubectl create secret docker-registry <REG_SECRET> --docker-server=<REGISTRY> --docker-username=<USERNAME> --docker-password=<PASSWORD> --docker-email=<EMAIL>

!!! note

    * REG_SECRET: is the name of the secret key, which can be defined by yourself.
    * REGISTRY: is your private registry address.
    * USERNAME: is username of your private registry.
    * PASSWORD: is password of your private registry.
    * EMAIL: is your email address, Optional.

For Example, use the following command to create a image pull secret:

    $ kubectl create secret docker-registry \
    lumo-secret \
    --docker-server=registry.cn-huhehaote.aliyuncs.com \
    --docker-username=******@test.aliyunid.com \
    --docker-password=******
    secret/lumo-secret created

You can check that the secret was created.

    $ kubectl get secrets | grep lumo-secret
    lumo-secret                                       kubernetes.io/dockerconfigjson        1      52s


## Submit a tfjob with imagePullSecrets

Submit the job by using ``--image-pull-secrets`` to specify the imagePullSecrets.

1\. Submit a tensorflow job, the following command is an example.

    $ arena submit tf \
        --name=tf-git-with-secret \
        --working-dir=/root \
        --gpus=1 \
        --image=registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu \
        --sync-mode=git \
        --sync-source=https://code.aliyun.com/xiaozhou/tensorflow-sample-code.git \
        --data=training-data:/mnist_data \
        --tensorboard \
        --tensorboard-image=registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.12.0-devel \
        --logdir=/mnist_data/tf_data/logs \
        --image-pull-secrets=lumo-secret \
        "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --log_dir /mnist_data/tf_data/logs  --data_dir /mnist_data/tf_data/"

!!! note

   * If you have many ``imagePullSecrets`` to use, you can use ``--image-pull-secrets`` multiple times, like:

        $ arena submit tf \
            --name=tf-git-with-secret \
            ... \
            --image-pull-secrets=lumo-secret \
            --image-pull-secrets=king-secret \
            --image-pull-secrets=test-secret
            ...   

2\. Get the details of the job.

    $ arena get tf-git-with-secret
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 17s
   
    NAME                STATUS   TRAINER  AGE  INSTANCE                    NODE
    tf-git-with-secret  RUNNING  TFJOB    17s  tf-git-with-secret-chief-0  172.16.0.202
   
    Your tensorboard will be available on:
    http://172.16.0.198:30080


## Submit a mpijob with imagePullSecrets


Submit the mpi job by using ``--image-pull-secrets`` to specify the imagePullSecrets. 

1\. Submit mpi job, the following command is an example:

    $ arena submit mpi \
        --name=mpi-dist-with-secret \
        --gpus=1 \
        --workers=2 \
        --image=registry.cn-beijing.aliyuncs.com/ai-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
        --env=GIT_SYNC_BRANCH=cnn_tf_v1.9_compatible \
        --sync-mode=git \
        --sync-source=https://github.com/tensorflow/benchmarks.git \
        --tensorboard \
        --tensorboard-image=registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.12.0-devel \
        --image-pull-secrets=lumo-secret  \
        "mpirun python code/benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64 --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"

2\. Get the details of the job.

    $ arena get mpi-dist-with-secret
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 9m
    
    NAME                  STATUS   TRAINER  AGE  INSTANCE                             NODE
    mpi-dist-with-secret  RUNNING  MPIJOB   9m   mpi-dist-with-secret-launcher-v8sgt  172.16.0.201
    mpi-dist-with-secret  RUNNING  MPIJOB   9m   mpi-dist-with-secret-worker-0        172.16.0.201
    mpi-dist-with-secret  RUNNING  MPIJOB   9m   mpi-dist-with-secret-worker-1        172.16.0.202
    
    Your tensorboard will be available on:
    http://172.16.0.198:30450

## Submit a pytorchjob with imagePullSecrets
   
Submit the pytorchjob by using ``--image-pull-secrets`` to specify the imagePullSecrets. 

1\. Submit pytorch job, the following command is an example:

    $ arena submit pytorch \
       --name=pytorch-git-with-secret \
       --gpus=1 \
       --working-dir=/root \
       --image=registry.cn-beijing.aliyuncs.com/ai-samples/pytorch-with-tensorboard-secret:1.5.1-cuda10.1-cudnn7-runtime \
       --sync-mode=git \
       --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
       --data=training-data:/mnist_data \
       --tensorboard \
       --tensorboard-image=registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.12.0-devel \
       --logdir=/mnist_data/pytorch_data/logs \
       --image-pull-secrets=lumo-secret \
       "python /root/code/mnist-pytorch/mnist.py --epochs 10 --backend nccl --dir /mnist_data/pytorch_data/logs --data /mnist_data/pytorch_data/"

2\. Get the details of the job.

    $ arena get pytorch-git-with-secret
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 2m

    NAME                     STATUS   TRAINER     AGE  INSTANCE                          NODE
    pytorch-git-with-secret  RUNNING  PYTORCHJOB  2m   pytorch-git-with-secret-master-0  172.16.0.202

    Your tensorboard will be available on:
    http://172.16.0.198:31155

## Load imagePullSecrets from arena configuration file

If you don't want to submit job by ``--image-pull-secrets`` every time. You can replace it with configuration of Arena.
Open the file ``~/.arena/config``, if not exist, create it. And fill in the following configurations.

    imagePullSecrets=lumo-secret,king-secret

!!! note

    * ``--image-pull-secrets`` will overwrite ``~/.arena/config``.

