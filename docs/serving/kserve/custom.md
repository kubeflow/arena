# KServe job with custom serving runtime

This guide walks through the steps to deploy and serve a custom serving runtime with kserve.

1\. Setup

Follow the [KServe Guide](https://kserve.github.io/website/master/admin/serverless/serverless/) to install Kserve.

2\. Submit your serving job into kserve

create a PVC 'training-data' before, and then download the 'bloom-560m' model from HuggingFace to the PVC.

deploy an InferenceService with a predictor that will load a bloom model with text-generation-inference.

    $ arena serve kserve \
        --name=bloom-560m \
        --image=ghcr.io/huggingface/text-generation-inference:1.0.2 \
        --gpus=1 \
        --cpu=12 \
        --memory=50Gi \
        --port=8080 \
        --env=STORAGE_URI=pvc://training-data \
        "text-generation-launcher --disable-custom-kernels --model-id /mnt/models/bloom-560m --num-shard 1 -p 8080"

    inferenceservice.serving.kserve.io/bloom-560m created
    INFO[0010] The Job bloom-560m has been submitted successfully
    INFO[0010] You can run `arena serve get bloom-560m --type kserve -n default` to check the job status

3\. Check the status of KServe job

    $ arena serve list
    NAME                 TYPE    VERSION  DESIRED  AVAILABLE  ADDRESS                                  PORTS
    bloom-560m   KServe  00001    1        1          http://bloom-560m.default-group.example.com  :80    1

    $ arena serve get sklearn-iris
    Name:       bloom-560m
    Namespace:  default
    Type:       KServe
    Version:    00001
    Desired:    1
    Available:  1
    Age:        7m
    Address:    http://bloom-560m.default.example.com
    Port:       :80
    GPU:        1
    
    LatestRevision:     bloom-560m-predictor-00001
    LatestPrecent:      100
    
    Instances:
      NAME                                                    STATUS   AGE  READY  RESTARTS  GPU  NODE
      ----                                                    ------   ---  -----  --------  ---  ----
      bloom-560m-predictor-00001-deployment-56b8bdbf87-sg8v8  Running  7m   2/2    0         1    192.168.5.241

4\. Perform inference

you can curl with the ingress gateway external IP using the HOST Header.

    $ curl -H "Host: bloom-560m.default.example.com" http://${INGRESS_HOST}:80/generate \
    -X POST \
    -d '{"inputs":"What is Deep Learning?","parameters":{"max_new_tokens":17}}' \
    -H 'Content-Type: application/json'

    {"generated_text":" Deep Learning is a new type of machine learning that is used to solve complex problems."}

5\. Update the InferenceService with the canary rollout strategy

Add the canaryTrafficPercent field to the predictor component and update command to use a new/updated model path /mnt/models/bloom-560m-v2.

    $ arena serve update kserve \
    --name bloom-560m \
    --canary-traffic-percent=10 \
    "text-generation-launcher --disable-custom-kernels --model-id /mnt/models/bloom-560m-v2 --num-shard 1 -p 8036"

After rolling out the canary model, traffic is split between the latest ready revision 2 and the previously rolled out revision 1.

    $ arena serve get bloom-560m
    Name:       bloom-560m
    Namespace:  default
    Type:       KServe
    Version:    00002
    Desired:    2
    Available:  2
    Age:        26m
    Address:    http://bloom-560m.default.example.com
    Port:       :80
    
    LatestRevision:     bloom-560m-predictor-00002
    LatestPrecent:      10
    PrevRevision:       bloom-560m-predictor-00001
    PrevPrecent:        90
    
    Instances:
      NAME                                                    STATUS   AGE   READY  RESTARTS  GPU  NODE
      ----                                                    ------   ---   -----  --------  ---  ----
      bloom-560m-predictor-00001-deployment-56b8bdbf87-sg8v8  Running  19m   2/2    0         1    192.168.5.241
      bloom-560m-predictor-00002-deployment-84dbb64cc4-647wx  Running  2m    2/2    0         1    192.168.5.239

6\. Promote the canary model

If the canary model is healthy/passes your tests, you can set canary-traffic-percent to 100.

    $ arena serve update kserve \
    --name bloom-560m \
    --canary-traffic-percent=100

Now all traffic goes to the revision 2 for the new model. The pods for revision generation 1 automatically scales down to 0 as it is no longer getting the traffic.

    $ arena serve get bloom-560m
    Name:       bloom-560m
    Namespace:  default
    Type:       KServe
    Version:    00002
    Desired:    2
    Available:  2
    Age:        26m
    Address:    http://bloom-560m.default.example.com
    Port:       :80

    LatestRevision:     bloom-560m-predictor-00002
    LatestPrecent:      100
    
    Instances:
      NAME                                                    STATUS       AGE  READY  RESTARTS  GPU  NODE
      ----                                                    ------       ---  -----  --------  ---  ----
      bloom-560m-predictor-00001-deployment-56b8bdbf87-sg8v8  Terminating  22m  1/2    0         0    192.168.5.241
      bloom-560m-predictor-00002-deployment-84dbb64cc4-647wx  Running      5m   2/2    0         1    192.168.5.239

7\. Delete the kserve job

    $ arena serve delete sklearn-iris

    









