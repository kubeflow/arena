# KServe job with supported serving runtime

This guide walks through the steps to deploy and serve a supported serving runtime with kserve.

1\. Setup

Follow the [KServe Guide](https://kserve.github.io/website/master/admin/serverless/serverless/) to install Kserve.

2\. Submit your serving job into kserve

deploy an InferenceService with a predictor that will load a scikit-learn model.

    $ arena serve kserve \
        --name=sklearn-iris \
        --model-format=sklearn \
        --storage-uri=gs://kfserving-examples/models/sklearn/1.0/model

    inferenceservice.serving.kserve.io/sklearn-iris created
    INFO[0009] The Job sklearn-iris has been submitted successfully
    INFO[0009] You can run `arena serve get sklearn-iris --type kserve -n default` to check the job status

3\. Check the status of KServe job

    $ arena serve list
    NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ADDRESS                                  PORTS
    sklearn-iris         KServe  00001    1        1          http://sklearn-iris.default.example.com  :80

    $ arena serve get sklearn-iris
    Name:       sklearn-iris
    Namespace:  default
    Type:       KServe
    Version:    00001
    Desired:    1
    Available:  1
    Age:        3m
    Address:    http://sklearn-iris.default.example.com
    Port:       :80
    
    LatestRevision:     sklearn-iris-predictor-00001
    LatestPrecent:      100
    
    Instances:
      NAME                                                      STATUS   AGE  READY  RESTARTS  NODE
      ----                                                      ------   ---  -----  --------  ----
      sklearn-iris-predictor-00001-deployment-7b4677c6b7-8cr84  Running  3m   2/2    0         192.168.5.239

4\. Perform inference

First, prepare your inference input request inside a file:

    $ cat <<EOF > "./iris-input.json"
    {
      "instances": [
        [6.8,  2.8,  4.8,  1.4],
        [6.0,  3.4,  4.5,  1.6]
      ]
    }
    EOF

you can curl with the ingress gateway external IP using the HOST Header.

    $ curl  -H "Host: sklearn-iris.default.example.com" http://${INGRESS_HOST}:80/v1/models/sklearn-iris:predict -d @./iris-input.json

5\. Update the InferenceService with the canary rollout strategy

Add the canaryTrafficPercent field to the predictor component and update the storageUri to use a new/updated model.

    $ arena serve update kserve \
    --name sklearn-iris \
    --canary-traffic-percent=10 \
    --storage-uri=gs://kfserving-examples/models/sklearn/1.0/model-2

After rolling out the canary model, traffic is split between the latest ready revision 2 and the previously rolled out revision 1.

    $ arena serve get sklearn-iris
    Name:       sklearn-iris
    Namespace:  default
    Type:       KServe
    Version:    00002
    Desired:    2
    Available:  2
    Age:        26m
    Address:    http://sklearn-iris.default.example.com
    Port:       :80
    
    LatestRevision:     sklearn-iris-predictor-00002
    LatestPrecent:      10
    PrevRevision:       sklearn-iris-predictor-00001
    PrevPrecent:        90
    
    Instances:
      NAME                                                      STATUS   AGE  READY  RESTARTS  NODE
      ----                                                      ------   ---  -----  --------  ----
      sklearn-iris-predictor-00001-deployment-7b4677c6b7-8cr84  Running  25m  2/2    0         192.168.5.239
      sklearn-iris-predictor-00002-deployment-7f677b9fd6-2dtpg  Running  3m   2/2    0         192.168.5.241

6\. Promote the canary model

If the canary model is healthy/passes your tests, you can set canary-traffic-percent to 100.

    $ arena serve update kserve \
    --name sklearn-iris \
    --canary-traffic-percent=100

Now all traffic goes to the revision 2 for the new model. The pods for revision generation 1 automatically scales down to 0 as it is no longer getting the traffic.

    $ arena serve get sklearn-iris
    Name:       sklearn-iris
    Namespace:  default
    Type:       KServe
    Version:    00002
    Desired:    1
    Available:  1
    Age:        32m
    Address:    http://sklearn-iris.default.example.com
    Port:       :80
    
    LatestRevision:     sklearn-iris-predictor-00002
    LatestPrecent:      100
    
    Instances:
      NAME                                                      STATUS       AGE  READY  RESTARTS  NODE
      ----                                                      ------       ---  -----  --------  ----
      sklearn-iris-predictor-00001-deployment-7b4677c6b7-8cr84  Terminating  31m  1/2    0         192.168.5.239
      sklearn-iris-predictor-00002-deployment-7f677b9fd6-2dtpg  Running      9m   2/2    0         192.168.5.241

7\. Delete the kserve job

    $ arena serve delete sklearn-iris

    









