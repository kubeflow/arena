# KFServing job with Custom type

This guide walks through the steps to deploy and serve a custom model with kfserving

1\. Setup

Follow the [KFserving Guide](https://github.com/kubeflow/kfserving#install-kfserving) to install kFserving.For the prerequisites,you should ensure 8g memery and 4 core cpu avaliable in your environment.

2\. summit your serving job into kfserving

    $ arena serve kfserving \
        --name=max-object-detector \
        --port=5000 \
        --image=codait/max-object-detector \
        --model-type=custom

    configmap/max-object-detector-202008221942-kfserving created
    configmap/max-object-detector-202008221942-kfserving labeled


3\. list the job you just serving

    $ arena serve list 
    NAME                 TYPE       VERSION       DESIRED  AVAILABLE  ENDPOINT_ADDRESS  PORTS
    max-object-detector  KFSERVING  202008221942  1        1          10.97.52.65       http:80

4\. test the model service

step1: Determine the ingress IP and ports

The first step is to [determine the ingress IP](https://github.com/kubeflow/kfserving/blob/master/README.md#determine-the-ingress-ip-and-ports) and ports and set INGRESS_HOST and INGRESS_PORT.

This example uses the [codait/max-object-detector](https://github.com/IBM/MAX-Object-Detector) image. The Max Object Detector api server expects a POST request to the /model/predict endpoint that includes an image multipart/form-data and an optional threshold query string.

    $ MODEL_NAME=max-object-detector-202008221942
    $ SERVICE_HOSTNAME=$(kubectl get inferenceservice ${MODEL_NAME} -o jsonpath='{.status.url}' | cut -d "/" -f 3)
    $ INGRESS_HOST=localhost
    $ INGRESS_PORT=80
    $ curl -v -F "image=@27-kfserving-custom.jpg" http://${INGRESS_HOST}:${INGRESS_PORT}/model/predict -H "Host: ${SERVICE_HOSTNAME}"

    *   Trying ::1...
    * TCP_NODELAY set
    * Connected to localhost (::1) port 80 (#0)
    > POST /model/predict HTTP/1.1
    > Host: max-object-detector-202008221942.default.example.com
    > User-Agent: curl/7.64.1
    > Accept: */*
    > Content-Length: 125769
    > Content-Type: multipart/form-data; boundary=------------------------56b67bc60fc7bdc7
    > Expect: 100-continue
    >
    < HTTP/1.1 100 Continue
    * We are completely uploaded and fine
    < HTTP/1.1 200 OK
    < content-length: 380
    < content-type: application/json
    < date: Sun, 23 Aug 2020 03:27:14 GMT
    < server: istio-envoy
    < x-envoy-upstream-service-time: 3566
    <
    {"status": "ok", "predictions": [{"label_id": "1", "label": "person", "probability": 0.9440352320671082, "detection_box": [0.12420991063117981, 0.12507185339927673, 0.8423266410827637, 0.5974075794219971]}, {"label_id": "18", "label": "dog", "probability": 0.8645510673522949, "detection_box": [0.10447663068771362, 0.17799144983291626, 0.8422801494598389, 0.7320016026496887]}]}
    * Connection #0 to host localhost left intact
    * Closing connection 0

5\. delete the serving job

    $ arena serve delete max-object-detector --version=202008221942                                                                                                   2 err 
    inferenceservice.serving.kubeflow.org "max-object-detector-202008221942" deleted
    configmap "max-object-detector-202008221942-kfserving" deleted
    INFO[0001] The Serving job max-object-detector with version 202008221942 has been deleted successfully 
