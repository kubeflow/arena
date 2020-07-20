# Using a Private Registry for jobs

You can use a private registry when submiting jobs(include tensorboard images).
Suppose the images in the following list from your private registry.
```shell
# pytorch
registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard-secret:1.5.1-cuda10.1-cudnn7-runtime
# tf
registry.cn-huhehaote.aliyuncs.com/lumo/tensorflow:1.5.0-devel-gpu
# mpi
registry.cn-huhehaote.aliyuncs.com/lumo/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5
# tensorboard (--tensorboard-image)
registry.cn-huhehaote.aliyuncs.com/lumo/tensorflow:1.12.0-devel
```

## Contents
* <a href="#create_secret">Create ImagePullSecrets</a>
* <a href="#tfjob">TFJob With Secret</a>
* <a href="#mpijob">MPIJob With Secret</a>
* <a href="#pytorchjob">PyTorchJob With Secret</a>


## <a name="create_secret">Create ImagePullSecrets</a>
* Create a [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) with kubectl. In this case, it's [imagePullSecrets](https://kubernetes.io/docs/concepts/containers/images/). 
    ```shell script
    kubectl create secret docker-registry [$Reg_Secret] --docker-server=[$Registry] --docker-username=[$Username] --docker-password=[$Password] --docker-email=[$Email]
    ```
    > Note：
    > [$Reg_Secret] is the name of the secret key, which can be defined by yourself.
    > [$Registry] is your private registry address.
    > [$Username] is username of your private registry.
    > [$Password] is password of your private registry.
    > [$Email] is your email address, Optional.

    For Example:
    ```shell
    kubectl create secret docker-registry \
    lumo-secret \
    --docker-server=registry.cn-huhehaote.aliyuncs.com \
    --docker-username=******@test.aliyunid.com \
    --docker-password=******
    secret/lumo-secret created
    ```
    You can check that the secret was created.
    ```shell
    # kubectl get secrets | grep lumo-secret
    lumo-secret                                       kubernetes.io/dockerconfigjson        1      52s
    ```
  
## <a name="tfjob">TFJob With Secret</a> 
Submit the job by using `--image-pull-secrets` to specify the imagePullSecrets.
1. Submit tf job.
    ```shell
    arena submit tf \
              --name=tf-git-with-secret \
              --working-dir=/root \
              --gpus=1 \
              --image=registry.cn-huhehaote.aliyuncs.com/lumo/tensorflow:1.5.0-devel-gpu \
              --sync-mode=git \
              --sync-source=https://code.aliyun.com/xiaozhou/tensorflow-sample-code.git \
              --data=training-data:/mnist_data \
              --tensorboard \
              --tensorboard-image=registry.cn-huhehaote.aliyuncs.com/lumo/tensorflow:1.12.0-devel \
              --logdir=/mnist_data/tf_data/logs \
              --image-pull-secrets=lumo-secret \
              "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --log_dir /mnist_data/tf_data/logs  --data_dir /mnist_data/tf_data/"
    configmap/tf-git-with-secret-tfjob created
    configmap/tf-git-with-secret-tfjob labeled
    service/tf-git-with-secret-tensorboard created
    deployment.apps/tf-git-with-secret-tensorboard created
    tfjob.kubeflow.org/tf-git-with-secret created
    INFO[0001] The Job tf-git-with-secret has been submitted successfully
    INFO[0001] You can run `arena get tf-git-with-secret --type tfjob` to check the job status
    ```
   > Note:
   > If you have many `imagePullSecrets` to use, you can use `--image-pull-secrets` multiple times.
   ```shell
   arena submit tf \
         --name=tf-git-with-secret \
         ... \
         --image-pull-secrets=lumo-secret \
         --image-pull-secrets=king-secret \
         --image-pull-secrets=test-secret
         ...   
   ```
