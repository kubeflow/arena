
`arena` allows to mount multiple data volumes into the training jobs. There is an example that mounts `data volume` into the training job.


1. You need to create `/data` in the NFS Server, and prepare `mnist data`

```
# mkdir -p /nfs
# mount -t nfs -o vers=4.0 NFS_SERVER_IP:/ /nfs
# mkdir -p /data
# cd /data
# wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/t10k-images-idx3-ubyte.gz
# wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/t10k-labels-idx1-ubyte.gz
# wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/train-images-idx3-ubyte.gz
# wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/train-labels-idx1-ubyte.gz
# cd /
# umount /nfs
```

2\. Create Persistent Volume. Moidfy `NFS_SERVER_IP` to yours.

```
# cat nfs-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: tfdata
  labels:
    tfdata: nas-mnist
spec:
  persistentVolumeReclaimPolicy: Retain
  capacity:
    storage: 10Gi
  accessModes:
  - ReadWriteMany
  nfs:
    server: NFS_SERVER_IP
    path: "/data"
    
 # kubectl create -f nfs-pv.yaml
```

3\. Create Persistent Volume Claim. 

```
# cat nfs-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: tfdata
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
       storage: 5Gi
  selector:
    matchLabels:
      tfdata: nas-mnist
# kubectl create -f nfs-pvc.yaml
```

4\. Check PV and PVC

```
# kubectl get pv,pvc
NAME                      CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM            STORAGECLASS   REASON    AGE
persistentvolume/tfdata   10Gi       RWX            Retain           Bound     default/tfdata                            38s

NAME                           STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/tfdata   Bound     tfdata    10Gi       RWX                           13s
```

5\. Now we can submit a distributed training job with `arena`, it will download the source code from github and point the data volume to PVC `tfdata`.

```
# arena submit tf --name=tf-dist-data         \
              --gpus=1              \
              --workers=2              \
              --workerImage=tensorflow/tensorflow:1.5.0-devel-gpu  \
              --syncMode=git \
              --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
              --ps=1              \
              --psImage=tensorflow/tensorflow:1.5.0-devel   \
              --tensorboard \
              --data=tfdata:/mnist_data \
              "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --logdir /training_logs --data_dir /mnist_data"
```

> `--data` specifies the data volume to mount to all the tasks of the job, like <name_of_datasource>:<mount_point_on_job>. In this example, the PVC is `tfdata`, and the target directory is `/mnist_data`.


6\. From the logs, we find that the training data is extracted from `/mnist_data` instead of downloading from internet directly.

```
# arena logs tf-dist-data
...
Extracting /mnist_data/train-images-idx3-ubyte.gz
Extracting /mnist_data/train-labels-idx1-ubyte.gz
Extracting /mnist_data/t10k-images-idx3-ubyte.gz
Extracting /mnist_data/t10k-labels-idx1-ubyte.gz
...
Accuracy at step 960: 0.9753
Accuracy at step 970: 0.9739
Accuracy at step 980: 0.9756
Accuracy at step 990: 0.9777
Adding run metadata for 999
```