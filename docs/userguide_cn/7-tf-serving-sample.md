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
```

### 2.暴露job服务的端口

由于arena以ClusterIP的方式创建了服务，外部主机不能直接访问pod，我们需要将服务暴露出来。为了简化时间，选择"NodePort"方式是一个不错的选择，首先，我们应该获取服务名称：
```
# kubectl get svc
NAME                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)             AGE
fast-style-transfer-alpha   ClusterIP   172.21.8.94   <none>        8001/TCP,5000/TCP   107m
kubernetes                  ClusterIP   172.21.0.1    <none>        443/TCP             4d20h
``` 
服务的名称为"fast-style-transfer-alpha"，可以下面的命令编辑这个服务：

```
# kubectl edit svc fast-style-transfer-alpha
```
需要修改"spect"域的内容，具体修改如下：将"type"的值由"ClusterIP"改为"NodePort",同时在"spec.ports[1]"添加"nodePort: 32655"(也可以是其他端口):

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
然后使用":wq"保存退出，现在我们可以使用"kubectl get svc"获取服务新的配置信息:

```
# kubectl get svc
NAME                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                         AGE
fast-style-transfer-alpha   NodePort    172.21.8.94   <none>        8001:30268/TCP,5000:32655/TCP   119m
kubernetes                  ClusterIP   172.21.0.1    <none>        443/TCP                         4d20h
```
"5000:32655/TCP" means that we can access service using nodes port 32655.

### 3.访问服务 

我们可以下载fast-style-transfer然后使用它的例子：

```
# git clone https://github.com/floydhub/fast-style-transfer
```
可以看到，在项目的"images"目录下有一张图片"taipei101.jpg":

```
# cd fast-style-transfer
# ll images
total 120
-rw-r--r-- 1 root root 120702 7  30 15:45 taipei101.jpg
```
然后随便选择一个k8s节点来提交我们的请求:

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
这里我们选择"cn-beijing.192.168.3.228"这个节点，获取它的ip：

```
# kubectl describe nodes cn-beijing.192.168.3.228 | grep "InternalIP"
  InternalIP:  192.168.3.228
```
最后使用curl命令提交我们的请求：

```
# cd fast-style-transfer
# mkdir /tmp/out
# curl -o /tmp/out/taipei_out.jpg -F "file=@./images/taipei101.jpg" http://192.168.3.228:32655
```
在上面的curl命令中"-o"选项指定输出文件，-F选项指定输入文件的位置，这里我们使用了项目中的taipei101.jpg

