
# Arena supports and simplifies volcano job.

Volcano is a batch system built on Kubernetes. It provides a suite of mechanisms currently missing from
Kubernetes that are commonly required by many classes of batch & elastic workload including:

1. machine learning/deep learning,
2. bioinformatics/genomics, and
3. other "big data" applications.

## pre requisites

- k8s deployment
- deploy the volcano following the steps from kubernetes-artifacts/volcano-operator/README.md

### 1. To run a batch/distributed volcano job, you may need to specify:
--minAvailable int       The minimal available pods to run for this Job. default value is 1 (default 1)
--name string            override name
--queue string           Specifies the queue that will be used in the scheduler, default queue is used this leaves empty (default "default")
--schedulerName string   Specifies the scheduler Name, default is volcano when not specified (default "volcano")
--taskCPU string         cpu request for each task replica / pod. default value is 250m (default "250m")
--taskImages strings     the docker images of different tasks of volcano job. default used 3 tasks with ubuntu,nginx and busybox images (default [ubuntu,nginx,busybox])
--taskMemory string      memory request for each task replica/pod.default value is 128Mi) (default "128Mi")
--taskName string        the task name of volcano job, default value is task (default "task")
--taskPort int           the task port number. default value is 2222 (default 2222)
--taskReplicas int       the task replica's number to run the distributed tasks. default value is 1 (default 1)

### 2. More information related to volcano job.

Arena volcano job is based on (https://github.com/volcano-sh/volcano).
You can get more information related to volcano from https://volcano.sh/

### 3. How to use Arena volcano job

##### install volcano
 
deploy the volcano following the steps from kubernetes-artifacts/volcano-operator/README.md 

To install the chart with the release name `volcano-release`

```bash
$ helm install --name volcano-release --namespace arena-system kubernetes-artifacts/volcano-operator
```

TO verify all deployments are running use the below command

```bash
    kubectl get deployment --all-namespaces | grep {release_name}
```
We should get similar output like given below, where three deployments for controller, admission, scheduler should be running.


NAME                       READY  UP-TO-DATE  AVAILABLE  AGE
{release_name}-admission    1/1    1           1          4s
{release_name}-controllers  1/1    1           1          4s
{release_name}-scheduler    1/1    1           1          4s

TO verify all pods are running use the below command

```bash
    kubectl get pods --all-namespaces | grep {release_name}
```

We should get similar output like given below, where three pods for controller, admission,admissioninit, scheduler should be running.

NAMESPACE     NAME                                          READY    STATUS             RESTARTS   AGE
default       volcano-release-admission-cbfdb8549-dz5hg      1/1     Running            0          33s
default       volcano-release-admission-init-7xmzd           0/1     Completed          0          33s
default       volcano-release-controllers-7967fffb8d-7vnn9   1/1     Running            0          33s
default       volcano-release-scheduler-746f6557d8-9pfg6     1/1     Running            0          33s


##### submit a volcano job

```$xslt
arena submit volcanojob --name=demo
```

The result is like below.
```$xslt

configmap/demo-volcanojob created
configmap/demo-volcanojob labeled
job.batch.volcano.sh/demo created
INFO[0003] The Job demo has been submitted successfully
INFO[0003] You can run `arena get demo --type volcanojob` to check the job status

```

if we want to provide more command line parameters then
```$xslt
./bin/arena submit volcanojob --name demo12 --taskImages busybox,busybox  --taskReplicas 2
```

in above case it creates two tasks each with 2 replicas  as shown below
```$xslt
arena get --type volcanojob demo12
```
the result is as below
```$xslt
STATUS: SUCCEEDED
NAMESPACE: default
TRAINING DURATION: 2m

NAME    STATUS     TRAINER     AGE  INSTANCE         NODE
demo12  SUCCEEDED  VOLCANOJOB  2m   demo12-task-0-0  11.245.101.184
demo12  SUCCEEDED  VOLCANOJOB  2m   demo12-task-0-1  11.245.101.184
demo12  SUCCEEDED  VOLCANOJOB  2m   demo12-task-1-0  11.245.101.184
demo12  SUCCEEDED  VOLCANOJOB  2m   demo12-task-1-1  11.245.101.184
```
##### get volcano job status

```$xslt
arena get --type=volcanojob demo
```
When the job running/succeed,you will see the result below.
```$xslt
STATUS: RUNNING/SUCCEEDED
NAMESPACE: default
TRAINING DURATION: 45s

NAME  STATUS     TRAINER     AGE  INSTANCE       NODE
demo  SUCCEEDED  VOLCANOJOB  59s  demo-task-0-0  11.245.101.184
demo  RUNNING    VOLCANOJOB  59s  demo-task-1-0  11.245.101.184
demo  SUCCEEDED  VOLCANOJOB  59s  demo-task-2-0  11.245.101.184

```
##### list arena jobs

```$xslt
arena list
```
we can observe the below data
```$xslt
NAME     STATUS   TRAINER     AGE  NODE
demo     RUNNING  VOLCANOJOB  2m   11.245.101.184
```

##### delete volcano job

```$xslt
arena delete --type=volcanojob demo
```
You will found the volcano job is deleted.
```$xslt
job.batch.volcano.sh "demo" deleted
configmap "demo-volcanojob" deleted
INFO[0000] The Job demo has been deleted successfully
```

Congratulations! You've run the batch/distributed volcano job with `arena` successfully.