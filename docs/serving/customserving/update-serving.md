After the custom serving has deployed，you can use `arena serve update` to update the serving docker images, replicas.

1\. Deploy the serving job

we use the app.py script in project to start restful server,you can use arena to deploy trainted model:

```shell
$ arena serve custom \
	--name=fast-style-transfer \
	--gpus=1 \
	--version=alpha \
	--replicas=1 \
	--restful-port=5000 \
	--image=happy365/fast-style-transfer:latest \
	"python app.py"
```


check the status of TensorFlow Serving Job:

```shell
$ arena serve list
NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ADDRESS       PORTS
fast-style-transfer  Custom  alpha    1        0          172.16.113.5  RESTFUL:5000
```


because the docker image is very large,pulling it requests some time,we can use kubectl to check the pod status:

```shell
$ kubectl get pods
NAME                                                       READY   STATUS              RESTARTS   AGE
fast-style-transfer-alpha-custom-serving-6988f57d4-dd6v6   0/1     ContainerCreating   0          34s
```

2\. Scale the serving replicas

if you want to scale the replicas，you can use arena serve update custom to update the serving.

```shell
$ arena serve update custom
  --name=fast-style-transfer 
  --replicas=2
```

check the pod number

```shell
$ kubectl get pods
NAME                                                       READY   STATUS    RESTARTS   AGE
fast-style-transfer-alpha-custom-serving-6988f57d4-dd6v6   1/1     Running   0          4m34s
fast-style-transfer-alpha-custom-serving-6988f57d4-tqqlc   1/1     Running   0          59s
```

2\. Update the serving docker image

if you want update the docker image, you can use the command below

```shell
$ arena serve update custom
  --name=fast-style-transfer 
  --image=happy365/fast-style-transfer:0.0.1
```

After you execute the command, the custom serving will do rolling update with the support of kubernetes deployment.
