# Stateful Elasticsearch Feedback Workflow for Metrics Server
In this example we will add statistical performance metrics capabilities by levering the Seldon metrics server with persistence through the elasticsearch setup.

Dependencies
* Seldon Core installed
* Ingress provider (Istio or Ambassador)
* Install [Elasticsearch for the Seldon Core Logging](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/logging.html)
* KNative eventing v0.18.3
* KNative serving v0.18.1 (optional)

See the centralized logging example (also in the examples directory) for how to set these up.

Easiest way is to run `examples/centralized-logging/full-kind-setup.sh` and then:
    `helm delete seldon-core-loadtesting`
    `helm delete seldon-single-model`

Then port-forward to that ingress on localhost:8080 in a separate terminal with either of (istio suggested):

Ambassador:

    kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8080

Istio:

    kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80





```python
%%writefile requirements-dev.txt
elasticsearch==7.9.1
```

    Overwriting requirements-dev.txt



```python
!pip install -r requirements-dev.txt
```

    Collecting elasticsearch==7.9.1
    [?25l  Downloading https://files.pythonhosted.org/packages/e4/b7/f8f03019089671486e2910282c1b6fce26ccc8a513322df72ac8994ab2de/elasticsearch-7.9.1-py2.py3-none-any.whl (219kB)
    [K     |████████████████████████████████| 225kB 1.1MB/s eta 0:00:01
    [?25hRequirement already satisfied: certifi in /home/alejandro/miniconda3/lib/python3.7/site-packages (from elasticsearch==7.9.1->-r requirements-dev.txt (line 1)) (2019.9.11)
    Requirement already satisfied: urllib3>=1.21.1 in /home/alejandro/miniconda3/lib/python3.7/site-packages (from elasticsearch==7.9.1->-r requirements-dev.txt (line 1)) (1.24.2)
    Installing collected packages: elasticsearch
      Found existing installation: elasticsearch 7.5.1
        Uninstalling elasticsearch-7.5.1:
          Successfully uninstalled elasticsearch-7.5.1
    Successfully installed elasticsearch-7.9.1



```python
!kubectl create namespace seldon || echo "namespace already created"
!kubectl create namespace seldon-logs || echo "namespace already created"
```

    namespace/seldon created



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "docker-desktop" modified.



```python
!mkdir -p config
```

Setting up Knative eventing routing for request logger 


```bash
%%bash
kubectl apply -f - <<EOF
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: default
  namespace: seldon-logs
EOF
```

Verify broker is up

```bash
%%bash
kubectl -n seldon-logs get broker default -o jsonpath='{.status.address.url}'
```

Adding payload request logger component for redirection of logs


```bash
%%bash
kubectl apply -f - << END
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: seldon-request-logger-trigger
  namespace: seldon-logs
spec:
  broker: default
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: seldon-request-logger
END

kubectl apply -f - << END
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: seldon-request-logger
  namespace: seldon-logs
  metadata:
    labels:
    fluentd: "true"
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
        - image: docker.io/seldonio/seldon-request-logger:1.5.1
          imagePullPolicy: Always
          env:
           - name: ELASTICSEARCH_HOST
             value: "elasticsearch-opendistro-es-client-service.seldon-logs.svc.cluster.local"
           - name: ELASTICSEARCH_PORT
             value: "9200"
           - name: ELASTICSEARCH_PROTOCOL
             value: "https"
           - name: ELASTICSEARCH_USER
             value: "admin"
           - name: ELASTICSEARCH_PASS
             value: "admin"
END
```

    trigger.eventing.knative.dev/seldon-request-logger-trigger unchanged
    deployment.apps/seldon-request-logger configured
    service/seldon-request-logger unchanged


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
        url: http://broker-ingress.knative-eventing.svc.cluster.local/seldon-logs/default
        mode: all
    name: default
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/multiclass-model configured



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=multiclass-model -o jsonpath='{.items[0].metadata.name}')
```

    deployment "multiclass-model-default-0-classifier" successfully rolled out


### Send Prediction Request


```python
import requests
url = "http://localhost:8080/seldon/seldon/multiclass-model/api/v1.0"
```


```python
pred_req_1 = {"data":{"ndarray":[[1,2,3,4]]}}
pred_resp_1 = requests.post(f"{url}/predictions", json=pred_req_1)
print(pred_resp_1.json())
assert(len(pred_resp_1.json()["data"]["ndarray"][0])==3)
```

    {'data': {'names': ['t:0', 't:1', 't:2'], 'ndarray': [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, 'meta': {}}


### Check data in Elasticsearch
We'll be able to check the elasticsearch through the service or the pods in our cluster.

To do this we'll have to port-forward to elastic in another window. e.g.

kubectl port-forward -n seldon-logs svc/elasticsearch-opendistro-es-client-service 9200

Verify by going to https://admin:admin@localhost:9200/_cat/indices


```python
from elasticsearch import Elasticsearch
es = Elasticsearch(['https://admin:admin@localhost:9200'],verify_certs=False)
```

See the indices that have been created


```python
es.indices.get_alias("*")
```




    {'inference-log-seldon-seldon-multiclass-model-default': {'aliases': {}}}



Look at the data that is stored in the elasticsearch index


```python
res = es.search(index="inference-log-seldon-seldon-multiclass-model-default", body={"query": {"match_all": {}}})
print("Logged Request:")
print(res["hits"]["hits"][0]["_source"]["request"])
print("\nLogged Response:")
print(res["hits"]["hits"][0]["_source"]["response"])
```

    Logged Request:
    {'payload': {'data': {'ndarray': [[1, 2, 3, 4]]}, 'meta': {'puid': 'hello'}}, 'dataType': 'tabular', 'elements': {}, 'instance': [1.0, 2.0, 3.0, 4.0], 'ce-time': '2020-11-02T13:06:44.024402323Z', 'ce-source': 'http::8000'}
    
    Logged Response:
    {'ce-source': 'http::8000', 'instance': [0.0006985194531162841, 0.003668039039435755, 0.9956334415074478], 'names': ['t:0', 't:1', 't:2'], 'payload': {'data': {'names': ['t:0', 't:1', 't:2'], 'ndarray': [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, 'meta': {'puid': 'hello'}}, 'dataType': 'tabular', 'elements': {'t:1': [0.003668039039435755], 't:0': [0.0006985194531162841], 't:2': [0.9956334415074478]}, 'ce-time': '2020-11-02T13:06:44.042412223Z'}



```python
res = es.get(index="inference-log-seldon-seldon-multiclass-model-default", id="7983e38c-29dc-45ff-8ffb-77252e7ac86d")
print(res["_source"]["response"])
```

    {'ce-source': 'http::8000', 'instance': [0.0006985194531162841, 0.003668039039435755, 0.9956334415074478], 'names': ['t:0', 't:1', 't:2'], 'payload': {'data': {'names': ['t:0', 't:1', 't:2'], 'ndarray': [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, 'meta': {'puid': 'hello'}}, 'dataType': 'tabular', 'elements': {'t:1': [0.003668039039435755], 't:0': [0.0006985194531162841], 't:2': [0.9956334415074478]}, 'ce-time': '2020-11-02T13:06:44.042412223Z'}


### Send feedback

We can now send the correction, or the truth value of the prediction.

For this we'll need to send the UUID of the feedback request to ensure it's added to the correct index.


```python
puid_seldon_1 = pred_resp_1.headers.get("seldon-puid")

print(puid_seldon_1)
```

    7f9e69df-6601-4947-b30b-e05a12f5726b


We can also be able to add extra metadata, such as the user providing the feedback, date, time, etc.


```python
feedback_tags_1 = {
    "user": "Seldon Admin",
    "date": "11/07/2020"
}
```

And finally we can put together the feedback request.


```python
feedback_req_1 = {
    "reward": 0,
    "truth": {
        'data': {
            'names': ['t:0', 't:1', 't:2'], 
            'ndarray': [[0, 0, 1]]
        },
        "meta": {
            "tags": feedback_tags_1
        }
    }
}
```

And send the feedback request


```python
feedback_resp_1 = requests.post(f"{url}/feedback", json=feedback_req_1, headers={"seldon-puid": puid_seldon_1})
print(feedback_resp_1)
```

    <Response [200]>


Check that feedback has been received and stored in the Elasticsearch index


```python
res = es.search(index="inference-log-seldon-seldon-multiclass-model-default", body={"query": {"match_all": {}}})

print(res["hits"]["hits"][-1]["_source"]["feedback"])
```

    {'reward': 0, 'ce-source': 'http::8000', 'truth': {'data': {'names': ['t:0', 't:1', 't:2'], 'ndarray': [[0, 0, 1]]}, 'meta': {'tags': {'date': '11/07/2020', 'user': 'Seldon Admin'}}}, 'ce-time': '2020-11-02T18:26:22.03643351Z'}


### Deploying Metrics Server

Now we'll be able to see how the metrics server makes use of this infrastructure patterns to provide real time performance metrics.


```python
%%writefile config/multiclass-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: seldon-multiclass-model-metrics
  namespace: seldon-logs
  labels:
    app: seldon-multiclass-model-metrics
spec:
  replicas: 1
  selector:
    matchLabels:
      app: seldon-multiclass-model-metrics
  template:
    metadata:
      labels:
        app: seldon-multiclass-model-metrics
      annotations:
        prometheus.io/path: /v1/metrics
        prometheus.io/scrape: "true"
    spec:
      securityContext:
          runAsUser: 8888
      containers:
      - name: user-container
        image: seldonio/alibi-detect-server:1.7.0-dev
        imagePullPolicy: IfNotPresent
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
        - --elasticsearch_uri
        - https://admin:admin@elasticsearch-opendistro-es-client-service.seldon-logs:9200
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
---
apiVersion: v1
kind: Service
metadata:
  name: seldon-multiclass-model-metrics
  namespace: seldon-logs
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
!kubectl apply -f config/multiclass-deployment.yaml
```

    deployment.apps/seldon-multiclass-model-metrics configured
    service/seldon-multiclass-model-metrics unchanged



```python
!kubectl rollout status -n seldon-logs deploy/seldon-multiclass-model-metrics
```

    deployment "seldon-multiclass-model-metrics" successfully rolled out


### Trigger for metrics server

The trigger will be created in the seldon-logs namespace as that is where the initial trigger will be sent to.


```bash
%%bash

kubectl apply -f - << END
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: multiclass-model-metrics-trigger
  namespace: seldon-logs
spec:
  broker: default
  filter:
    attributes:
      inferenceservicename: multiclass-model
      type: io.seldon.serving.feedback
  subscriber:
    uri: http://seldon-multiclass-model-metrics.seldon-logs:80
END
```

    trigger.eventing.knative.dev/multiclass-model-metrics-trigger created



```python
import time
time.sleep(20)
```

### (Alternative) create kservice

If you want to create a kservice, and you've installed knative eventing and knative serving, you can use the instructions below.

The value of the file `config/multiclass-service.yaml` would be:
```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: seldon-multiclass-model-metrics
  namespace: seldon-logs
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        prometheus.io/path: /v1/metrics
        prometheus.io/scrape: "true"
    spec:
      containers:
      - image: "seldonio/alibi-detect-server:1.7.0-dev"
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

```bash
%%bash
kubectl apply -f - <<EOF
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: default
  namespace: seldon-logs
EOF
```

And then the trigger contents:

```
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: multiclass-model-metrics-trigger
  namespace: seldon-logs
spec:
  broker: default
  filter:
    attributes:
      inferenceservicename: multiclass-model
      type: io.seldon.serving.feedback
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: seldon-multiclass-model-metrics
```

And you can run it with:
```
kubectl apply -f config/trigger.yaml
```

### Confirm empty metrics


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -v -X GET "http://seldon-multiclass-model-metrics.seldon-logs.svc.cluster.local:80/v1/metrics" 
```

    
    
    
    
    
    
    
    
    
    
    
    


### Send Feedback


```python
feedback_resp_1 = requests.post(f"{url}/feedback", json=feedback_req_1, headers={"seldon-puid": puid_seldon_1})
print(feedback_resp_1)
```

    <Response [200]>


### Confirm metrics available


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -v -X GET "http://seldon-multiclass-model-metrics.seldon-logs.svc.cluster.local:80/v1/metrics" 
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    


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
