This guide walks through the steps to submit a elastic training job with horovod.

1. Build image for training environment
You can use the [registry.cn-hangzhou.aliyuncs.com/ai-samples/horovod:0.20.0-tf2.3.0-torch1.6.0-mxnet1.6.0.post0-py3.7-cuda10.1]() image directly.
In addition, you can also build your own image with the help of this document [elastic-training-sample-image](https://code.aliyun.com/370272561/elastic-training-sample-image).

2. Submit a elastic training job. Example code from [pytorch_synthetic_benchmark_elastic.py](https://github.com/horovod/horovod/blob/master/examples/elastic/pytorch_synthetic_benchmark_elastic.py)
    ```shell script
    arena submit etjob \
            --name=elastic-training-synthetic \
            --gpus=1 \
            --workers=3 \
            --max-workers=9 \
            --min-workers=1 \
            --image=registry.cn-hangzhou.aliyuncs.com/ai-samples/horovod:0.20.0-tf2.3.0-torch1.6.0-mxnet1.6.0.post0-py3.7-cuda10.1 \
            --working-dir=/examples \
            "horovodrun
            --verbose
            --log-level=DEBUG
            -np \$((\${workers}*\${gpus}))
            --min-np \$((\${minWorkers}*\${gpus}))
            --max-np \$((\${maxWorkers}*\${gpus}))
            --start-timeout 100
            --elastic-timeout 1000
            --host-discovery-script /usr/local/bin/discover_hosts.sh
            python /examples/elastic/pytorch_synthetic_benchmark_elastic.py
            --num-iters=10000
            --num-warmup-batches=0"
    ```
    Output:
    ```
    configmap/elastic-training-synthetic-etjob created
    configmap/elastic-training-synthetic-etjob labeled
    trainingjob.kai.alibabacloud.com/elastic-training-synthetic created
    INFO[0000] The Job elastic-training-synthetic has been submitted successfully
    INFO[0000] You can run `arena get elastic-training-synthetic --type etjob` to check the job status
    ```

3. List your job.
    ```shell script
    arena list
    ```
    Output:
    ```
    NAME                        STATUS     TRAINER  AGE  NODE
    elastic-training-synthetic  RUNNING    ETJOB    2m   192.168.0.112
    ```

4. Get your job details.
    ```shell script
    arena get elastic-training-synthetic
    ```
    Output:
    ```
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 3m

    NAME                        STATUS   TRAINER  AGE  INSTANCE                             NODE
    elastic-training-synthetic  RUNNING  ETJOB    3m   elastic-training-synthetic-launcher  192.168.0.112
    elastic-training-synthetic  RUNNING  ETJOB    3m   elastic-training-synthetic-worker-0  192.168.0.116
    elastic-training-synthetic  RUNNING  ETJOB    3m   elastic-training-synthetic-worker-1  192.168.0.117
    elastic-training-synthetic  RUNNING  ETJOB    3m   elastic-training-synthetic-worker-2  192.168.0.116
    ```

5. Check logs
    ```shell script
    arena logs elastic-training-synthetic --tail 10
    ```
    Output:
    ```
    Tue Sep  8 09:24:20 2020[0]<stdout>:Iter #54: 95.3 img/sec per GPU
    Tue Sep  8 09:24:23 2020[0]<stdout>:Iter #55: 95.3 img/sec per GPU
    Tue Sep  8 09:24:27 2020[0]<stdout>:Iter #56: 94.6 img/sec per GPU
    Tue Sep  8 09:24:30 2020[0]<stdout>:Iter #57: 97.1 img/sec per GPU
    Tue Sep  8 09:24:33 2020[0]<stdout>:Iter #58: 99.7 img/sec per GPU
    Tue Sep  8 09:24:36 2020[0]<stdout>:Iter #59: 99.8 img/sec per GPU
    Tue Sep  8 09:24:40 2020[0]<stdout>:Iter #60: 98.0 img/sec per GPU
    Tue Sep  8 09:24:43 2020[0]<stdout>:Iter #61: 97.1 img/sec per GPU
    Tue Sep  8 09:24:46 2020[0]<stdout>:Iter #62: 96.1 img/sec per GPU
    Tue Sep  8 09:24:50 2020[0]<stdout>:Iter #63: 100.4 img/sec per GPU
    ```


6. Scaleout your job. Will add one worker into jobs.
    ```shell script
    arena scaleout etjob --name="elastic-training-synthetic" --count=1 --timeout=1m
    ```
    Output:
    ```
    configmap/elastic-training-synthetic-1599557124-scaleout created
    configmap/elastic-training-synthetic-1599557124-scaleout labeled
    scaleout.kai.alibabacloud.com/elastic-training-synthetic-1599557124 created
    INFO[0000] The scaleout job elastic-training-synthetic-1599557124 has been submitted successfully
    ```

7. Get your job details. We can see new worker(elastic-training-synthetic-worker-3) has been "RUNNING".
    ```shell script
    arena get elastic-training-synthetic
    ```
    Output:
    ```
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 5m

    NAME                        STATUS   TRAINER  AGE  INSTANCE                             NODE
    elastic-training-synthetic  RUNNING  ETJOB    5m   elastic-training-synthetic-launcher  192.168.0.112
    elastic-training-synthetic  RUNNING  ETJOB    5m   elastic-training-synthetic-worker-0  192.168.0.116
    elastic-training-synthetic  RUNNING  ETJOB    5m   elastic-training-synthetic-worker-1  192.168.0.117
    elastic-training-synthetic  RUNNING  ETJOB    5m   elastic-training-synthetic-worker-2  192.168.0.116
    elastic-training-synthetic  RUNNING  ETJOB    5m   elastic-training-synthetic-worker-3  192.168.0.112
    ```

8. Check logs.
    ```shell script
    arena logs elastic-training-synthetic --tail 10
    ```
    Output:
    ```
    Tue Sep  8 09:26:03 2020[0]<stdout>:Iter #76: 65.0 img/sec per GPU
    Tue Sep  8 09:26:08 2020[0]<stdout>:Iter #77: 64.0 img/sec per GPU
    Tue Sep  8 09:26:13 2020[0]<stdout>:Iter #78: 65.4 img/sec per GPU
    Tue Sep  8 09:26:18 2020[0]<stdout>:Iter #79: 64.4 img/sec per GPU
    Tue Sep  8 09:26:23 2020[0]<stdout>:Iter #80: 62.9 img/sec per GPU
    Tue Sep  8 09:26:28 2020[0]<stdout>:Iter #81: 64.0 img/sec per GPU
    Tue Sep  8 09:26:33 2020[0]<stdout>:Iter #82: 64.4 img/sec per GPU
    Tue Sep  8 09:26:38 2020[0]<stdout>:Iter #83: 64.9 img/sec per GPU
    Tue Sep  8 09:26:43 2020[0]<stdout>:Iter #84: 62.7 img/sec per GPU
    Tue Sep  8 09:26:48 2020[0]<stdout>:Iter #85: 64.2 img/sec per GPU
    ```

9. Scalein your job. Will remove one worker from current jobs.
    ```shell script
    arena scalein etjob --name="elastic-training-synthetic" --count=1 --timeout=1m
    ```
    Output:
    ```
    configmap/elastic-training-synthetic-1599557271-scalein created
    configmap/elastic-training-synthetic-1599557271-scalein labeled
    scalein.kai.alibabacloud.com/elastic-training-synthetic-1599557271 created
    INFO[0000] The scalein job elastic-training-synthetic-1599557271 has been submitted successfully
    ```

10. Get your job details. We can see that `elastic-training-synthetic-worker-3` has been removed.
    ```shell script
    arena get elastic-training-synthetic
    ```
    Output:
    ```
    STATUS: RUNNING
    NAMESPACE: default
    PRIORITY: N/A
    TRAINING DURATION: 7m

    NAME                        STATUS   TRAINER  AGE  INSTANCE                             NODE
    elastic-training-synthetic  RUNNING  ETJOB    7m   elastic-training-synthetic-launcher  192.168.0.112
    elastic-training-synthetic  RUNNING  ETJOB    7m   elastic-training-synthetic-worker-0  192.168.0.116
    elastic-training-synthetic  RUNNING  ETJOB    7m   elastic-training-synthetic-worker-1  192.168.0.117
    elastic-training-synthetic  RUNNING  ETJOB    7m   elastic-training-synthetic-worker-2  192.168.0.116
    ```

11. Check logs.
    ```shell script
    arena logs elastic-training-synthetic --tail 10
    ```
    Output:
    ```
    DEBUG:root:host elastic-training-synthetic-worker-3 has been blacklisted, ignoring exit from local_rank=0
    Process 3 exit with status code 134.
    Tue Sep  8 09:27:56 2020[0]<stdout>:Iter #97: 96.0 img/sec per GPU
    Tue Sep  8 09:28:00 2020[0]<stdout>:Iter #98: 95.4 img/sec per GPU
    Tue Sep  8 09:28:03 2020[0]<stdout>:Iter #99: 96.9 img/sec per GPU
    Tue Sep  8 09:28:06 2020[0]<stdout>:Iter #100: 97.2 img/sec per GPU
    Tue Sep  8 09:28:10 2020[0]<stdout>:Iter #101: 98.5 img/sec per GPU
    Tue Sep  8 09:28:13 2020[0]<stdout>:Iter #102: 95.8 img/sec per GPU
    Tue Sep  8 09:28:16 2020[0]<stdout>:Iter #103: 97.3 img/sec per GPU
    Tue Sep  8 09:28:20 2020[0]<stdout>:Iter #104: 97.3 img/sec per GPU
    Tue Sep  8 09:28:23 2020[0]<stdout>:Iter #105: 98.9 img/sec per GPU
    ```