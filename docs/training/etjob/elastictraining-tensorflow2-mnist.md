# Submit a elastic training job(tensorflow)

This guide walks through the steps to submit a elastic training job with horovod.

1\. Build image for training environment

You can use the following image directly.

    registry.cn-hangzhou.aliyuncs.com/ai-samples/horovod:0.20.0-tf2.3.0-torch1.6.0-mxnet1.6.0.post0-py3.7-cuda10.1

In addition, you can also build your own image with the help of this document [elastic-training-sample-image](https://code.aliyun.com/370272561/elastic-training-sample-image).

2\. Submit a elastic training job. Example code from [tensorflow2_mnist_elastic.py](https://github.com/horovod/horovod/blob/master/examples/elastic/tensorflow2_mnist_elastic.py).

    $ arena submit etjob \
        --name=elastic-training \
        --gpus=1 \
        --workers=3 \
        --max-workers=9 \
        --min-workers=1 \
        --image=registry.cn-hangzhou.aliyuncs.com/ai-samples/horovod:0.20.0-tf2.3.0-torch1.6.0-mxnet1.6.0.post0-py3.7-cuda10.1 \
        --working-dir=/examples \
        "horovodrun
        -np \$((\${workers}*\${gpus}))
        --min-np \$((\${minWorkers}*\${gpus}))
        --max-np \$((\${maxWorkers}*\${gpus}))
        --host-discovery-script /usr/local/bin/discover_hosts.sh
        python /examples/elastic/tensorflow2_mnist_elastic.py
        "

    configmap/elastic-training-etjob created
    configmap/elastic-training-etjob labeled
    trainingjob.kai.alibabacloud.com/elastic-training created
    INFO[0000] The Job elastic-training has been submitted successfully
    INFO[0000] You can run `arena get elastic-training --type etjob` to check the job status

3\. List then job.

    $ arena list
    NAME              STATUS   TRAINER  AGE  NODE
    elastic-training  RUNNING  ETJOB    52s  192.168.0.116


4\. Get the job details.

    $ arena get elastic-training

    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 1m

    NAME              STATUS   TRAINER  AGE  INSTANCE                   NODE
    elastic-training  RUNNING  ETJOB    1m   elastic-training-launcher  192.168.0.116
    elastic-training  RUNNING  ETJOB    1m   elastic-training-worker-0  192.168.0.114
    elastic-training  RUNNING  ETJOB    1m   elastic-training-worker-1  192.168.0.116
    elastic-training  RUNNING  ETJOB    1m   elastic-training-worker-2  192.168.0.116

5\. Check logs of the job.

    $ arena logs elastic-training --tail 10

    Tue Sep  8 08:32:50 2020[1]<stdout>:Step #2170	Loss: 0.021992
    Tue Sep  8 08:32:50 2020[0]<stdout>:Step #2180	Loss: 0.000902
    Tue Sep  8 08:32:50 2020[1]<stdout>:Step #2180	Loss: 0.023190
    Tue Sep  8 08:32:50 2020[2]<stdout>:Step #2180	Loss: 0.013149
    Tue Sep  8 08:32:51 2020[0]<stdout>:Step #2190	Loss: 0.029536
    Tue Sep  8 08:32:51 2020[2]<stdout>:Step #2190	Loss: 0.017537
    Tue Sep  8 08:32:51 2020[1]<stdout>:Step #2190	Loss: 0.018273
    Tue Sep  8 08:32:51 2020[2]<stdout>:Step #2200	Loss: 0.038399
    Tue Sep  8 08:32:51 2020[0]<stdout>:Step #2200	Loss: 0.007017
    Tue Sep  8 08:32:51 2020[1]<stdout>:Step #2200	Loss: 0.017495

6\. Scaleout your job. the following sample command Will add one worker into jobs.

    $ arena scaleout etjob --name="elastic-training" --count=1 --timeout=1m

    configmap/elastic-training-1599548177-scaleout created
    configmap/elastic-training-1599548177-scaleout labeled
    scaleout.kai.alibabacloud.com/elastic-training-1599548177 created
    INFO[0000] The scaleout job elastic-training-1599548177 has been submitted successfully


7\. Get your job details. We can see new worker(elastic-training-worker-3) has been "RUNNING".

    $ arena get elastic-training

    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 2m

    NAME              STATUS   TRAINER  AGE  INSTANCE                   NODE
    elastic-training  RUNNING  ETJOB    2m   elastic-training-launcher  192.168.0.116
    elastic-training  RUNNING  ETJOB    2m   elastic-training-worker-0  192.168.0.114
    elastic-training  RUNNING  ETJOB    2m   elastic-training-worker-1  192.168.0.116
    elastic-training  RUNNING  ETJOB    2m   elastic-training-worker-2  192.168.0.116
    elastic-training  RUNNING  ETJOB    2m   elastic-training-worker-3  192.168.0.117


8\. Check logs of the job.

    $ arena logs elastic-training --tail 10

    Tue Sep  8 08:33:33 2020[1]<stdout>:Step #3140	Loss: 0.014412
    Tue Sep  8 08:33:33 2020[0]<stdout>:Step #3140	Loss: 0.004425
    Tue Sep  8 08:33:33 2020[3]<stdout>:Step #3150	Loss: 0.000513
    Tue Sep  8 08:33:33 2020[2]<stdout>:Step #3150	Loss: 0.062282
    Tue Sep  8 08:33:33 2020[1]<stdout>:Step #3150	Loss: 0.020650
    Tue Sep  8 08:33:33 2020[0]<stdout>:Step #3150	Loss: 0.008056
    Tue Sep  8 08:33:34 2020[3]<stdout>:Step #3160	Loss: 0.002170
    Tue Sep  8 08:33:34 2020[2]<stdout>:Step #3160	Loss: 0.009676
    Tue Sep  8 08:33:34 2020[1]<stdout>:Step #3160	Loss: 0.051425
    Tue Sep  8 08:33:34 2020[0]<stdout>:Step #3160	Loss: 0.023769


9\. Scalein your job. Will remove one worker from current jobs.

    $ arena scalein etjob --name="elastic-training" --count=1 --timeout=1m

    configmap/elastic-training-1599554041-scalein created
    configmap/elastic-training-1599554041-scalein labeled
    scalein.kai.alibabacloud.com/elastic-training-1599554041 created
    INFO[0000] The scalein job elastic-training-1599554041 has been submitted successfully


10\. Get your job details. We can see that ``elastic-training-worker-3`` has been removed.

    $ arena get elastic-training

    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 3m

    NAME              STATUS   TRAINER  AGE  INSTANCE                   NODE
    elastic-training  RUNNING  ETJOB    3m   elastic-training-launcher  192.168.0.116
    elastic-training  RUNNING  ETJOB    3m   elastic-training-worker-0  192.168.0.114
    elastic-training  RUNNING  ETJOB    3m   elastic-training-worker-1  192.168.0.116
    elastic-training  RUNNING  ETJOB    3m   elastic-training-worker-2  192.168.0.116
    ```

11\. Check logs of the job.

    $ arena logs elastic-training --tail 10

    Tue Sep  8 08:34:43 2020[0]<stdout>:Step #5210	Loss: 0.005627
    Tue Sep  8 08:34:43 2020[2]<stdout>:Step #5220	Loss: 0.002142
    Tue Sep  8 08:34:43 2020[1]<stdout>:Step #5220	Loss: 0.002978
    Tue Sep  8 08:34:43 2020[0]<stdout>:Step #5220	Loss: 0.011404
    Tue Sep  8 08:34:44 2020[2]<stdout>:Step #5230	Loss: 0.000689
    Tue Sep  8 08:34:44 2020[1]<stdout>:Step #5230	Loss: 0.024597
    Tue Sep  8 08:34:44 2020[0]<stdout>:Step #5230	Loss: 0.040936
    Tue Sep  8 08:34:44 2020[0]<stdout>:Step #5240	Loss: 0.000125
    Tue Sep  8 08:34:44 2020[2]<stdout>:Step #5240	Loss: 0.026498
    Tue Sep  8 08:34:44 2020[1]<stdout>:Step #5240	Loss: 0.000308
