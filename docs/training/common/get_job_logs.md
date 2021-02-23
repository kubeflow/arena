# Get training job logs

You can use ``arena logs`` to get the training job logs, we will introduce some usages for you.

1\. ``arena logs`` can help you to get the training job logs.

    $ arena logs tf-standalone-test

    ...
    Accuracy at step 9900: 0.9822
    Save checkpoint at step 9900: 0.9822
    Accuracy at step 9910: 0.982
    Accuracy at step 9920: 0.9819
    Accuracy at step 9930: 0.9821
    Accuracy at step 9940: 0.9827
    Accuracy at step 9950: 0.982
    Accuracy at step 9960: 0.9814
    Accuracy at step 9970: 0.9818
    Accuracy at step 9980: 0.9823
    Accuracy at step 9990: 0.9823
    Adding run metadata for 9999
    Total Train-accuracy=0.9823

2\. If you want to get the last n lines of the training job logs, you can use ``-t n``(or ``--tail n``), the following command will display the last 10 lines of the training job.

    $ arena logs tf-standalone-test -n 10

    Accuracy at step 9920: 0.9819
    Accuracy at step 9930: 0.9821
    Accuracy at step 9940: 0.9827
    Accuracy at step 9950: 0.982
    Accuracy at step 9960: 0.9814
    Accuracy at step 9970: 0.9818
    Accuracy at step 9980: 0.9823
    Accuracy at step 9990: 0.9823
    Adding run metadata for 9999
    Total Train-accuracy=0.9823

3\. If you want to get the logs of target training job instance, you can use ``-i <INSTANCE_NAME>``(or ``--instance <INSTANCE_NAME>``).

Get the instance of training job from ``arena get`` command.

    $ arena get tf-standalone-test

    Name:        tf-standalone-test
    Status:      SUCCEEDED
    Namespace:   default
    Priority:    N/A
    Trainer:     TFJOB
    Duration:    11m

    Instances:
    NAME                        STATUS     AGE  IS_CHIEF  GPU(Requested)  NODE
    ----                        ------     ---  --------  --------------  ----
    tf-standalone-test-chief-0  Completed  11m  true      1               cn-beijing.192.168.1.112

    Tensorboard:
    Your tensorboard will be available on:
    http://192.168.1.101:31869

Get the logs of training job instance.

    $ arena logs tf-standalone-test -i tf-standalone-test-chief-0 -n 10

    Accuracy at step 9920: 0.9819
    Accuracy at step 9930: 0.9821
    Accuracy at step 9940: 0.9827
    Accuracy at step 9950: 0.982
    Accuracy at step 9960: 0.9814
    Accuracy at step 9970: 0.9818
    Accuracy at step 9980: 0.9823
    Accuracy at step 9990: 0.9823
    Adding run metadata for 9999
    Total Train-accuracy=0.9823


4\. If you want to get a training job logs in a duration, for example, ``--since 5m`` represents that getting the training job logs from five minutes ago to the present.

    $ arena logs tf-standalone-test --since 5m

5\. If you want to real-time display the training job logs, ``-f`` is required.

    $ arena logs tf-standalone-test -f