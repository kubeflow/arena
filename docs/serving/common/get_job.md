# Get serving job details

You can use ``arena serve get`` to get the serving job details, we will introduce some usages  about ``arena serve get``.

1\. Get the serving job details.

    $ arena serve get fast-style-transfer
    Name:           fast-style-transfer
    Namespace:      default
    Type:           Custom
    Version:        alpha
    Desired:        1
    Available:      1
    Age:            11m
    Address:        172.28.14.93
    Port:           RESTFUL:31129->5000
    GPUs:           1

    Instances:
    NAME                                                       STATUS   AGE  READY  RESTARTS  GPUs  NODE
    ----                                                       ------   ---  -----  --------  ----  ----
    fast-style-transfer-alpha-custom-serving-856dbcdbcb-sxx2n  Running  11m  1/1    0         1     cn-beijing.192.168.1.112

2\. you cant get the serving job details with json(or yaml) format only add option ``-o json``(or ``-o yaml``).

    $ arena serve get fast-style-transfer -o json
    {
        "name": "fast-style-transfer",
        "namespace": "default",
        "type": "Custom",
        "version": "alpha",
        "age": "13m",
        "desiredInstances": 1,
        "availableInstances": 1,
        "endpoints": [
            {
                "name": "RESTFUL",
                "port": 5000,
                "nodePort": 31129
            }
        ],
        "ip": "172.28.14.93",
        "instances": [
            {
                "name": "fast-style-transfer-alpha-custom-serving-856dbcdbcb-sxx2n",
                "status": "Running",
                "age": "13m",
                "readyContainers": 1,
                "totalContainers": 1,
                "restartCount": 0,
                "nodeIP": "192.168.1.112",
                "nodeName": "cn-beijing.192.168.1.112",
                "ip": "172.27.0.35",
                "requestGPUs": 1,
                "requestGPUMemory": 0
            }
        ],
        "requestGPUs": 1,
        "requestGPUMemory": 0
    }

3\. You can use ``--type``(or ``-T``) to specify the serving job type.

    $ arena serve get fast-style-transfer -T custom
    Name:           fast-style-transfer
    Namespace:      default
    Type:           Custom
    Version:        alpha
    Desired:        1
    Available:      1
    Age:            14m
    Address:        172.28.14.93
    Port:           RESTFUL:31129->5000
    GPUs:           1

    Instances:
    NAME                                                       STATUS   AGE  READY  RESTARTS  GPUs  NODE
    ----                                                       ------   ---  -----  --------  ----  ----
    fast-style-transfer-alpha-custom-serving-856dbcdbcb-sxx2n  Running  14m  1/1    0         1     cn-beijing.192.168.1.112

4\. You can use ``-v``(or ``--version``) to specify the serving job version.

    $ arena serve get fast-style-transfer -T custom -v alpha
    Name:           fast-style-transfer
    Namespace:      default
    Type:           Custom
    Version:        alpha
    Desired:        1
    Available:      1
    Age:            16m
    Address:        172.28.14.93
    Port:           RESTFUL:31129->5000
    GPUs:           1

    Instances:
    NAME                                                       STATUS   AGE  READY  RESTARTS  GPUs  NODE
    ----                                                       ------   ---  -----  --------  ----  ----
    fast-style-transfer-alpha-custom-serving-856dbcdbcb-sxx2n  Running  16m  1/1    0         1     cn-beijing.192.168.1.112