
`arena` 允许在训练作业中挂载多个数据卷。下面的示例将数据卷挂载到训练作业中。


1.您需要在 NFS Server 中创建 `/data` 并准备 `mnist data`

```
#mkdir -p /nfs
#mount -t nfs -o vers=4.0 NFS_SERVER_IP://nfs
#mkdir -p /data
#cd /data
#wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/t10k-images-idx3-ubyte.gz
#wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/t10k-labels-idx1-ubyte.gz
#wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/train-images-idx3-ubyte.gz
#wget https://raw.githubusercontent.com/cheyang/tensorflow-sample-code/master/data/train-labels-idx1-ubyte.gz
#cd /
#umount /nfs
```

2\.创建持久卷。将 `NFS_SERVER_IP` 更改为您的相应 NFS Server IP 地址。

```
#cat nfs-pv.yaml
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

3\.创建持久卷声明。 

```
#cat nfs-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: tfdata
  annotations:
    description: "this is the mnist demo"
    owner: Tom
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
       storage: 5Gi
  selector:
    matchLabels:
      tfdata: nas-mnist
#kubectl create -f nfs-pvc.yaml
```

> 注意：建议添加 `description` 和 `owner`

4\.检查数据卷

```
#arena data list 
NAME ACCESSMODE DESCRIPTION OWNER AGE
tfdata ReadWriteMany this is for mnist demo myteam 43d
```

5\.现在，我们可以通过 `arena` 提交分布式训练作业，它会从 github 下载源代码，并将数据卷 `tfdata` 挂载到 `/mnist_data`。

```
#arena submit tf --name=tf-dist-data         \
              --gpus=1 \
              --workers=2 \
              --work-image=tensorflow/tensorflow:1.5.0-devel-gpu \
              --sync-mode=git \
              --sync-source=https://github.com/cheyang/tensorflow-sample-code.git \
              --ps=1 \
              --ps-image=tensorflow/tensorflow:1.5.0-devel \
              --tensorboard \
              --data=tfdata:/mnist_data \
              "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --log_dir /training_logs --data_dir /mnist_data"
```

> `--data` 指定了要挂载到作业的所有任务的数据卷，例如 :。在本例中，数据卷是 `tfdata`，目标目录是 `/mnist_data`。


6\.通过日志，我们发现训练数据提取自 `/mnist_data`，而非直接通过互联网下载得到。

```
#arena logs tf-dist-data
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
