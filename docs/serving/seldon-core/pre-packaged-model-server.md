This guide walks through the steps to deploy and serve a model with seldon core pre-packaged model server.

1\. Setup

Follow the Seldon Core [guide](https://github.com/SeldonIO/seldon-core#install-seldon-core) to install Seldon Core.For the prerequisites,you should ensure 8g memory and 4 core cpu available in your environment.


2\. Summit your serving job into seldon core

```shell
$ arena serve seldon --implementation SKLEARN_SERVER --modelUri gs://seldon-models/sklearn/iris --name sklearn-iris
configmap/sklearn-iris-202102222213-seldon-core-serving created
configmap/sklearn-iris-202102222213-seldon-core-serving labeled
seldondeployment.machinelearning.seldon-core.io/sklearn-iris created

```

3\. List the job you were just serving


```shell
$ arena serve list 
NAME          TYPE        VERSION       DESIRED  AVAILABLE  ADDRESS       PORTS
sklearn-iris  Seldon      202102222213  1        1          172.29.11.50  RESTFUL:8000,GRPC:5001
```

4\. Test the model service

```shell
$ kubectl get svc
NAME                                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
sklearn-iris-default                          ClusterIP   172.29.2.217    <none>        8000/TCP,5001/TCP   2m8s
sklearn-iris-default-inference                ClusterIP   172.29.13.193   <none>        9000/TCP,9500/TCP   2m46s

$ kubectl port-forward svc/sklearn-iris-default 8000:8000
Forwarding from 127.0.0.1:8000 -> 8000
Forwarding from [::1]:8000 -> 8000

$ curl -X POST http://localhost:8000/api/v1.0/predictions -H 'Content-Type: application/json' -d '{ "data": { "ndarray": [[1,2,3,4]] } }' 

{
   "data" : {
      "names" : [
         "t:0",
         "t:1",
         "t:2"
      ],
      "ndarray" : [
         [
            0.000698519453116284,
            0.00366803903943576,
            0.995633441507448
         ]
      ]
   }
}

```

5\. Delete the inference service

```shell
$ arena serve delete sklearn-iris                                                                                          
seldondeployment.machinelearning.seldon-core.io "sklearn-iris" deleted
configmap "sklearn-iris-202102222213-seldon-serving" deleted
INFO[0001] The serving job sklearn-iris with version 202102222213 has been deleted successfully 
```