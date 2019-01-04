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
# arena submit mpi --name=mpi-dist              \
              --rdma \
              --gpus=1              \
              --workers=2              \
              --image=uber/horovod:0.13.11-tf1.10.0-torch0.4.0-py3.5  \
              --env=GIT_SYNC_BRANCH=cnn_tf_v1.9_compatible \
              --syncMode=git \
              --syncSource=https://github.com/tensorflow/benchmarks.git \
              --tensorboard \
              "mpirun python code/benchmarks/scripts/tf_cnn_benchmarks/tf_cnn_benchmarks.py --model resnet101 --batch_size 64     --variable_update horovod --train_dir=/training_logs --summary_verbosity=3
              --save_summaries_steps=10"
```
