## Istio Notebook Examples

Assumes

 * You have installed istio as per their docs
 * You have exposed the ingressgateway as an external loadbalancer
 
 
tested with:

```
istioctl version
1.13.2

istioctl install --set profile=demo -y
```


```python
INGRESS_IP=!kubectl get svc istio-ingressgateway -n istio-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
INGRESS_IP=INGRESS_IP[0]
import os
os.environ['INGRESS_IP'] = INGRESS_IP
INGRESS_IP
```




    '172.21.255.1'



### Istio Single Model Example


```python
!kustomize build config/single-model
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      requirements:
      - sklearn
      storageUri: gs://seldon-models/mlserver/iris
    ---
    apiVersion: networking.istio.io/v1beta1
    kind: Gateway
    metadata:
      name: seldon-gateway
      namespace: seldon-mesh
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
      - hosts:
        - '*'
        port:
          name: http
          number: 80
          protocol: HTTP
      - hosts:
        - '*'
        port:
          name: https
          number: 443
          protocol: HTTPS
        tls:
          mode: SIMPLE
          privateKey: /etc/istio/ingressgateway-certs/tls.key
          serverCertificate: /etc/istio/ingressgateway-certs/tls.crt
    ---
    apiVersion: networking.istio.io/v1beta1
    kind: VirtualService
    metadata:
      name: iris-route
      namespace: seldon-mesh
    spec:
      gateways:
      - istio-system/seldon-gateway
      hosts:
      - '*'
      http:
      - match:
        - uri:
            prefix: /v2
        name: iris-http
        route:
        - destination:
            host: seldon-mesh.seldon-mesh.svc.cluster.local
          headers:
            request:
              set:
                seldon-model: iris
      - match:
        - uri:
            prefix: /inference.GRPCInferenceService
        name: iris-grpc
        route:
        - destination:
            host: seldon-mesh.seldon-mesh.svc.cluster.local
          headers:
            request:
              set:
                seldon-model: iris



```python
!kustomize build config/single-model | kubectl apply -f -
```

    model.mlops.seldon.io/iris unchanged
    gateway.networking.istio.io/seldon-gateway unchanged
    virtualservice.networking.istio.io/iris-route configured



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

    model.mlops.seldon.io/iris condition met



```python
!curl -v http://${INGRESS_IP}/v2/models/iris/infer -H "Content-Type: application/json" \
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

    *   Trying 172.21.255.1...
    * Connected to 172.21.255.1 (172.21.255.1) port 80 (#0)
    > POST /v2/models/iris/infer HTTP/1.1
    > Host: 172.21.255.1
    > User-Agent: curl/7.47.0
    > Accept: */*
    > Content-Type: application/json
    > Content-Length: 94
    > 
    * upload completely sent off: 94 out of 94 bytes
    < HTTP/1.1 200 OK
    < content-length: 196
    < content-type: application/json
    < date: Sat, 16 Apr 2022 15:34:11 GMT
    < server: istio-envoy
    < x-envoy-upstream-service-time: 802
    < seldon-route: iris_1
    < 
    * Connection #0 to host 172.21.255.1 left intact
    {"model_name":"iris_1","model_version":"1","id":"83520c4a-c7f1-4363-9bfd-60c5d8ee2dc5","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}


```python
!grpcurl -d '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
    -plaintext \
    -import-path ../../apis \
    -proto ../../apis/mlops/v2_dataplane/v2_dataplane.proto \
    ${INGRESS_IP}:80 inference.GRPCInferenceService/ModelInfer
```

    {
      "modelName": "iris_1",
      "modelVersion": "1",
      "outputs": [
        {
          "name": "predict",
          "datatype": "INT64",
          "shape": [
            "1"
          ],
          "contents": {
            "int64Contents": [
              "2"
            ]
          }
        }
      ]
    }



```python
!kustomize build config/single-model | kubectl delete -f -
```

    model.mlops.seldon.io "iris" deleted
    gateway.networking.istio.io "seldon-gateway" deleted
    virtualservice.networking.istio.io "iris-route" deleted


### Traffic Split Two Models


```python
!kustomize build config/traffic-split
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris1
      namespace: seldon-mesh
    spec:
      requirements:
      - sklearn
      storageUri: gs://seldon-models/mlserver/iris
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
      namespace: seldon-mesh
    spec:
      requirements:
      - sklearn
      storageUri: gs://seldon-models/mlserver/iris
    ---
    apiVersion: networking.istio.io/v1beta1
    kind: Gateway
    metadata:
      name: seldon-gateway
      namespace: seldon-mesh
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
      - hosts:
        - '*'
        port:
          name: http
          number: 80
          protocol: HTTP
    ---
    apiVersion: networking.istio.io/v1beta1
    kind: VirtualService
    metadata:
      name: iris-route
      namespace: seldon-mesh
    spec:
      gateways:
      - seldon-gateway
      hosts:
      - '*'
      http:
      - match:
        - uri:
            prefix: /v2/models/iris
        name: iris-http
        route:
        - destination:
            host: seldon-mesh.seldon-mesh.svc.cluster.local
          headers:
            request:
              set:
                seldon-model: iris1
          weight: 50
        - destination:
            host: seldon-mesh.seldon-mesh.svc.cluster.local
          headers:
            request:
              set:
                seldon-model: iris2
          weight: 50
      - match:
        - uri:
            prefix: /inference.GRPCInferenceService
        name: iris-grpc
        route:
        - destination:
            host: seldon-mesh.seldon-mesh.svc.cluster.local
          headers:
            request:
              set:
                seldon-model: iris1
          weight: 50
        - destination:
            host: seldon-mesh.seldon-mesh.svc.cluster.local
          headers:
            request:
              set:
                seldon-model: iris2
          weight: 50



```python
!kustomize build config/traffic-split | kubectl apply -f -
```

    model.mlops.seldon.io/iris1 created
    model.mlops.seldon.io/iris2 created
    gateway.networking.istio.io/seldon-gateway created
    virtualservice.networking.istio.io/iris-route created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

    model.mlops.seldon.io/iris1 condition met
    model.mlops.seldon.io/iris2 condition met



```python
!curl -v http://${INGRESS_IP}/v2/models/iris/infer -H "Content-Type: application/json" \
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

    *   Trying 172.21.255.1...
    * Connected to 172.21.255.1 (172.21.255.1) port 80 (#0)
    > POST /v2/models/iris/infer HTTP/1.1
    > Host: 172.21.255.1
    > User-Agent: curl/7.47.0
    > Accept: */*
    > Content-Type: application/json
    > Content-Length: 94
    > 
    * upload completely sent off: 94 out of 94 bytes
    < HTTP/1.1 200 OK
    < content-length: 197
    < content-type: application/json
    < date: Sat, 16 Apr 2022 15:35:01 GMT
    < server: istio-envoy
    < x-envoy-upstream-service-time: 801
    < seldon-route: iris1_1
    < 
    * Connection #0 to host 172.21.255.1 left intact
    {"model_name":"iris1_1","model_version":"1","id":"b54e6d8c-d253-4bb9-bb64-02c2ee49e89f","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}


```python
!grpcurl -d '{"model_name":"iris1","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
    -plaintext \
    -import-path ../../apis \
    -proto ../../apis/mlops/v2_dataplane/v2_dataplane.proto \
    ${INGRESS_IP}:80 inference.GRPCInferenceService/ModelInfer
```

    {
      "modelName": "iris1_1",
      "modelVersion": "1",
      "outputs": [
        {
          "name": "predict",
          "datatype": "INT64",
          "shape": [
            "1"
          ],
          "contents": {
            "int64Contents": [
              "2"
            ]
          }
        }
      ]
    }



```python
!kustomize build config/traffic-split | kubectl delete -f -
```

    model.mlops.seldon.io "iris1" deleted
    model.mlops.seldon.io "iris2" deleted
    gateway.networking.istio.io "seldon-gateway" deleted
    virtualservice.networking.istio.io "iris-route" deleted



```python

```
