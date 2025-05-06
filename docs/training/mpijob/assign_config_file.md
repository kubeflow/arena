# MPI job with specified configuration files

the following steps will help you pass the configuration files to containers when submitting jobs.


1\. prepare the sample configuration files, create a test file which name is "test-config.json",its' path is "/tmp/test-config.json". we want push this file to containers of a mpi job and the path in container is "/etc/config/config.json".

    $ cat /tmp/test-config.json
    {
        "key": "job-config"

    }

2\. submit a mpijob, and assign the configuration file with option ``--config-file``.

    $ arena submit mpi --name=mpi-dist \
        --gpus=1 \
        --workers=1 \
        --image=centos:7 \
        --tensorboard \
        --config-file /tmp/test-config.json:/etc/config/config.json \
        "sleep 600"


you can use ``--config-file <host_path_file>:<container_path_file>`` to assign a configuration file to containers.

!!! note

    there is some rules:

    * if assignd ``<host_path_file>`` and not assign ``<container_path_file>`` , we see ``<container_path_file>`` is the same as ``<host_path_file>``.
    * ``<container_path_file>`` must be a file with absolute path.
    *  you can use ``--config-file`` more than one in a command,eg: ``--config-file <file1>:<container_file1> --config-file <file2>:<container_file2>``.


3\. query the job details and make sure the job is "RUNNING".

    $ arena get mpi-dist
    Name:        mpi-dist
    Status:      RUNNING
    Namespace:   default
    Priority:    N/A
    Trainer:     MPIJOB
    Duration:    19s

    Instances:
    NAME                     STATUS   AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                     ------   ---  --------  --------------  ----
    mpi-dist-launcher-6fwhd  Running  19s  true      0               cn-beijing.192.168.1.112
    mpi-dist-worker-0        Running  19s  false     1               cn-beijing.192.168.1.112

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.1.101:30600


4\. use kubectl to check file is in container or not.

    $ kubectl exec -ti mpi-dist-worker-0 -- cat /etc/config/config.json
    {
        "key": "job-config"

    }

    $ kubectl exec -ti mpi-dist-launcher-6fwhd -- cat /etc/config/config.json
    {
        "key": "job-config"

    }

as you see,the file /etc/config/config.json is existed in the containers.
