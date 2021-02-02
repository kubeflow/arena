The Distributed Tensorflow job has some roles, includes: Worker,PS,Chief,Evaluator. Sometimes, you may need to decide the sequence when creating them, for example, you may need to create "Worker" role first and then create "PS" role second, This guide will help you.

1. Now, assume that you want to submit a Distributed Tensorflow job，the tensorflow job has four roles: Worker,PS,Chief,Evaluator and you need the role starting sequence is "Worker,Chief,PS,Evaluator", it is simple for you only add option "--role-sequence" when submitting the job,the following command is an example:

```
    $ arena submit tfjob \
    --name=tf-distributed-test \
    --role-sequence "Worker,Chief,PS,Evaluator" \
    --chief \
    --evaluator \
    --gpus=1 \
    --workers=1 \
    --worker-image=cheyang/tf-mnist-distributed:gpu \
    --ps-image=cheyang/tf-mnist-distributed:cpu \
    --ps=1 \
    --tensorboard \
    --tensorboard-image="registry.cn-hongkong.aliyuncs.com/ai-samples/tensorflow:1.12.0-devel" \
    "python /app/main.py"
```

the "--role-sequence Worker,Chief,PS,Evaluator" is the same as "--role-sequence w,c,p,e" and "w" represents "Worker", "c" represents "Chief", "p" represents "PS" and "e" represents "Evaluator". 

2. you can validate it by querying the tf-operator logs.

```
    $ kubectl get po -n arena-system
    NAME                                READY   STATUS    RESTARTS   AGE
    et-operator-576887864c-lvmrs        1/1     Running   1          19d
    mpi-operator-66b4cf9b76-kl2fm       1/1     Running   0          26d
    pytorch-operator-8545c46f98-cffgw   1/1     Running   4          26d
    tf-job-dashboard-78478bfc45-msbzn   1/1     Running   0          19d
    tf-job-operator-554d594cff-5vxfg    1/1     Running   0          101m
```

query the logs of tf-job-operator-554d594cff-5vxfg.

```
$  kubectl logs tf-job-operator-554d594cff-5vxfg -n arena-system  | grep "the Role Sequence" | tail -n 1
{"filename":"tensorflow/controller.go:453","job":"default.tf-distributed-test","level":"info","msg":"the Role Sequence of job tf-distributed-test is: [Worker Chief PS Evaluator]","time":"2021-02-01T13:22:23Z","uid":"7db02629-4591-4e0c-a938-c6e4a1cfc074"}
```

As you see the sequence of tf-operator handles the tfjob roles is match the sequence you specified.

If you don't want to specify the role sequence every time when submitting the tfjob, you can save the role sequence to the arena configuration file "~/.arena/config", like: 

```
tfjob_role_sequence = Worker,PS,Chief,Evaluator
```

or 

```
tfjob_role_sequence = w,p,c,e
```