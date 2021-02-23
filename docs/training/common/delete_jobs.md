# Delete the training jobs

When the training jobs have been finished, you can delete them, for example, the following training jobs has been finished. 

    $ arena list
    NAME                           STATUS     TRAINER  DURATION  GPU(Requested)  GPU(Allocated)  NODE
    tf-standalone-test             SUCCEEDED  TFJOB    11m       1               N/A             192.168.1.112
    mpi-dist                       SUCCEEDED  MPIJOB   10m       1               N/A             192.168.1.112
    elastic-training               SUCCEEDED  ETJOB    2d        0               N/A             N/A
    horovod-resnet50-v2-4x8-fluid  SUCCEEDED  MPIJOB   1h        0               N/A             N/A
    horovod-resnet50-v2-4x8-nfs    SUCCEEDED  MPIJOB   2h        0               N/A             N/A

You can use ``arena delete`` to delete the finished training jobs,like:

    $ arena delete tf-standalone-test  mpi-dist
