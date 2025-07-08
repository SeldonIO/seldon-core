# Scaling Examples

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * curl
 * grpcurl
 * pygmentize
 

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html) to setup Seldon Core with an ingress - either Ambassador or Istio.

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

 * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080`


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Replica Settings

A deployment that illustrate the settings for

  * `.spec.replicas`
  * `.spec.predictors[].replicas`
  * `.spec.predictors[].componentSpecs[].replicas`


Below you can see a configuration file that outlines these spec components mentioned (and different replicas):


```python
%%writefile resources/model_replicas.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: test-replicas
spec:
  replicas: 1
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier2
      replicas: 3
    graph:
      endpoint:
        type: REST
      name: classifier
      type: MODEL
      children:
      - name: classifier2
        type: MODEL
        endpoint:
          type: REST
    name: example
    replicas: 2
    traffic: 50
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier3
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier3
      type: MODEL
    name: example2
    traffic: 50
```

    Overwriting resources/model_replicas.yaml



```python
!kubectl create -f resources/model_replicas.yaml
```

We can now wait until each of the models are fully deployed


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=test-replicas -o jsonpath='{.items[0].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=test-replicas -o jsonpath='{.items[1].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=test-replicas -o jsonpath='{.items[2].metadata.name}')
```

Check each container is running in a deployment with correct number of replicas


```python
classifierReplicas = !kubectl get deploy test-replicas-example-0-classifier -o jsonpath='{.status.replicas}'
classifierReplicas = int(classifierReplicas[0])
assert classifierReplicas == 2
```


```python
classifier2Replicas = !kubectl get deploy test-replicas-example-1-classifier2 -o jsonpath='{.status.replicas}'
classifier2Replicas = int(classifier2Replicas[0])
assert classifier2Replicas == 3
```


```python
classifier3Replicas = !kubectl get deploy test-replicas-example2-0-classifier3 -o jsonpath='{.status.replicas}'
classifier3Replicas = int(classifier3Replicas[0])
assert classifier3Replicas == 1
```

We can now just send a simple request


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test-replicas/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f resources/model_replicas.yaml
```

## Scale SeldonDeployment

Now we can actually scale the seldon deployment and see how it actually scales.

First we want to deploy a simple model with a single replica:


```python
%%writefile resources/model_scale.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-scale
spec:
  replicas: 1  
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: example
```

    Overwriting resources/model_scale.yaml



```python
!kubectl create -f resources/model_scale.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-scale -o jsonpath='{.items[0].metadata.name}')
```

We can actually confirm that there is only 1 replica currently running


```python
replicas = !kubectl get deploy seldon-scale-example-0-classifier -o jsonpath='{.status.replicas}'
replicas = int(replicas[0])
assert replicas == 1
```

And then we can actually see how the model can be scaled up


```python
!kubectl scale --replicas=2 sdep/seldon-scale
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-scale -o jsonpath='{.items[0].metadata.name}')
```

And now we can verify that there are actually two replicas instead of 1


```python
replicas = !kubectl get deploy seldon-scale-example-0-classifier -o jsonpath='{.status.replicas}'
replicas = int(replicas[0])
assert replicas == 2
```

And now when we send requests to the model, these get directed to the respective replica.


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/seldon-scale/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f resources/model_scale.yaml
```


```python

```
