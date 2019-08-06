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

### 2.Access the service  

we can use a client to access the service,run the follow command to create a client:
```
# kubectl run  sample-client \
	--generator=run-pod/v1 \
	--image=happy365/arena-serve-custem-sample-client:latest \
	--command -- \
	/bin/sleep infinity
```

then,we can query the status of sample-client:
```
# kubectl get po  sample-client
NAME            READY   STATUS    RESTARTS   AGE
sample-client   1/1     Running   0          87s 

```
we should query the sevice name,it is a combination of job name and version(the sample job name is fast-style-transfer and version is alpha):

```
# kubectl get svc fast-style-transfer-alpha
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
fast-style-transfer-alpha   ClusterIP   172.21.1.114   <none>        5000/TCP   31m
```

now,we can use the "kubectl exec" command to login the container:

```
# kubectl exec -ti sample-client /bin/sh
#
```

then we use "curl" command to access the custom serving job:
```
# curl -o /root/output/beijing_out.jpg  -F "file=@/root/input/beijing.jpg" http://fast-style-transfer-alpha:5000
```
the input is an image which name is "beijing.jpg" ![beijing.jpg](15-custom-serving-sample-beijing.jpg),the image is stored in "/root/input",the output is  stored in "/root/output". you can use "kubectl cp" command to copy output image from container to host:
```
# kubectl cp sample-client:/root/output/beijing_out.jpg ~/beijing_out.jpg
```
now you can view the image in ~/beijing_out.jpg,there is "beijing_out.jpg" ![beijing_out.jpg](15-custom-serving-sample-beijing_out.jpg)



