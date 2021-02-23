# Tensorflow job with configuration files

The following steps will help you pass the configuration files to containers when submiting jobs.

1\. prepare the sample configuration files, create a test file which name is "test-config.json",its' path is "/tmp/test-config.json". we want push this file to containers of a tfjob (or mpijob) and the path in container is "/etc/config/config.json".

    $ cat /tmp/test-config.json
    {
        "key": "job-config"

    }

2\. submit a tfjob, and assign the configuration file with option ``--config-file``.

    $ arena submit tfjob \
        --name=tf \
        --gpus=1 \
        --workers=1 \
        --worker-image=cheyang/tf-mnist-distributed:gpu \
        --ps-image=cheyang/tf-mnist-distributed:cpu \
        --ps=1 \
        --tensorboard \
        --config-file /tmp/test-config.json:/etc/config/config.json \
        "python /app/main.py"


you can use ``--config-file <host_path_file>:<container_path_file>`` to assign a configuration file to containers.

!!! note

    there is some rules:

    * if assignd ``<host_path_file>`` and not assign ``<container_path_file>`` , we see ``<container_path_file>`` is the same as ``<host_path_file>``.
    * ``<container_path_file>`` must be a file with absolute path.
    *  you can use ``--config-file`` more than one in a command,eg: ``--config-file <file1>:<container_file1> --config-file <file2>:<container_file2>``.


3\. query the job details and make sure the job is "RUNNING".

    $ arena get tf
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 16s

    NAME  STATUS   TRAINER  AGE  INSTANCE     NODE
    tf    RUNNING  TFJOB    16s  tf-ps-0      192.168.7.18
    tf    RUNNING  TFJOB    16s  tf-worker-0  192.168.7.16

    Your tensorboard will be available on:
    http://192.168.7.10:31825


4\. use kubectl to check file is in container or not.

    $ kubectl exec -ti tf-ps-0 -- cat /etc/config/config.json
    {
        "key": "job-config"

    }

    $ kubectl exec -ti tf-worker-0 -- cat /etc/config/config.json
    {
        "key": "job-config"

    }

as you see,the file /etc/config/config.json is existed in the container.
