This guide walks through the steps required to deploy and serve a TensorFlow model using Kubernetes (K8s) and Istio (if you want to experience the advanced features such as version based traffic splitting).

1. Setup

Before using `Arena` for TensorFlow serving, we need to setup the environment including Kubernetes cluster and Istio (optional).

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


3\. Tensorflow serving without Istio enabled

You can deploy and serve a Tensorflow model without Istio enabled. 

Submit tensorflow serving job to deploy and serve machine learning models using the following command.

```
Usage:
  arena serve tensorflow [flags]

Aliases:
  tensorflow, tf

Flags:
      --command string             the command will inject to container's command.
      --cpu string                 the request cpu of each replica to run the serve.
  -d, --data stringArray           specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>
      --enable-istio               enable Istio for serving or not (disable Istio by default)
  -e, --envs stringArray           the environment variables
      --expose-service             expose service using Istio gateway for external access or not (not expose by default)
      --gpumemory int              the limit GPU memory of each replica to run the serve.
      --gpus int                   the limit GPU count of each replica to run the serve.
  -h, --help                       help for tensorflow
      --image string               the docker image name of serve job, and the default image is tensorflow/serving:latest (default "tensorflow/serving:latest")
      --image-pull-policy string   the policy to pull the image, and the default policy is IfNotPresent (default "IfNotPresent")
      --memory string              the request memory of each replica to run the serve.
      --model-name string          the model name for serving
      --model-path string          the model path for serving in the container
      --modelConfigFile string     Corresponding with --model_config_file in tensorflow serving
      --name string                the serving name
      --port int                   the port of tensorflow gRPC listening port (default 8500)
      --replicas int               the replicas number of the serve job. (default 1)
      --restfulPort int            the port of tensorflow RESTful listening port (default 8501)
      --version string             the serving version
      --versionPolicy string       support latest, latest:N, specific:N, all

Options inherited from parent commands:
      --arena-namespace string   The namespace of arena system service, like tf-operator (default "arena-system")
      --config string            Path to a kube config. Only required if out-of-cluster
      --loglevel string          Set the logging level. One of: debug|info|warn|error (default "info")
  -n, --namespace string         the namespace of the job (default "default")
      --pprof                    enable cpu profile
      --trace                    enable trace
```

For example, you can submit a Tensorflow model with specific version policy as below.

```
arena serve tensorflow \
  --name=mymnist \
  --model-name=mnist \
  --image=tensorflow/serving:latest \
  --data=tfmodel:/tfmodel \
  --model-path=/tfmodel/mnist \
  --versionPolicy=specific:1  \
  --loglevel=debug
```

Once this command is triggered, one Kubernetes service will be created to expose gRPC and RESTful APIs of mnist model.


4\. Tensorflow serving with Istio enabled (optional)

If you need to enable Istio for Tensorflow serving,  you can append the parameter `--enableIstio` into the command above (disable Istio by default).

For example,  you can submit a Tensorflow model with Istio enabled as below.

```
$ arena serve tensorflow \
  --enableIstio \
  --name=mymnist \
  --servingVersion=v1 \
  --model-name=mnist \
  --data=myoss1pvc:/data2 \
  --model-path=/data2/models/mnist \
  --versionPolicy=specific:1 \

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

5\. List all the serving jobs

You can use the following command to list all the serving jobs.

```
# arena serve list
  NAME        VERSION  STATUS
  mymnist-v1  v1       DEPLOYED
```

6\. Adjust traffic routing dynamically for tfserving jobs
   
You can leverage Istio to control traffic routing to multiple versions of your serving models.

Supposing you've performed step 4, and had v1 model serving deployed already. Now deploy one new version of Tensorflow model with Istio enabled:
```
arena serve tensorflow \
  --enableIstio \
  --name=mymnist \
  --servingVersion=v2 \
  --modelName=mnist  \
  --data=myoss1pvc:/data2 \
  --model-path=/data2/models/mnist 
```

Then you can adjust traffic routing dynamically with relative weights for both two versions of tfserving jobs.
            
```
arena serve traffic-router-split \
  --name=mymnist \
  --servingVersions=v1,v2 \
  --weights=50,50
