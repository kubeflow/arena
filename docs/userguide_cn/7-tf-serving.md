本指南详细介绍了使用 Kubernetes (K8s) 和 Istio 部署和提供 TensorFlow 模型时所需的步骤。

1.设置

在使用 `Arena` 传送 TensorFlow 之前，我们需要设置环境，包括 Kubernetes 集群和 Istio。

确保您的 Kubernetes 集群处于运行状态。

按照 Istio [文档](https://istio.io/docs/setup/kubernetes/quick-start/#installation-steps) 安装 Istio。安装完成之后，您应该会在命名空间 `istio-system` 内看到 `istio-pilot` 和 `istio-mixer` 服务。

Istio 默认 [拒绝出站数据流量](https://istio.io/docs/tasks/traffic-management/egress.html)。由于 TensorFlow Serving组件可能需要从外部都模型文件，因此我们需要某些特定于云的 [设置](https://istio.io/docs/tasks/traffic-management/egress.html#calling-external-services-directly)。 

2\.为模型文件创建持久卷

在 NFS Server 中创建 /tfmodel，执行如下命令准备 mnist 模型：

```
mount -t nfs -o vers=4.0 NFS_SERVER_IP://tfmodel/
wget https://github.com/osswangxining/tensorflow-sample-code/raw/master/models/tensorflow/mnist.tar.gz
tar xvf mnist.tar.gz
``` 

然后执行如下命令（以 NFS 为例）创建持久卷和持久卷声明：

持久卷：
```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: tfmodel
  labels:
    tfmodel: nas-mnist
spec:
  persistentVolumeReclaimPolicy: Retain
  capacity:
    storage: 10Gi
  accessModes:
  - ReadWriteMany
  nfs:
    server: NFS_SERVER_IP
    path: "/tfmodel"
```

持久卷声明：

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: tfmodel
  annotations:
    description: "this is tfmodel for mnist"
    owner: tester
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
       storage: 5Gi
  selector:
    matchLabels:
      tfmodel: nas-mnist
```

检查数据卷：
```
arena data list
NAME ACCESSMODE DESCRIPTION OWNER AGE
tfmodel ReadWriteMany this is tfmodel for mnist tester 31s
```


3\.为 Tensorflow Serving禁用 Istio

您可以在不启用 Istio 的情况下部署并传送 Tensorflow 模型。 

执行以下命令，提交 tensorflow Serving作业，部署并传送机器学习模型。

```
arena serve tensorflow [flags]

options:
      --command string the command will inject to container's command.
      --cpu string the request cpu of each replica to run the serve.
  -d, --data stringArray specify the trained models datasource to mount for serving, like :
      --enableIstio enable Istio for serving or not (disable Istio by default)
  -e, --envs stringArray the environment variables
      --gpus int the limit GPU count of each replica to run the serve.
  -h, --help help for tensorflow
      --image string the docker image name of serve job, default image is tensorflow/serving:latest (default "tensorflow/serving:latest")
      --memory string the request memory of each replica to run the serve.
      --modelConfigFile string Corresponding with --model_config_file in tensorflow serving
      --modelName string the model name for serving
      --modelPath string the model path for serving in the container
      --port int the port of tensorflow gRPC listening port (default 8500)
      --replicas int the replicas number of the serve job.(default 1)
      --restfulPort int the port of tensorflow RESTful listening port (default 8501)
      --servingName string the serving name
      --servingVersion string the serving version
      --versionPolicy string support latest, latest:N, specific:N, all

继承自父命令的选项
      --arenaNamespace string The namespace of arena system service, like TFJob (default "arena-system")
      --config string Path to a kube config.Only required if out-of-cluster
      --loglevel string Set the logging level.One of: debug|info|warn|error (default "info")
      --namespace string the namespace of the job (default "default")
      --pprof enable cpu profile      
```

例如，您可以使用如下所示的特定版本策略提交 Tensorflow 模型。

```
arena serve tensorflow --servingName=mymnist --modelName=mnist --image=tensorflow/serving:latest --data=tfmodel:/tfmodel --modelPath=/tfmodel/mnist --versionPolicy=specific:1 --loglevel=debug
```

触发该命令之后，系统将创建相应 Kubernetes 服务，以提供公开的 gRPC 和 RESTful API。


4\.为 Tensorflow Serving启用 Istio

如果您需要为 Tensorflow Serving启用 Istio，则可以在上述命令中附上参数 `--enableIstio`（默认禁用 Istio）。

例如，您可以在提交 Tensorflow 模型的同时启用 Istio，如下所示。

```
#arena serve tensorflow --enableIstio --servingName=mymnist --servingVersion=v1 --modelName=mnist  --data=myoss1pvc:/data2 --modelPath=/data2/models/mnist --versionPolicy=specific:1 

NAME:   mymnist-v1
LAST DEPLOYED: Wed Sep 26 17:28:13 2018
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/ConfigMap
NAME DATA AGE
mymnist-v1-tensorflow-serving-cm 1 1s

==> v1/Service
NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
mymnist ClusterIP 172.19.12.176  8500/TCP,8501/TCP 1s

==> v1beta1/Deployment
NAME DESIRED CURRENT UP-TO-DATE AVAILABLE AGE
mymnist-v1-tensorflow-serving 1 1 1 0 1s

==> v1alpha3/DestinationRule
NAME AGE
mymnist 1s

==> v1alpha3/VirtualService
mymnist 1s

==> v1/Pod(related)
NAME READY STATUS RESTARTS AGE
mymnist-v1-tensorflow-serving-757b669bbb-5vsmf 0/2 Init:0/1 0 1s


NOTES:
Getting Started:

**** NOTE: It may take a few minutes for the LoadBalancer IP to be available.                                                                 ****
**** You can watch the status of by running 'kubectl get svc --namespace default -w mymnist-v1-tensorflow-serving' ****
  export TF_SERVING_SERVICE_IP=$(kubectl get svc --namespace default mymnist-v1-tensorflow-serving -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo docker run -it --rm cheyang/tf-mnist:grpcio_upgraded /serving/bazel-bin/tensorflow_serving/example/mnist_client --num_tests=1000 --server=$TF_SERVING_SERVICE_IP:9090`

```

5\.列出所有传送作业

您可以使用如下命令列出所有传送作业。

```
#arena serve list
  NAME VERSION STATUS
  mymnist-v1 v1 DEPLOYED
```

6\.为 tfserving 作业动态调整流量路由
   
部署一个新版本的 Tensorflow 模型，同时启用 Istio：
```
#arena serve tensorflow --enableIstio --servingName=mymnist --servingVersion=v2 --modelName=mnist  --data=myoss1pvc:/data2 --modelPath=/data2/models/mnist 
```

随后您可以为全部这两个版本的 tfserving 作业动态调整流量路由。
            
```
#arena serve traffic-router-split --servingName=mymnist  --servingVersions=v1,v2 --weights=50,50
```

7\.运行测试

启动 `sleep` 服务，以便使用 `curl` 提供负载：

```
#cat <<EOF | istioctl kube-inject -f - | kubectl create -f -
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: sleep
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: sleep
    spec:
      containers:
      - name: sleep
        image: tutum/curl
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: IfNotPresent
EOF
```

找到 `sleep` pod 的名称并进入此 pod，例如：

```
#kubectl exec -it sleep-5dd9955c58-km59h -c sleep bash
```

在此容器内，使用 `curl` 调用公开的 Tensorflow Serving API：

```
#curl -X POST   http://mymnist:8501/v1/models/mnist:predict    -d '{"signature_name":"predict","instances":[{"sepal_length":[6.8],"sepal_width":[3.2],"petal_length":[5.9],"petal_width":[2.3]}]}' 
```

8\.删除一个传送作业

您可以使用如下命令来删除传送作业及其相关的 pod
                                     
```
#arena serve delete mymnist-v1
release "mymnist-v1" deleted
```

