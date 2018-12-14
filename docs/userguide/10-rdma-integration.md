Arena supports RDMA For distributed Training. We can allocate RDMA device for worker jobs by adding parameter `--rdma`

1. Deploy rdma device plugin

```
# Deploy RDMA device plugin
kubectl create -f kubernetes-artifacts/rdma/rdma-config.yaml
kubectl create -f kubernetes-artifacts/rdma/device-plugin.yaml
```

2\. Label your node with infiniband device

```
# Label RDMA NODE
kubectl label node <your node> accelerator/rdma=true
```

```
# Check Device plugin status
kubectl -n kube-system get ds
NAME                       DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                     AGE
rdma-sriov-dp-ds           1         1         1       1            1           accelerator/rdma=true      46d
```

3\. Enable arena RDMA config

```
find /charts/ -name values.yaml | xargs sed -i "/enableRDMA/s/false/true/g"
```

4\. Submit a Tensorflow training job using RDMA

```
# arena submit tf --name=tf-dist-git \
              --rdma \
              --gpus=1 \
              --workers=2 \
              --workerImage=tensorflow/tensorflow:1.5.0-devel-gpu \
              --syncMode=git \
              --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
              --ps=1 \
              --psImage=tensorflow/tensorflow:1.5.0-devel \
              --tensorboard \
              "python code/tensorflow-sample-code/tfjob/docker/v1alpha2/distributed-mnist/main.py --logdir /training_logs"

NAME:   tf-dist-git
LAST DEPLOYED: Fri Dec 14 18:47:28 2018
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1alpha2/TFJob
NAME               AGE
tf-dist-git-tfjob  0s

==> v1/Pod(related)
NAME                                READY  STATUS   RESTARTS  AGE
tf-dist-git-tfjob-54c6cd95d6-hc5n9  0/1    Pending  0         0s

==> v1/Service
NAME               TYPE      CLUSTER-IP   EXTERNAL-IP  PORT(S)         AGE
tf-dist-git-tfjob  NodePort  172.19.9.80  <none>       6006:32339/TCP  0s

==> v1beta1/Deployment
NAME               DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
tf-dist-git-tfjob  1        1        1           0          0s
```