2. Get the details of the job.
   ```shell 
   # arena get tf-git-with-secret
   STATUS: RUNNING
   NAMESPACE: default
   PRIORITY: N/A
   TRAINING DURATION: 17s
   
   NAME                STATUS   TRAINER  AGE  INSTANCE                    NODE
   tf-git-with-secret  RUNNING  TFJOB    17s  tf-git-with-secret-chief-0  172.16.0.202
   
   Your tensorboard will be available on:
   http://172.16.0.198:30080
   ```
  
## <a name="mpijob">MPIJob With Secret</a>
Submit the job by using `--image-pull-secrets` to specify the imagePullSecrets.         
1. Submit mpi job.
   ```shell 
   arena submit mpi \
          --name=mpi-dist-with-secret \
          --gpus=1 \
          --workers=2 \
          --image=registry.cn-huhehaote.aliyuncs.com/lumo/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
          --env=GIT_SYNC_BRANCH=cnn_tf_v1.9_compatible \
          --sync-mode=git \
          --sync-source=https://github.com/tensorflow/benchmarks.git \
          --tensorboard \
          --tensorboard-image=registry.cn-huhehaote.aliyuncs.com/lumo/tensorflow:1.12.0-devel \
          --image-pull-secrets=lumo-secret  \
          "mpirun python code/benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64 --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10"
   configmap/mpi-dist-with-secret-mpijob created
   configmap/mpi-dist-with-secret-mpijob labeled
   service/mpi-dist-with-secret-tensorboard created
   deployment.apps/mpi-dist-with-secret-tensorboard created
   mpijob.kubeflow.org/mpi-dist-with-secret created
   INFO[0001] The Job mpi-dist-with-secret has been submitted successfully
   INFO[0001] You can run `arena get mpi-dist-with-secret --type mpijob` to check the job status
   ```
2. Get the details of the job.
   ```shell 
   # arena get mpi-dist-with-secret
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
   ```

## <a name="pytorchjob">PyTorchJob With Secret</a>     
Submit the job by using `--image-pull-secrets` to specify the imagePullSecrets.  
1. Submit pytorch job.
   ```shell
   arena submit pytorch \
       --name=pytorch-git-with-secret \
       --gpus=1 \
       --working-dir=/root \
       --image=registry.cn-huhehaote.aliyuncs.com/lumo/pytorch-with-tensorboard-secret:1.5.1-cuda10.1-cudnn7-runtime \
       --sync-mode=git \
       --sync-source=https://code.aliyun.com/370272561/mnist-pytorch.git \
       --data=training-data:/mnist_data \
       --tensorboard \
       --tensorboard-image=registry.cn-huhehaote.aliyuncs.com/lumo/tensorflow:1.12.0-devel \
       --logdir=/mnist_data/pytorch_data/logs \
       --image-pull-secrets=lumo-secret \
       "python /root/code/mnist-pytorch/mnist.py --epochs 10 --backend nccl --dir /mnist_data/pytorch_data/logs --data /mnist_data/pytorch_data/"
   configmap/pytorch-git-with-secret-pytorchjob created
   configmap/pytorch-git-with-secret-pytorchjob labeled
   service/pytorch-git-with-secret-tensorboard created
   deployment.apps/pytorch-git-with-secret-tensorboard created
   pytorchjob.kubeflow.org/pytorch-git-with-secret created
   INFO[0001] The Job pytorch-git-with-secret has been submitted successfully
   INFO[0001] You can run `arena get pytorch-git-with-secret --type pytorchjob` to check the job status
   ```
2. Get the details of the job.
   ```shell 
   # arena get pytorch-git-with-secret
   STATUS: RUNNING
   NAMESPACE: default
   PRIORITY: N/A
   TRAINING DURATION: 2m
    
   NAME                     STATUS   TRAINER     AGE  INSTANCE                          NODE
   pytorch-git-with-secret  RUNNING  PYTORCHJOB  2m   pytorch-git-with-secret-master-0  172.16.0.202
    
   Your tensorboard will be available on:
   http://172.16.0.198:31155
   ```
