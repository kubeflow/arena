This guide walks through the steps to deploy and serve a model with nvidia triton server.

1\. Create a pvc named triton-pvc with models to deploy.

2\. Summit your serving job into nvidia triton server.

```shell
$ arena serve triton \
 --name=test-triton \
 --namespace=triton \
 --gpus=1 \
 --image=nvcr.io/nvidia/tritonserver:20.12-py3 \
 --data=triton-pvc:/mnt/models \
 --model-repository=/mnt/models/ai/triton/model_repository
 
configmap/test-triton-202105312038-triton-serving created
configmap/test-triton-202105312038-triton-serving labeled
service/test-triton-202105312038-tritoninferenceserver created
deployment.apps/test-triton-202105312038-tritoninferenceserver created
INFO[0001] The Job test-triton has been submitted successfully 
INFO[0001] You can run `arena get test-triton --type triton-serving` to check the job status 
```

3\. List the job you were just serving


```shell
$ arena serve list -n triton
NAME         TYPE    VERSION       DESIRED  AVAILABLE  ADDRESS       PORTS
test-triton  Triton  202105312038  1        1          172.16.72.43  RESTFUL:8000,GRPC:8001
```

4\. Test the model service

```shell
$ kubectl get svc -n triton
NAME                                             TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                                        AGE
test-triton-202105312038-tritoninferenceserver   ClusterIP      172.16.72.43    <none>          8000/TCP,8001/TCP,8002/TCP                     5m41s

$ kubectl port-forward svc/test-triton-202105312038-tritoninferenceserver -n triton 8000:8000
Forwarding from 127.0.0.1:8000 -> 8000
Forwarding from [::1]:8000 -> 8000

# check models deploy success
$ curl -v localhost:8000/v2/health/ready
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8000 (#0)
> GET /v2/health/ready HTTP/1.1
> Host: localhost:8000
> User-Agent: curl/7.64.1
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Length: 0
< Content-Type: text/plain
<
* Connection #0 to host localhost left intact
* Closing connection 0
```

5\. Delete the inference service

```shell
$ arena serve delete test-triton -n triton                                                                                         
service "test-triton-202105312038-tritoninferenceserver" deleted
deployment.apps "test-triton-202105312038-tritoninferenceserver" deleted
configmap "test-triton-202105312038-triton-serving" deleted
INFO[0001] The serving job test-triton with version 202105312038 has been deleted successfully 
```