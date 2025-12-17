# Stateful Model Feedback Metrics Server
In this example we will add statistical performance metrics capabilities by levering the Seldon metrics server.

Dependencies
* Seldon Core installed
* Ingress provider (Istio or Ambassador)

An easy way is to run `examples/centralized-logging/full-kind-setup.sh` and then:
```bash
    helm delete seldon-core-loadtesting
    helm delete seldon-single-model
```
    
Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

Ambassador:

    kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080

Istio:

    kubectl port-forward -n istio-system svc/istio-ingressgateway 8003:80





```python
!kubectl create namespace seldon || echo "namespace already created"
```

    Error from server (AlreadyExists): namespaces "seldon" already exists
    namespace already created



```python
!mkdir -p config
```


```python
from IPython.core.magic import register_line_cell_magic

@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, 'w') as f:
        f.write(cell.format(**globals()))
```


```python
VERSION=!cat ../../../version.txt
VERSION=VERSION[0]
VERSION
```




    '1.19.0-dev'



### Create a simple model
We create a multiclass classification model - iris classifier.

The iris classifier takes an input array, and returns the prediction of the 4 classes.

The prediction can be done as numeric or as a probability array.


```python
%%writetemplate config/multiclass-model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: multiclass-model
  namespace: seldon
spec:
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v{VERSION}/sklearn/iris
      name: classifier
      logger:
        url: http://seldon-multiclass-model-metrics.seldon.svc.cluster.local:80/
        mode: all
    name: default
    replicas: 1
```


```python
!kubectl apply -f config/multiclass-model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/multiclass-model created



```python
!kubectl wait sdep/multiclass-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/multiclass-model condition met


#### Send test request


```python
from tenacity import retry, stop_after_delay, wait_exponential
import json

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
        res=!curl -X POST "http://localhost:8003/seldon/seldon/multiclass-model/api/v1.0/predictions" \
                -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
        print(res)
        j=json.loads(res[-1])
        assert(len(j["data"]["ndarray"][0])==3)

predict()
```

    ['  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current', '                                 Dload  Upload   Total   Spent    Left  Speed', '', '  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0', '100   268  100   204  100    64  14369   4507 --:--:-- --:--:-- --:--:-- 19142', '{"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"hello","requestPath":{"classifier":"seldonio/sklearnserver:1.19.0-dev"}}}']


### Metrics Server
You can create a kubernetes deployment of the metrics server with this:


```python
%%writetemplate config/multiclass-deployment.yaml
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
      labels:
        app: seldon-multiclass-model-metrics
    spec:
      securityContext:
          runAsUser: 8888
      containers:
      - name: user-container
        image: seldonio/alibi-detect-server:{VERSION}
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
        - MetricsServer
        env:
        - name: "SELDON_DEPLOYMENT_ID"
          value: "multiclass-model"
        - name: "PREDICTIVE_UNIT_ID"
          value: "classifier"
        - name: "PREDICTIVE_UNIT_IMAGE"
          value: "seldonio/alibi-detect-server:{VERSION}"
        - name: "PREDICTOR_ID"
          value: "default"
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


```python
!kubectl apply -f config/multiclass-deployment.yaml -n seldon
```

    deployment.apps/seldon-multiclass-model-metrics unchanged
    service/seldon-multiclass-model-metrics unchanged



```python
!kubectl rollout status deploy/seldon-multiclass-model-metrics -n seldon
```

    deployment "seldon-multiclass-model-metrics" successfully rolled out



```python
import time

time.sleep(20)
```

### Send feedback


```python

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def send_feedback():
        res=!curl -X POST "http://localhost:8003/seldon/seldon/multiclass-model/api/v1.0/feedback" \
                -H "Content-Type: application/json" \
                -d '{"response": {"data": {"ndarray": [[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]}}, "truth":{"data": {"ndarray": [[0,0,1]]}}}'
        print(res)
        import json
        j=json.loads(res[-1])
        assert("data" in j)

send_feedback()
```

    ['  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current', '                                 Dload  Upload   Total   Spent    Left  Speed', '', '  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0', '100   252  100   108  100   144  18960  25280 --:--:-- --:--:-- --:--:-- 50400', '{"data":{"tensor":{"shape":[0]}},"meta":{"requestPath":{"classifier":"seldonio/sklearnserver:1.19.0-dev"}}}']



```python
import time

time.sleep(3)
```

### Check that metrics are recorded


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def check_logs():
    res=!kubectl logs -n seldon $(kubectl get pods -n seldon -l app=seldon-multiclass-model-metrics \
                        -n seldon -o jsonpath='{.items[0].metadata.name}') | grep "PROCESSING Feedback Event"
    print(res)
    assert(len(res)>0)

check_logs()
```

    ['INFO:root:PROCESSING Feedback Event.', 'INFO:root:PROCESSING Feedback Event.']


### Cleanup


```python
!kubectl delete -n seldon -f config/multiclass-deployment.yaml
```

    deployment.apps "seldon-multiclass-model-metrics" deleted
    service "seldon-multiclass-model-metrics" deleted



```python
!kubectl delete -n seldon sdep multiclass-model
```

    seldondeployment.machinelearning.seldon.io "multiclass-model" deleted

