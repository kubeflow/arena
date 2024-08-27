# Submit a Ray job

Arena supports and simplifies ray job. Arena Ray job is based on [kuberay](https://github.com/ray-project/kuberay). For more information, please refer to [Ray on Kubernetes](https://docs.ray.io/en/latest/cluster/kubernetes/index.html).

## Prerequisites

- k8s deployment  
- deploy the KubeRay operator following the steps from <https://ray-project.github.io/kuberay/deploy/installation/>

## Submit an example Ray job

To submit a ray job, you need to specify the following arguments:

- `--name` Job name. (required)
- `--image` The docker image name of Ray job to use. (required)
- `--worker-replicas` Denotes the number of desired Pods for this worker group (required)

For more information about the arguments to submit a ray job, run the following command:

```shell
arena submit rayjob --help
```

The following command shows how to use arena to submit an example Ray job:

```shell
$ arena submit rayjob \
   --name=rayjob-tk \
   --image=rayproject/ray:2.34.0 \
   --worker-replicas=1 \
   'python -c "import ray; ray.init(); print(ray.cluster_resources())"'
rayjob.ray.io/rayjob-tk created
INFO[0003] The Job rayjob-tk has been submitted successfully 
INFO[0003] You can run `arena get rayjob-tk --type rayjob -n default` to check the job status 
```
Note:
- Currently, submitting a rayjob through arena cannot reuse an existing raycluster and can only create a new raycluster.
- After rayjob succeeds or fails, arena will automatically delete raycluster. You can extend the retention time of raycluster (default 10s) by setting --ttl-seconds-after-finished, or set it not to be deleted by setting --shutdown-after-finished=false.
- You can use --image to specify a unified image for the head pod and worker pod. You can also use --head-image and --worker-image to specify specific images for the head pod and worker pod respectively.

## Get the information of Ray job

```shell
$ arena get rayjob-tk --type rayjob -n default
Name:        rayjob-tk
Status:      SUCCEEDED
Namespace:   default
Priority:    N/A
Trainer:     RAYJOB
Duration:    51s
CreateTime:  2024-09-12 16:29:33
EndTime:     2024-09-12 16:30:24

Instances:
  NAME                                                   STATUS     AGE  IS_CHIEF  GPU(Requested)  NODE
  ----                                                   ------     ---  --------  --------------  ----
  rayjob-tk-raycluster-bt7p4-head-lt4h2                  Running    51s  false     0               10.0.102.157
  rayjob-tk-raycluster-bt7p4-worker-default-group-2p8vz  Running    51s  false     0               10.0.102.157
  rayjob-tk-bbvxm                                        Completed  51s  true      0               10.0.102.156
```

## Get the logs of Ray job

```shell
$ arena logs rayjob-tk --tail 15
2024-09-12 01:30:17,428	INFO cli.py:290 -- Query the logs of the job:
2024-09-12 01:30:17,428	INFO cli.py:292 -- ray job logs rayjob-tk-hbzmd
2024-09-12 01:30:17,428	INFO cli.py:294 -- Query the status of the job:
2024-09-12 01:30:17,428	INFO cli.py:296 -- ray job status rayjob-tk-hbzmd
2024-09-12 01:30:17,429	INFO cli.py:298 -- Request the job to be stopped:
2024-09-12 01:30:17,429	INFO cli.py:300 -- ray job stop rayjob-tk-hbzmd
2024-09-12 01:30:17,444	INFO cli.py:307 -- Tailing logs until the job exits (disable with --no-wait):
2024-09-12 01:30:16,961	INFO job_manager.py:531 -- Runtime env is setting up.
2024-09-12 01:30:18,850	INFO worker.py:1456 -- Using address 10.1.99.58:6379 set in the environment variable RAY_ADDRESS
2024-09-12 01:30:18,850	INFO worker.py:1596 -- Connecting to existing Ray cluster at address: 10.1.99.58:6379...
2024-09-12 01:30:18,862	INFO worker.py:1772 -- Connected to Ray cluster. View the dashboard at 10.1.99.58:8265 
{'CPU': 2.0, 'memory': 5368709120.0, 'object_store_memory': 1194262118.0, 'node:10.1.100.55': 1.0, 'node:10.1.99.58': 1.0, 'node:__internal_head__': 1.0}
2024-09-12 01:30:21,482	SUCC cli.py:63 -- -------------------------------
2024-09-12 01:30:21,482	SUCC cli.py:64 -- Job 'rayjob-tk-hbzmd' succeeded
2024-09-12 01:30:21,482	SUCC cli.py:65 -- -------------------------------
```

## Delete the ray job

```shell
$ arena delete rayjob-tk
INFO[0003] The training job rayjob-tk has been deleted successfully 
```
