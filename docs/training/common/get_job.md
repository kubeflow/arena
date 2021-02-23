# Get training job details

You can use ``arena get`` to get the training job details, we will introduce some usages  about ``arena get``.

1\. Get the training job details.

    $ arena get mpi-dist

    Name:        mpi-dist
    Status:      SUCCEEDED
    Namespace:   default
    Priority:    N/A
    Trainer:     MPIJOB
    Duration:    10m

    Instances:
    NAME                     STATUS     AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                     ------     ---  --------  --------------  ----
    mpi-dist-launcher-6fwhd  Completed  10m  true      0               cn-beijing.192.168.1.112

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.1.101:30600

2\. you cant get the training job details with json(or yaml) format only add option ``-o json``(or ``-o yaml``).

    $ arena get mpi-dist -o json
    {
        "name": "mpi-dist",
        "namespace": "default",
        "duration": "618s",
        "status": "SUCCEEDED",
        "trainer": "mpijob",
        "tensorboard": "http://192.168.1.101:30600",
        "chiefName": "mpi-dist-launcher-6fwhd",
        "instances": [
            {
                "ip": "172.27.0.10",
                "status": "Completed",
                "name": "mpi-dist-launcher-6fwhd",
                "age": "601s",
                "node": "cn-beijing.192.168.1.112",
                "nodeIP": "192.168.1.112",
                "chief": true,
                "requestGPUs": 0,
                "gpuMetrics": {}
            }
        ],
        "priority": "N/A",
        "requestGPUs": 1,
        "allocatedGPUs": 0
    }

3\. ``-e`` can display the events of training job instances.

    $ arena get tf-standalone-test -e

    Name:        tf-standalone-test
    Status:      PENDING
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    9s

    Instances:
    NAME                        STATUS             AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                        ------             ---  --------  --------------  ----
    tf-standalone-test-chief-0  ContainerCreating  9s   true      0               N/A

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.1.101:32221

    Events:
    SOURCE                          TYPE    AGE  MESSAGE
    ------                          ----    ---  -------
    pod/tf-standalone-test-chief-0  Normal  9s   [Scheduled] Successfully assigned default/tf-standalone-test-chief-0 to cn-beijing.192.168.1.112
    pod/tf-standalone-test-chief-0  Normal  9s   [Pulling] Pulling image "registry.cn-beijing.aliyuncs.com/ai-samples/tensorflow:1.5.0-devel-gpu"

4\. When the status of training job is ``RUNNING``(or ``COMPLETEDED``), ``-g`` can display the gpu utilization of training job(this option depends prometheus).

    $ arena get tf-standalone-test -g

    Name:        tf-standalone-test
    Status:      RUNNING
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    6s

    Instances:
    NAME                        STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                        ------   ---  --------  --------------  ----
    tf-standalone-test-chief-0  Running  6s   true      1               cn-beijing.192.168.1.112

    GPUs:
    INSTANCE                    NODE(IP)       GPU(Requested)  GPU(IndexId)  GPU(DutyCycle)  GPU Memory(Used/Total)
    --------                    --------       --------------  ------------  --------------  ----------------------
    tf-standalone-test-chief-0  192.168.1.112  1               N/A           N/A             N/A

    Allocated/Requested GPUs of Job: 1/1

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.1.101:31869

5\. You can use ``--type`` to specify the training job type.

    $ arena get tf-standalone-test --type tf
    
    Name:        tf-standalone-test
    Status:      RUNNING
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    6m

    Instances:
    NAME                        STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                        ------   ---  --------  --------------  ----
    tf-standalone-test-chief-0  Running  6m   true      1               cn-beijing.192.168.1.112

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.1.101:31869