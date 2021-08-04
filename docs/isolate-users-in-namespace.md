# Isolate users in namespace

Sometimes, you submit a job(training job or serving job) in a namespace, but you want others have no privileges to operate(eg: list,get,delete...) the job you submit in the namespace, this doc can help you.


## Create namespace

Firstly, admin user create a namespace and label the namespace with 'arena.kubeflow.org/isolate-user=true'. In the sample,we create a namespace whose name is 'training'.

```
$ kubectl create ns training

$ kubectl label ns training  arena.kubeflow.org/isolate-user=true
```

the label 'arena.kubeflow.org/isolate-user=true' represents the namespace should isolate users, if you no need to isolate users,you can delete the label.

## Create Two Users

To show the effect, we will create two users(eg: 'tom' and 'bob') by the script arena-gen-kubeconfig.sh, this step should done by admin user.

Create user 'tom' and make sure he can use namespace 'training'. 

```
$ arena-gen-kubeconfig.sh --user-name tom --user-namespace training

2021-08-03/17:06:04  DEBUG  found arena charts in /Users/yangjunfeng/charts
2021-08-03/17:06:04  DEBUG  the user configuration not set,use the default configuration file
resourcequota/arena-quota-tom created
serviceaccount/tom created
clusterrole.rbac.authorization.k8s.io/arena:training:tom created
clusterrolebinding.rbac.authorization.k8s.io/arena:training:tom created
role.rbac.authorization.k8s.io/arena:tom created
rolebinding.rbac.authorization.k8s.io/arena:tom created
configmap/arena-user-tom created
Cluster "https://39.107.123.34:6443" set.
User "tom" set.
Context "tom" created.
Switched to context "tom".
2021-08-03/17:06:05  DEBUG  kubeconfig written to file ./tom.kubeconfig
```

This script will generate the kubeconfig file in the current directory and name is 'tom.kubeconfig'.

Then create the kubeconfig file for user bob.

```
$ arena-gen-kubeconfig.sh --user-name bob --user-namespace training

2021-08-03/17:11:40  DEBUG  namespace training has been existed,skip to create it
2021-08-03/17:11:40  DEBUG  found arena charts in /Users/yangjunfeng/charts
2021-08-03/17:11:40  DEBUG  the user configuration not set,use the default configuration file
resourcequota/arena-quota-bob created
serviceaccount/bob created
clusterrole.rbac.authorization.k8s.io/arena:training:bob created
clusterrolebinding.rbac.authorization.k8s.io/arena:training:bob created
role.rbac.authorization.k8s.io/arena:bob created
rolebinding.rbac.authorization.k8s.io/arena:bob created
configmap/arena-user-bob created
Cluster "https://39.107.123.34:6443" set.
User "bob" set.
Context "bob" created.
Switched to context "bob".
2021-08-03/17:11:41  DEBUG  kubeconfig written to file ./bob.kubeconfig
```

The kubeconfig file is stored in ./bob.kubeconfig


## Submit a Training Job by user tom

Firstly,submit a training job by user tom.

```
$ export KUBECONFIG=./tom.kubeconfig

$ arena submit mpijob \
	--name=mpi-test-tom \
	--gpus=1 \
	--workers=2 \
	--working-dir=/perseus-demo/tensorflow-demo/ \
	--image=registry.cn-hongkong.aliyuncs.com/ai-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
	'mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10'
```

Then,list the training jobs.

```
$ arena list
NAME          STATUS   TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
mpi-test-tom  RUNNING  MPIJOB   6s        2               2               192.168.6.83
```

Get the training job information.

```
$ arena get mpi-test-tom
Name:      mpi-test-tom
Status:    RUNNING
Namespace: default
Priority:  N/A
Trainer:   MPIJOB
Duration:  15s

Instances:
  NAME                         STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
  ----                         ------   ---  --------  --------------  ----
  mpi-test-tom-launcher-2jwqj  Running  15s  true      0               cn-beijing.192.168.6.83
  mpi-test-tom-worker-0        Running  15s  false     0               cn-beijing.192.168.6.83
  mpi-test-tom-worker-1        Running  15s  false     0               cn-beijing.192.168.6.84 
```

## Submit a Training Job by user bob


Firstly,submit a training job by user tom.

```
$ export KUBECONFIG=./bob.kubeconfig

$ arena submit mpijob \
	--name=mpi-test-bob \
	--gpus=1 \
	--workers=2 \
	--working-dir=/perseus-demo/tensorflow-demo/ \
	--image=registry.cn-hongkong.aliyuncs.com/ai-samples/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5 \
	'mpirun python /benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3 --save_summaries_steps=10'
```

List the training jobs.

```
$ arena list
NAME          STATUS   TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
mpi-test-bob  PENDING  MPIJOB   5s        2               2               N/A
```

As you can see, the user 'bob' only find a training job created by him and the training job 'mpi-test-tom' is not visible for him.

If you get the training job 'mpi-test-tom' information,arena will return an error.

```
$ arena get mpi-test-tom
ERRO[0000] you have no privileges to operate the job,because the owner of job is not you
```

And delete the training job  mpi-test-tom also return an error.

```
$ arena delete mpi-test-tom
ERRO[0000] you have no privileges to operate the job,because the owner of job is not you
```
