# Distributed Tracing with Jaeger

Illustrate the configuration for allowing distributed tracing using Jaeger.

## Setup Seldon Core

Install Seldon Core as described in [docs](../install/installation.md)

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

* Ambassador:

`kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`

* Istio:

`kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:80`

```python
!kubectl create namespace seldon
```

```
Error from server (AlreadyExists): namespaces "seldon" already exists
```

```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

```
Context "kind-ansible" modified.
```

## Install Jaeger

Follow the Jaeger docs to [install on Kubernetes](https://www.jaegertracing.io/docs/1.38/operator/).

```python
!kubectl create namespace observability
!kubectl create -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.38.0/jaeger-operator.yaml -n observability
```

```
namespace/observability created
customresourcedefinition.apiextensions.k8s.io/jaegers.jaegertracing.io created
serviceaccount/jaeger-operator created
role.rbac.authorization.k8s.io/jaeger-operator created
rolebinding.rbac.authorization.k8s.io/jaeger-operator created
deployment.apps/jaeger-operator created
```

```python
!pygmentize simplest.yaml
```

```
[94mapiVersion[39;49;00m: jaegertracing.io/v1
[94mkind[39;49;00m: Jaeger
[94mmetadata[39;49;00m:
  [94mname[39;49;00m: simplest
[94mspec[39;49;00m:
  [94magent[39;49;00m:
    [94mstrategy[39;49;00m: DaemonSet
```

```python
!kubectl apply -f simplest.yaml
```

```
jaeger.jaegertracing.io/simplest created
```

Port forward to Jaeger UI

```bash
kubectl port-forward $(kubectl get pods -l app.kubernetes.io/name=simplest -n seldon -o jsonpath='{.items[0].metadata.name}') 16686:16686 -n seldon
```

## Run Example REST Deployment

```python
!pygmentize deployment_rest.yaml
```

```
[94mapiVersion[39;49;00m: machinelearning.seldon.io/v1
[94mkind[39;49;00m: SeldonDeployment
[94mmetadata[39;49;00m:
  [94mname[39;49;00m: tracing-example
  [94mnamespace[39;49;00m: seldon
[94mspec[39;49;00m:
  [94mname[39;49;00m: tracing-example
  [94mpredictors[39;49;00m:
  - [94mcomponentSpecs[39;49;00m:
    - [94mspec[39;49;00m:
        [94mcontainers[39;49;00m:
        - [94menv[39;49;00m:
          - [94mname[39;49;00m: TRACING
            [94mvalue[39;49;00m: [33m'[39;49;00m[33m0[39;49;00m[33m'[39;49;00m
          - [94mname[39;49;00m: JAEGER_AGENT_HOST
            [94mvalueFrom[39;49;00m:
              [94mfieldRef[39;49;00m:
                [94mfieldPath[39;49;00m: status.hostIP
          - [94mname[39;49;00m: JAEGER_AGENT_PORT
            [94mvalue[39;49;00m: [33m'[39;49;00m[33m5775[39;49;00m[33m'[39;49;00m
          - [94mname[39;49;00m: JAEGER_SAMPLER_TYPE
            [94mvalue[39;49;00m: const
          - [94mname[39;49;00m: JAEGER_SAMPLER_PARAM
            [94mvalue[39;49;00m: [33m'[39;49;00m[33m1[39;49;00m[33m'[39;49;00m
          [94mimage[39;49;00m: seldonio/mock_classifier:1.9.0-dev
          [94mname[39;49;00m: model1
        [94mterminationGracePeriodSeconds[39;49;00m: 1
    [94mgraph[39;49;00m:
      [94mchildren[39;49;00m: []
      [94mendpoint[39;49;00m:
        [94mtype[39;49;00m: REST
      [94mname[39;49;00m: model1
      [94mtype[39;49;00m: MODEL
    [94mname[39;49;00m: tracing
    [94mreplicas[39;49;00m: 1
    [94msvcOrchSpec[39;49;00m:
      [94menv[39;49;00m:
      - [94mname[39;49;00m: TRACING
        [94mvalue[39;49;00m: [33m'[39;49;00m[33m1[39;49;00m[33m'[39;49;00m
      - [94mname[39;49;00m: JAEGER_AGENT_HOST
        [94mvalueFrom[39;49;00m:
          [94mfieldRef[39;49;00m:
            [94mfieldPath[39;49;00m: status.hostIP
      - [94mname[39;49;00m: JAEGER_AGENT_PORT
        [94mvalue[39;49;00m: [33m'[39;49;00m[33m5775[39;49;00m[33m'[39;49;00m
      - [94mname[39;49;00m: JAEGER_SAMPLER_TYPE
        [94mvalue[39;49;00m: const
      - [94mname[39;49;00m: JAEGER_SAMPLER_PARAM
        [94mvalue[39;49;00m: [33m'[39;49;00m[33m1[39;49;00m[33m'[39;49;00m
```

```python
!kubectl create -f deployment_rest.yaml
```

```
seldondeployment.machinelearning.seldon.io/tracing-example created
```

```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=tracing-example -o jsonpath='{.items[0].metadata.name}')
```

```
deployment "tracing-example-tracing-0-model1" successfully rolled out
```

```python
!curl -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/tracing-example/api/v1.0/predictions \
   -H "Content-Type: application/json"
```

```
{"data":{"names":["proba"],"ndarray":[[0.43782349911420193]]},"meta":{"requestPath":{"model1":"seldonio/mock_classifier:1.9.0-dev"}}}
```

Check the Jaeger UI. You should be able to find traces like below:

![rest](../.gitbook/assets/jaeger-ui-rest-example.png)

```python
!kubectl delete -f deployment_rest.yaml
```

```
seldondeployment.machinelearning.seldon.io "tracing-example" deleted
```

## Run Example GRPC Deployment

```python
!pygmentize deployment_grpc.yaml
```

```python
!kubectl create -f deployment_grpc.yaml
```

```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=tracing-example -o jsonpath='{.items[0].metadata.name}')
```

```python
!cd ../../../executor/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' \
         -rpc-header seldon:tracing-example -rpc-header namespace:seldon \
         -plaintext \
         -proto ./prediction.proto  0.0.0.0:8003 seldon.protos.Seldon/Predict
```

Check the Jaeger UI. You should be able to find traces like below:

![grpc](../.gitbook/assets/jaeger-ui-grpc-example.png)

```python
!kubectl delete -f deployment_grpc.yaml
```

```python
!kubectl delete -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/crds/jaegertracing.io_jaegers_crd.yaml
!kubectl delete -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/service_account.yaml
!kubectl delete -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/role.yaml
!kubectl delete -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/role_binding.yaml
!kubectl delete -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/operator.yaml
!kubectl delete namespace observability
```

```python
```
