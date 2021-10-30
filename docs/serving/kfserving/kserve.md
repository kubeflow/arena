# KFServing job 

This guide walks through the steps to deploy and serve a  model with kfserving

1\. Setup

Follow the [KFserving Guide](https://github.com/kubeflow/kfserving#install-kfserving) to install kFserving.For the prerequisites,you should ensure 8g memery and 4 core cpu avaliable in your environment.

2\. summit your serving job into kfserving

    $ arena serve kfserving \
      >         --name=sklearn-demo \
      >         --storage-uri="https://github.com/tduffy000/kfserving-uri-examples/blob/master/sklearn/frozen/model.joblib?raw=true" \
      >         --model-type=sklearn
      INFO[0000] The Job sklearn-demo has been submitted successfully 
      inferenceservice.serving.kserve.io/sklearn-demo-202110301430 created
      INFO[0009] The Job sklearn-demo has been submitted successfully 



3\. list the job you just serving

    $ arena serve ls                                                                        
     NAME                       TYPE       VERSION  DESIRED  AVAILABLE  ADDRESS  PORTS
     sklearn-demo-202110301430  KFServing  1428228  0        0                   N/A


4\. test the model service

step1: Determine the ingress IP and ports

The first step is to [determine the ingress IP](https://github.com/kubeflow/kfserving/blob/master/README.md#determine-the-ingress-ip-and-ports) and ports and set INGRESS_HOST and INGRESS_PORT.

This example uses the [codait/max-object-detector](https://github.com/IBM/MAX-Object-Detector) image. The Max Object Detector api server expects a POST request to the /model/predict endpoint that includes an image multipart/form-data and an optional threshold query string.

    MODEL_NAME=sklearn-demo-202110301430
    SERVICE_HOSTNAME=$(kubectl get inferenceservice ${MODEL_NAME} -o jsonpath='{.status.url}' | cut -d "/" -f 3)
    INGRESS_HOST=localhost
    INGRESS_PORT=80
    curl -v -H "Host: ${SERVICE_HOSTNAME}" http://${INGRESS_HOST}:${INGRESS_PORT}/v1/models/$MODEL_NAME:predict -d '{"instances": [[6.8,  2.8,  4.8,  1.4],[6.0,  3.4,  4.5,  1.6]]}' 
    
    *   Trying ::1...
    * TCP_NODELAY set
    * Connected to localhost (::1) port 80 (#0)
    > POST /v1/models/sklearn-demo-202110301430:predict HTTP/1.1
    > Host: sklearn-demo-202110301430.default.example.com
    > User-Agent: curl/7.64.1
    > Accept: */*
    > Content-Length: 64
    > Content-Type: application/x-www-form-urlencoded
    > 
    * upload completely sent off: 64 out of 64 bytes
    < HTTP/1.1 200 OK
    < content-length: 23
    < content-type: application/json; charset=UTF-8
    < date: Sat, 30 Oct 2021 06:49:17 GMT
    < server: istio-envoy
    < x-envoy-upstream-service-time: 10
    < 
    * Connection #0 to host localhost left intact
    {"predictions": [1, 1]}* Closing connection 0
