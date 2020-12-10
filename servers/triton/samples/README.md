# Triton examples

## Prerequisties

Install Ambassador and port-forward to 

## Simple Model

```
kubectl create -f simple.yaml
```

Curl request.

```
curl -v -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'    -X POST http://localhost:8003/seldon/default/triton/v2/models/simple/infer    -H "Content-Type: application/json"
```