```

7\. Test RESTful APIs of serving models 

Deploy the `sleep` pod so you can use `curl` to test above serving models via RESTful APIs.

If you disable Istio, run the following:
```
# cat <<EOF | kubectl create -f -
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

If you enable Istio, run the following:
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

In this pod, use `curl` to call the exposed Tensorflow serving RESTful API:

```
# curl -X POST http://mymnist:8501/v1/models/mnist:predict -d '{"signature_name": "predict_images", "instances": [[0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.3294117748737335, 0.7254902124404907, 0.6235294342041016, 0.5921568870544434, 0.2352941334247589, 0.1411764770746231, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.8705883026123047, 0.9960784912109375, 0.9960784912109375, 0.9960784912109375, 0.9960784912109375, 0.9450981020927429, 0.7764706611633301, 0.7764706611633301, 0.7764706611633301, 0.7764706611633301, 0.7764706611633301, 0.7764706611633301, 0.7764706611633301, 0.7764706611633301, 0.6666666865348816, 0.2039215862751007, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.26274511218070984, 0.44705885648727417, 0.2823529541492462, 0.44705885648727417, 0.6392157077789307, 0.8901961445808411, 0.9960784912109375, 0.8823530077934265, 0.9960784912109375, 0.9960784912109375, 0.9960784912109375, 0.9803922176361084, 0.8980392813682556, 0.9960784912109375, 0.9960784912109375, 0.5490196347236633, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.06666667014360428, 0.25882354378700256, 0.05490196496248245, 0.26274511218070984, 0.26274511218070984, 0.26274511218070984, 0.23137256503105164, 0.08235294371843338, 0.9254902601242065, 0.9960784912109375, 0.41568630933761597, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.32549020648002625, 0.9921569228172302, 0.8196079134941101, 0.07058823853731155, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.08627451211214066, 0.9137255549430847, 1.0, 0.32549020648002625, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.5058823823928833, 0.9960784912109375, 0.9333333969116211, 0.1725490242242813, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.23137256503105164, 0.9764706492424011, 0.9960784912109375, 0.24313727021217346, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.5215686559677124, 0.9960784912109375, 0.7333333492279053, 0.019607843831181526, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.03529411926865578, 0.803921639919281, 0.9725490808486938, 0.22745099663734436, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.4941176772117615, 0.9960784912109375, 0.7137255072593689, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.29411765933036804, 0.9843137860298157, 0.9411765336990356, 0.22352942824363708, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.07450980693101883, 0.8666667342185974, 0.9960784912109375, 0.6509804129600525, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.011764707043766975, 0.7960785031318665, 0.9960784912109375, 0.8588235974311829, 0.13725490868091583, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.14901961386203766, 0.9960784912109375, 0.9960784912109375, 0.3019607961177826, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.12156863510608673, 0.8784314393997192, 0.9960784912109375, 0.45098042488098145, 0.003921568859368563, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.5215686559677124, 0.9960784912109375, 0.9960784912109375, 0.2039215862751007, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.2392157018184662, 0.9490196704864502, 0.9960784912109375, 0.9960784912109375, 0.2039215862751007, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.4745098352432251, 0.9960784912109375, 0.9960784912109375, 0.8588235974311829, 0.1568627506494522, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.4745098352432251, 0.9960784912109375, 0.8117647767066956, 0.07058823853731155, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0]]}'
```

You should update request url with your model service name accordingly. 

The value of "instances" is actually a list of numeric pixels of the first image (which is a hand-written digit "7") in MNIST test dataset.

So you may get response as below. It means the model predicts the input data as "7" with the highest probability among all 10 digits.

```
{
    "predictions": [[2.04608e-05, 1.72722e-09, 7.741e-05, 0.00364778, 1.25223e-06, 2.27522e-05, 1.14669e-08, 0.995975, 3.68833e-05, 0.000218786]]
}
```

8\. Delete one serving job

You can use the following command to delete a tfserving job and its associated pods
                                     
```
# arena serve delete mymnist --version v1
release "mymnist-v1" deleted
```
