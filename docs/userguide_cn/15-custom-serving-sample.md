# 用arena服务训练模型

你可以适用arena部署你的训练模型，通过RESTful API的方式访问。为了说明怎样使用，我们将会使用一个案例[fast-style-transfer](https://github.com/floydhub/fast-style-transfer)，同时为了节约时间，直接使用这个项目已经训练好的模型并且把模型加入docker镜像中。

### 1.部署训练模型

使用项目中的app.py这个脚本启动一个restful服务器，你可以使用如下的命令去部署模型：

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
检查TensorFlow Serving Job的状态：

```
# arena serve list
NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ENDPOINT_ADDRESS  PORTS
fast-style-transfer  CUSTOM  alpha    1        0          172.21.8.94       grpc:8001,restful:5000
```
因为docker镜像比较大，拉取它需要一定的时间，我们可以使用"kubectl"检查pod运行情况:

```
# kubectl get po
NAME                                                        READY   STATUS              RESTARTS   AGE
fast-style-transfer-alpha-custom-serving-845ffbf7dd-btbhj   0/1     ContainerCreating   0          6m44s


### 2.访问服务 

我们可以使用一个带有curl命令的容器作为客户端去访问刚才创建的服务，但是首先我们需要创建这个客户端：
```
# kubectl run  sample-client \
	--generator=run-pod/v1 \
	--image=happy365/arena-serve-custem-sample-client:latest \
	--command -- \
	/bin/sleep infinity
```
然后，可以查询客户端的状态：
```
# kubectl get po  sample-client
NAME            READY   STATUS    RESTARTS   AGE
sample-client   1/1     Running   0          87s 

```
在用客户端访问custom service之前，我们需要查询服务名称，它是一个任务名和版本的结合（本例中，任务名为fast-style-transfer，版本为alpha)：

```
# kubectl get svc fast-style-transfer-alpha
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
fast-style-transfer-alpha   ClusterIP   172.21.1.114   <none>        5000/TCP   31m
```
现在我们可以可以使用kubectl exec 进入容器当中：

```
# kubectl exec -ti sample-client /bin/sh
#
```
接着在容器当中使用curl命令去访问aren创建的自定义服务:
```
# curl -o /root/output/beijing_out.jpg  -F "file=@/root/input/beijing.jpg" http://fast-style-transfer-alpha:5000
```
在上面的命令中，输入文件的名称为![beijing.jpg](15-custom-serving-sample-beijing.jpg)，存放的路径为"/root/input"，输出文件的路径为"/root/output/beijing_out.jpg"，现在需要退出容器然后在master节点上执行kubectl cp命令将结果从容器中拷贝出来：
```
# kubectl cp sample-client:/root/output/beijing_out.jpg ~/beijing_out.jpg
```
图片![beijing_out.jpg](15-custom-serving-sample-beijing_out.jpg)将会复制到当前用户的家目录下面。



