This guide walks through the steps required to deploy and serve a TensorFlow model using Kubernetes (K8s) and Istio.

1. Setup

Before using `Arena` for TensorFlow serving, we need to setup the environment including Kubernetes cluster and Istio.

Make sure that your Kubernetes cluster is running.

Follow the Istio [doc](https://istio.io/docs/setup/kubernetes/quick-start/#installation-steps) to install Istio. After the installation, you should see services `istio-pilot` and `istio-mixer` in namespace `istio-system`.

Istio by default [denies egress traffic](https://istio.io/docs/tasks/traffic-management/egress.html). Since TensorFlow serving component might need to read model files from outside, we need some cloud-specific [setting](https://istio.io/docs/tasks/traffic-management/egress.html#calling-external-services-directly). 

2\. Create Persistent Volume for Model Files

Create /tfmodel in the NFS Server, and prepare mnist models by following the command:

```
mount -t nfs -o vers=4.0 NFS_SERVER_IP:/ /tfmodel/
wget https://github.com/osswangxining/tensorflow-sample-code/raw/master/models/tensorflow/mnist.tar.gz
tar xvf mnist.tar.gz
``` 

Then create Persistent Volume and Persistent Volume Claim by following the command (using NFS as sample):

Persistent Volume:
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

Persistent Volume Claim:

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

Check the data volume:
```
arena data list
NAME    ACCESSMODE     DESCRIPTION                OWNER   AGE
tfmodel  ReadWriteMany this is tfmodel for mnist  tester  31s
```


3\. Disable Istio for Tensorflow serving

You can deploy and serve a Tensorflow model without Istio enablement. 

Submit tensorflow serving job to deploy and serve machine learning models using the following command.

```
arena serve tensorflow [flags]

options:
      --command string           the command will inject to container's command.
      --cpu string               the request cpu of each replica to run the serve.
  -d, --data stringArray         specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>
      --enableIstio              enable Istio for serving or not (disable Istio by default)
  -e, --envs stringArray         the environment variables
      --gpus int                 the limit GPU count of each replica to run the serve.
  -h, --help                     help for tensorflow
      --image string             the docker image name of serve job, default image is tensorflow/serving:latest (default "tensorflow/serving:latest")
      --memory string            the request memory of each replica to run the serve.
      --modelConfigFile string   Corresponding with --model_config_file in tensorflow serving
      --modelName string         the model name for serving
      --modelPath string         the model path for serving in the container
      --port int                 the port of tensorflow gRPC listening port (default 8500)
      --replicas int             the replicas number of the serve job. (default 1)
      --restfulPort int          the port of tensorflow RESTful listening port (default 8501)
      --servingName string       the serving name
      --servingVersion string    the serving version
      --versionPolicy string     support latest, latest:N, specific:N, all

Options inherited from parent commands
      --arenaNamespace string   The namespace of arena system service, like TFJob (default "arena-system")
      --config string           Path to a kube config. Only required if out-of-cluster
      --loglevel string         Set the logging level. One of: debug|info|warn|error (default "info")
      --namespace string        the namespace of the job (default "default")
      --pprof                   enable cpu profile      
```

For example, you can submit a Tensorflow model with specific version policy as below.

```
arena serve tensorflow --servingName=mymnist --modelName=mnist --image=tensorflow/serving:latest  --data=tfmodel:/tfmodel --modelPath=/tfmodel/mnist --versionPolicy=specific:1  --loglevel=debug
```

Once this command is triggered, one Kubernetes service will be created to provide the exposed gRPC and RESTful APIs.


4\. Enable Istio for Tensorflow serving

If you need to enable Istio for Tensorflow serving,  you can append the parameter `--enableIstio` into the command above (disable Istio by default).

For example,  you can submit a Tensorflow model with Istio enablement as below.

```
# arena serve tensorflow --enableIstio --servingName=mymnist --servingVersion=v1 --modelName=mnist  --data=myoss1pvc:/data2 --modelPath=/data2/models/mnist --versionPolicy=specific:1 

NAME:   mymnist-v1
LAST DEPLOYED: Wed Sep 26 17:28:13 2018
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/ConfigMap
NAME                              DATA  AGE
mymnist-v1-tensorflow-serving-cm  1     1s

==> v1/Service
NAME     TYPE       CLUSTER-IP     EXTERNAL-IP  PORT(S)            AGE
mymnist  ClusterIP  172.19.12.176  <none>       8500/TCP,8501/TCP  1s

==> v1beta1/Deployment
NAME                           DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
mymnist-v1-tensorflow-serving  1        1        1           0          1s

==> v1alpha3/DestinationRule
NAME     AGE
mymnist  1s

==> v1alpha3/VirtualService
mymnist  1s

==> v1/Pod(related)
NAME                                            READY  STATUS    RESTARTS  AGE
mymnist-v1-tensorflow-serving-757b669bbb-5vsmf  0/2    Init:0/1  0         1s


NOTES:
Getting Started:

**** NOTE: It may take a few minutes for the LoadBalancer IP to be available.                                                                 ****
****       You can watch the status of by running 'kubectl get svc --namespace default -w mymnist-v1-tensorflow-serving' ****
  export TF_SERVING_SERVICE_IP=$(kubectl get svc --namespace default mymnist-v1-tensorflow-serving -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo docker run -it --rm cheyang/tf-mnist:grpcio_upgraded /serving/bazel-bin/tensorflow_serving/example/mnist_client --num_tests=1000 --server=$TF_SERVING_SERVICE_IP:9090`

```

5\. List all the serving

You can use the following command to list all the serving jobs.

```
# arena serve list
  NAME        VERSION  STATUS
  mymnist-v1  v1       DEPLOYED
```

6\. Adjust traffic routing dynamically for tfserving jobs
   
Deploy one new version of Tensorflow model with Istio enablement:
```
# arena serve tensorflow --enableIstio --servingName=mymnist --servingVersion=v2 --modelName=mnist  --data=myoss1pvc:/data2 --modelPath=/data2/models/mnist 
```

Then you can adjust traffic routing dynamically for all these two versions of tfserving jobs.
            
```
# arena serve traffic-router-split --servingName=mymnist  --servingVersions=v1,v2 --weights=50,50
```

7\. Run the test

Start the `sleep` service so you can use `curl` to provide load:

```
# cat <<EOF | istioctl kube-inject -f - | kubectl create -f -
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

Find the name of  `sleep` pod and enter into this pod, for example:

```
# kubectl exec -it sleep-5dd9955c58-km59h -c sleep bash
```

In this container, use `curl` to call the exposed Tensorflow serving API:

```
# curl -X POST   http://mymnist:8501/v1/models/mnist:predict    -d '{"signature_name":"predict","instances":[{"sepal_length":[6.8],"sepal_width":[3.2],"petal_length":[5.9],"petal_width":[2.3]}]}' 
```

8\. Delete one serving job

You can use the following command to delete a serving job and its associated pods
                                     
```
# arena serve delete mymnist-v1
release "mymnist-v1" deleted
```
