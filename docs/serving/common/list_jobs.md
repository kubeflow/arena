# List serving jobs

If you want to get all serving jobs' names, the command ``arena serve list`` can help you.

1\. List all serving jobs.

    $ arena serve list
    NAME                 TYPE        VERSION       DESIRED  AVAILABLE  ADDRESS       PORTS
    mymnist1             Tensorflow  202101162119  1        0          172.28.3.123  GRPC:8500,RESTFUL:8501
    fast-style-transfer  Custom      alpha         1        0          172.28.14.93  RESTFUL:31129->5000

As you see, there is two serving job types: ``Tensorflow`` and ``Custom``.

2\. If you want to list all serving jobs by the serving job type, you can use ``-T`` or ``--type`` to specify the job type. For example, the following command is used to list all custom serving jobs:

    $ arena serve list -T custom
    NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ADDRESS       PORTS
    fast-style-transfer  Custom  alpha    1        0          172.28.14.93  RESTFUL:31129->5000

3\. ``arena serve list`` will list the all serving jobs in the ``default`` namespace, if you want to get all serving jobs in other namespaces, ``-n`` or ``--namespace`` can help you. The example command will list all serving jobs in namespace ``test``.

    $  arena serve list  -n test 
    NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ADDRESS       PORTS
    fast-style-transfer  Custom  alpha    1        0          172.28.14.93  RESTFUL:31129->5000

!!! note

    * ``--all-namespaces`` or ``-A`` will list all serving jobs in all namespaces.

4\. If you want to get the output of ``arena serve list`` with json(or yaml) format, ``-o json`` (or ``-o yaml``) can help you.

    $ arena serve list -o json
    [
        {
            "name": "fast-style-transfer",
            "namespace": "default",
            "type": "Custom",
            "version": "alpha",
            "age": "5m",
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
                    "age": "5m",
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
        },
        {
            "name": "mymnist1",
            "namespace": "default",
            "type": "Tensorflow",
            "version": "202101162119",
            "age": "5m",
            "desiredInstances": 1,
            "availableInstances": 0,
            "endpoints": [
                {
                    "name": "GRPC",
                    "port": 8500,
                    "nodePort": 0
                },
                {
                    "name": "RESTFUL",
                    "port": 8501,
                    "nodePort": 0
                }
            ],
            "ip": "172.28.3.123",
            "instances": [
                {
                    "name": "mymnist1-202101162119-tensorflow-serving-694875fd47-mfhwc",
                    "status": "Pending",
                    "age": "5m",
                    "readyContainers": 0,
                    "totalContainers": 1,
                    "restartCount": 0,
                    "nodeIP": "",
                    "nodeName": "",
                    "ip": "",
                    "requestGPUs": 1,
                    "requestGPUMemory": 0
                }
            ],
            "requestGPUs": 1,
            "requestGPUMemory": 0
        }
    ]