# Serving Trained Model with arena

You can use arena to deploy your trained model as RESTful APIs.to illustrate usage,we use a sample project [fast-style-transfer](https://github.com/floydhub/fast-style-transfer).in order to save time,we use its' trainted model and add the model to docker images.

### 1.Serve Mode

we use the app.py script in project to start restful server,you can use arena to deploy trainted model:

```
# arena serve custom \
	--name=fast-style-transfer \
	--gpus=1 \
        --version=alpha \
	--replicas=1 \
	--restful-port=5000 \
	--image=happy365/fast-style-transfer:latest \
	"python app.py"
``` 

check the status of TensorFlow Serving Job:

```
# arena serve list
NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ENDPOINT_ADDRESS  PORTS
fast-style-transfer  CUSTOM  alpha    1        0          172.21.8.94       grpc:8001,restful:5000
```

because the docker image is very large,pulling it requests some time,we can use kubectl to check the pod status:

```
# kubectl get po
NAME                                                        READY   STATUS              RESTARTS   AGE
fast-style-transfer-alpha-custom-serving-845ffbf7dd-btbhj   0/1     ContainerCreating   0          6m44s
```

### 2.Expose Job port 

we should expose service when all pods' status are running,in order to simplify operations,chosing "NodePort" method is a good way,first we should get the service name using follow command:
```
# kubectl get svc
NAME                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)             AGE
fast-style-transfer-alpha   ClusterIP   172.21.8.94   <none>        8001/TCP,5000/TCP   107m
kubernetes                  ClusterIP   172.21.0.1    <none>        443/TCP             4d20h
``` 
the service name is "fast-style-transfer-alpha",then we edit the service using follow command:

```
# kubectl edit svc fast-style-transfer-alpha
```
content of "spec" field should be modified,the value of "type" should be "NodePort" and add "nodePort: 32655" to "spec.ports[1]"(port "32655" can be customized):

```
spec:
  clusterIP: 172.21.8.94
  ports:
  - name: grpc
    port: 8001
    protocol: TCP
    targetPort: 8001
  - name: restful
    port: 5000
    protocol: TCP
    targetPort: 5000
    nodePort: 32655
  selector:
    app: custom-serving
    servingVersion: alpha
  sessionAffinity: None
  type: NodePort
```
then use ":wq" to save config,we can use "kubectl get svc" to get new configuration of service:

```
# kubectl get svc
NAME                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                         AGE
fast-style-transfer-alpha   NodePort    172.21.8.94   <none>        8001:30268/TCP,5000:32655/TCP   119m
kubernetes                  ClusterIP   172.21.0.1    <none>        443/TCP                         4d20h
```
"5000:32655/TCP" means that we can access service using nodes port 32655.

### 3.access service 

we should download fast-style-transfer to use its' examples firstly:

```
# git clone https://github.com/floydhub/fast-style-transfer
```
there is an image in "images" directory:

```
# cd fast-style-transfer
# ll images
total 120
-rw-r--r-- 1 root root 120702 7  30 15:45 taipei101.jpg
```
then we chose a node to post our request:

```
# kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.3.225   Ready    master   4d20h   v1.12.6-aliyun.1
cn-beijing.192.168.3.226   Ready    master   4d20h   v1.12.6-aliyun.1
cn-beijing.192.168.3.227   Ready    master   4d20h   v1.12.6-aliyun.1
cn-beijing.192.168.3.228   Ready    <none>   4d20h   v1.12.6-aliyun.1
cn-beijing.192.168.3.229   Ready    <none>   4d20h   v1.12.6-aliyun.1
cn-beijing.192.168.3.230   Ready    <none>   4d20h   v1.12.6-aliyun.1
```

in here,we chose node "cn-beijing.192.168.3.228" and get its' ip:

```
# kubectl describe nodes cn-beijing.192.168.3.228 | grep "InternalIP"
  InternalIP:  192.168.3.228
```
then use "curl" to post our request:

```
# cd fast-style-transfer
# mkdir /tmp/out
# curl -o /tmp/out/taipei_out.jpg -F "file=@./images/taipei101.jpg" http://192.168.3.228:32655
```

the output image is "/tmp/out/taipei_out.jpg",input image is "fast-style-transfer/images/taipei101.jpg".

