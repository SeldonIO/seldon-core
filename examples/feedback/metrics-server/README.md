# Stateful Model Feedback Metrics Server
In this example we will add statistical performance metrics capabilities by levering the Seldon metrics server.

Dependencies
* Seldon Core installed
* Ingress provider (Istio or Ambassador)

See the centralized logging example (also in the examples directory) for how to set these up.

An easy way is to run `examples/centralized-logging/full-kind-setup.sh` and then:
    `helm delete seldon-core-loadtesting`
    `helm delete seldon-single-model`

Then port-forward to that ingress on localhost:8080 in a separate terminal either with:

Ambassador:

    kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8080:8080

Istio:

    kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80




```python
!kubectl create namespace seldon || echo "namespace already created"
```

    Error from server (AlreadyExists): namespaces "seldon" already exists
    namespace already created



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "docker-desktop" modified.



```python
!mkdir -p config
```

### Create a simple model
We create a multiclass classification model - iris classifier.

The iris classifier takes an input array, and returns the prediction of the 4 classes.

The prediction can be done as numeric or as a probability array.


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: multiclass-model
spec:
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.11.0-dev/sklearn/iris
      name: classifier
      logger:
        url: http://seldon-multiclass-model-metrics.seldon.svc.cluster.local:80/
        mode: all
    name: default
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/multiclass-model created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=multiclass-model -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "multiclass-model-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "multiclass-model-default-0-classifier" successfully rolled out


#### Send test request


```python
res=!curl -X POST "http://localhost:8080/seldon/seldon/multiclass-model/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
print(res)
import json
j=json.loads(res[-1])
assert(len(j["data"]["ndarray"][0])==3)
```

    ['  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current', '                                 Dload  Upload   Total   Spent    Left  Speed', '', '  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0', '100   203  100   139  100    64   5148   2370 --:--:-- --:--:-- --:--:--  7518', '{"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"hello"}}']


### Metrics Server
You can create a kubernetes deployment of the metrics server with this:


```python
%%writefile config/multiclass-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: seldon-multiclass-model-metrics
  namespace: seldon
  labels:
    app: seldon-multiclass-model-metrics
spec:
  replicas: 1
  selector:
    matchLabels:
      app: seldon-multiclass-model-metrics
  template:
    metadata:
      annotations:
        prometheus.io/path: "/v1/metrics"
        prometheus.io/scrape: "true"
      labels:
        app: seldon-multiclass-model-metrics
    spec:
      securityContext:
          runAsUser: 8888
      containers:
      - name: user-container
        image: seldonio/alibi-detect-server:1.7.0-dev
        imagePullPolicy: Never
        args:
        - --model_name
        - multiclassserver
        - --http_port
        - '8080'
        - --protocol
        - seldonfeedback.http
        - --storage_uri
        - "adserver.cm_models.multiclass_one_hot.MulticlassOneHot"
        - --reply_url
        - http://message-dumper.default        
        - --event_type
        - io.seldon.serving.feedback.metrics
        - --event_source
        - io.seldon.serving.feedback
        - MetricsServer
        env:
        - name: "SELDON_DEPLOYMENT_ID"
          value: "multiclass-model"
        - name: "PREDICTIVE_UNIT_ID"
          value: "classifier"
        - name: "PREDICTIVE_UNIT_IMAGE"
          value: "alibi-detect-server:1.7.0-dev"
        - name: "PREDICTOR_ID"
          value: "default"
        ports:
        - containerPort: 8080
          name: metrics
          protocol: TCP
---
apiVersion: v1
kind: Service
metadata:
  name: seldon-multiclass-model-metrics
  namespace: seldon
  labels:
    app: seldon-multiclass-model-metrics
spec:
  selector:
    app: seldon-multiclass-model-metrics
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

    Overwriting config/multiclass-deployment.yaml



```python
!kubectl apply -n seldon -f config/multiclass-deployment.yaml
```

    deployment.apps/seldon-multiclass-model-metrics unchanged
    service/seldon-multiclass-model-metrics unchanged



```python
!kubectl rollout status deploy/seldon-multiclass-model-metrics
```

    deployment "seldon-multiclass-model-metrics" successfully rolled out



```python
import time
time.sleep(20)
```

In the above the SeldonDeployment is set to log directly to the metrics server. If you want to use async queuing, see the feedback-metrics-server example.


### Send feedback


```python
res=!curl -X POST "http://localhost:8080/seldon/seldon/multiclass-model/api/v1.0/feedback" \
        -H "Content-Type: application/json" \
        -d '{"response": {"data": {"ndarray": [[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]}}, "truth":{"data": {"ndarray": [[0,0,1]]}}}'
print(res)
import json
j=json.loads(res[-1])
assert("data" in j)
```

    ['  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current', '                                 Dload  Upload   Total   Spent    Left  Speed', '', '  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0', '100   188  100    44  100   144   1419   4645 --:--:-- --:--:-- --:--:--  6064', '{"data":{"tensor":{"shape":[0]}},"meta":{}}']



```python
import time
time.sleep(3)
```

### Check that metrics are recorded


```python
res=!kubectl logs $(kubectl get pods -l app=seldon-multiclass-model-metrics \
                    -n seldon -o jsonpath='{.items[0].metadata.name}') | grep "PROCESSING Feedback Event"
print(res)
assert(len(res)>0)
```

    ['[I 201008 15:32:09 cm_model:77] PROCESSING Feedback Event.', '[I 201008 15:33:24 cm_model:77] PROCESSING Feedback Event.', '[I 201008 15:44:15 cm_model:77] PROCESSING Feedback Event.']


### Cleanup


```python
!kubectl delete -n seldon -f config/multiclass-deployment.yaml
```

    deployment.apps "seldon-multiclass-model-metrics" deleted
    service "seldon-multiclass-model-metrics" deleted



```python
!kubectl delete sdep multiclass-model
```

    seldondeployment.machinelearning.seldon.io "multiclass-model" deleted



```python

```
