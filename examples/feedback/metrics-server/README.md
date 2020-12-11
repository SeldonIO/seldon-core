# Stateful Model Feedback Metrics Server
In this example we will add statistical performance metrics capabilities by levering the Seldon metrics server.

Dependencies
* Seldon Core installed
* Ingress provider (Istio or Ambassador)
* KNative eventing v0.11.0 (optional)
* KNative serving v0.11.1 (optional)

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

Ambassador:

    kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080

Istio:

    kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:80





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
      modelUri: gs://seldon-models/sklearn/iris
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
res=!curl -X POST "http://localhost:8003/seldon/seldon/multiclass-model/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
print(res)
import json
j=json.loads(res[-1])
assert(len(j["data"]["ndarray"][0])==3)
```

    ['  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current', '                                 Dload  Upload   Total   Spent    Left  Speed', '', '  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0', '100   203  100   139  100    64   5148   2370 --:--:-- --:--:-- --:--:--  7518', '{"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"hello"}}']


### Kubernetes Deployment
You can create a kubernetes deployment instead of a kservice with the yaml below. For kservice the details are below.


```python
%%writefile config/multiclass-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: seldon-multiclass-model-metrics
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
        image: seldonio/alibi-detect-server:1.6.0-dev
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
          value: "alibi-detect-server:1.6.0-dev"
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

### (Alternative) create kservice

If you want to create a kservice, and you've installed knative eventing and knative serving, you can use the instructions below.

The value of the file `config/multiclass-service.yaml` would be:
```
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: seldon-multiclass-model-metrics
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        prometheus.io/path: "/v1/metrics"
        prometheus.io/scrape: "true"
    spec:
      containers:
      - image: "seldonio/alibi-detect-server:1.6.0-dev"
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
          value: "alibi-detect-server:1.6.0-dev"
        - name: "PREDICTOR_ID"
          value: "default"
        ports:
        - containerPort: 8080
          name: metrics
          protocol: TCP
        securityContext:
            runAsUser: 8888
```

You can run the kservice with the command below:
```
kubectl apply -f config/multiclass-service.yaml
```
And then check with:

```
kubectl get kservice
```

You'll then have to create the trigger, first by creating the broker:

```
kubectl label namespace default knative-eventing-injection=enabled --overwrite=true
```

And then the trigger contents:

```
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  name: multiclass-model-metrics-trigger
  namespace: default
spec:
  filter:
    sourceAndType:
      type: io.seldon.serving.feedback
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: seldon-multiclass-model-metrics
```

And you can run it with:
```
kubectl apply -f config/trigger.yaml
```

### Send feedback


```python
res=!curl -X POST "http://localhost:8003/seldon/seldon/multiclass-model/api/v1.0/feedback" \
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
