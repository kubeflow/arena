1.1\ prepare kubeconfig

```
mkdir -p /root/.kube
scp root@master_ip:/etc/kubernetes/admin.conf /root/.kube/config
```

1.2\ Install helm

```
wget http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou-internal.aliyuncs.com/public/pkg/helm/helm-v2.8.2-linux-amd64.tar.gz
helm init --upgrade -i registry-vpc.cn-zhangjiakou.aliyuncs.com/acs/tiller:v2.8.2 --skip-refresh
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
```


```
scp root@master_ip:/usr/local/bin/helm /usr/local/bin/helm
```

1.3\ Place chart

```
mkdir /charts
cd /charts
curl -O http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou-internal.aliyuncs.com/public/charts/training-0.1.0.tgz
tar -xvf training-0.1.0.tgz
curl -O http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou-internal.aliyuncs.com/public/charts/horovod-0.1.0.tgz
tar -xvf horovod-0.1.0.tgz
```


1.4\ Place arena

```

curl -o /usr/local/bin/arena http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou-internal.aliyuncs.com/public/charts/arena
chmod u+x /usr/local/bin/arena
```

1.5\ Create job mon role

```
wget http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou.aliyuncs.com/public/charts/jobmon-role.yaml
kubectl create -f jobmon-role.yaml
```

1.6\ install dashboard

```
# login master0
kubectl delete -f pkg/kubernetes/1.9.3/module/kubernetes-dashboard-with-client-cert.yml
wget http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou.aliyuncs.com/public/charts/dashboard.yaml
kubectl create -f dashboard.yaml
```