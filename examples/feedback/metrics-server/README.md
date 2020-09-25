# Stateful Model Feedback Metrics Server
In this example we will add statistical performance metrics capabilities by levering the Seldon metrics server.

Dependencies
* Seldon Core installed
* KNative eventing v0.11.0
* KNative serving v0.11.1 (optional)



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
        mode: all
    name: default
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/multiclass-model configured


#### Send test request


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -v "http://multiclass-model-default.default.svc.cluster.local:8000/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    


### Create Metrics Service


```python
%%writefile config/multiclass-service.yaml
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: seldon-multiclass-model-metrics
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
      - image: "seldonio/alibi-detect-server:1.3.0-dev"
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
          value: "alibi-detect-server:1.3.0-dev"
        - name: "PREDICTOR_ID"
          value: "default"
        securityContext:
            runAsUser: 8888
```

    Overwriting config/multiclass-service.yaml



```python
!kubectl apply -f config/multiclass-service.yaml
```

    service.serving.knative.dev/multiclass-model-metrics-kservice created



```python
!kubectl get kservice
```

    NAME                                URL                                                            LATESTCREATED                             LATESTREADY   READY   REASON
    multiclass-model-metrics-kservice   http://multiclass-model-metrics-kservice.default.example.com   multiclass-model-metrics-kservice-8nh2p                 False   RevisionMissing


### (Alternative) Kubernetes Deployment
Alternatively you can also create a kubernetes deployment instead of a kservice with the yaml below.


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
      labels:
        app: seldon-multiclass-model-metrics
    spec:
      securityContext:
          runAsUser: 8888
      containers:
      - name: user-container
        image: seldonio/alibi-detect-server:1.3.0-dev
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
          value: "alibi-detect-server:1.3.0-dev"
        - name: "PREDICTOR_ID"
          value: "default"
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
!kubectl apply -f config/multiclass-deployment.yaml
```

    deployment.apps/seldon-multiclass-model-metrics created
    service/seldon-multiclass-model-metrics created



```python
!kubectl get pods
```

    NAME                                               READY   STATUS        RESTARTS   AGE
    seldon-multiclass-model-metrics-5f9776bf69-25dxk   1/1     Running       0          20s
    seldon-multiclass-model-metrics-5f9776bf69-55jzn   1/1     Terminating   0          10m


### Create Trigger


```python
!kubectl label namespace default knative-eventing-injection=enabled --overwrite=true
```

    namespace/default not labeled



```python
!kubectl get broker
```

    NAME      READY   REASON   URL                                               AGE
    default   True             http://default-broker.default.svc.cluster.local   2m53s



```python
%%writefile config/trigger.yaml
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

    Overwriting config/trigger.yaml



```python
!kubectl apply -f config/trigger.yaml
```

    trigger.eventing.knative.dev/multiclass-model-metrics-trigger created



```python
!kubectl get trigger
```

    NAME                               READY   REASON   BROKER    SUBSCRIBER_URI                                                      AGE
    multiclass-model-metrics-trigger   True             default   http://seldon-multiclass-model-metrics.default.svc.cluster.local/   1s


### Send feedback


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -v "http://multiclass-model-default.default.svc.cluster.local:8000/api/v1.0/feedback" \
        -H "Content-Type: application/json" \
        -d '{"response": {"data": {"ndarray": [[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]}}, "truth":{"data": {"ndarray": [[0,0,1]]}}}'
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    


### Check that metrics are recorded


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X GET -v "http://seldon-multiclass-model-metrics.default.svc.cluster.local:80/v1/metrics"
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    



```python

```
