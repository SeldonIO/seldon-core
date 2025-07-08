# Graph Examples

Port-forward to that ingress on localhost:8003 in a separate terminal either with:

  * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
  * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080`



```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Input and output transformer with model in same pod


```python
!cat tin-model-tout.yaml
```


```python
!kubectl create -f tin-model-tout.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f tin-model-tout.yaml
```

## Input and output transformer with model in separate pods


```python
!cat tin-model-tout-sep-pods.yaml
```


```python
!kubectl create -f tin-model-tout-sep-pods.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f tin-model-tout-sep-pods.yaml
```

## Input and output transformer with svcOrch in separate pod


```python
!cat tin-model-tout-sep-svcorch.yaml
```


```python
!kubectl create -f tin-model-tout-sep-svcorch.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f tin-model-tout-sep-svcorch.yaml
```

## Combiner sperate pods


```python
!cat combiner-sep-pods.yaml
```


```python
!kubectl create -f combiner-sep-pods.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f combiner-sep-pods.yaml
```

## Combiner


```python
!cat combiner.yaml
```


```python
!kubectl create -f combiner.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f combiner.yaml
```

## Combiner seperate pods prepack server


```python
!cat combiner-prepack-sep-pods.yaml
```


```python
!kubectl create -f combiner-prepack-sep-pods.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0, 1.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f combiner-sep-pods.yaml
```

## Combiner prepack server same pod


```python
!cat combiner-prepack.yaml
```


```python
!kubectl create -f combiner-prepack.yaml
```


```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```


```python
!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0, 1.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/test/api/v1.0/predictions \
   -H "Content-Type: application/json"
```


```python
!kubectl delete -f combiner-sep-pods.yaml
```


```python

```
