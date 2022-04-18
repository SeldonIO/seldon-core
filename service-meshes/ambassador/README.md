## Ambassador Notebook Example

Assumes

 * You have installed emissary as per their docs
 
 Tested with
 
 emissary-ingress-7.3.2 insatlled via [helm](https://www.getambassador.io/docs/emissary/latest/tutorials/getting-started/)


```python
INGRESS_IP=!kubectl get svc emissary-ingress -n emissary -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
INGRESS_IP=INGRESS_IP[0]
import os
os.environ['INGRESS_IP'] = INGRESS_IP
INGRESS_IP
```




    '172.21.255.1'



### Ambassador Single Model Example


```python
!kustomize build config/single-model
```

    apiVersion: getambassador.io/v3alpha1
    kind: Host
    metadata:
      name: wildcard
      namespace: seldon-mesh
    spec:
      hostname: '*'
      requestPolicy:
        insecure:
          action: Route
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Listener
    metadata:
      name: emissary-ingress-listener-8080
      namespace: seldon-mesh
    spec:
      hostBinding:
        namespace:
          from: ALL
      port: 8080
      protocol: HTTP
      securityModel: INSECURE
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: iris-grpc
      namespace: seldon-mesh
    spec:
      add_request_headers:
        seldon-model:
          value: iris
      grpc: true
      hostname: '*'
      prefix: /inference.GRPCInferenceService
      rewrite: ""
      service: seldon-mesh:80
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: iris-http
      namespace: seldon-mesh
    spec:
      add_request_headers:
        seldon-model:
          value: iris
      hostname: '*'
      prefix: /v2/
      rewrite: ""
      service: seldon-mesh:80
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      requirements:
      - sklearn
      storageUri: gs://seldon-models/mlserver/iris



```python
!kustomize build config/single-model | kubectl apply --validate=false -f -
```

    host.getambassador.io/wildcard created
    listener.getambassador.io/emissary-ingress-listener-8080 created
    mapping.getambassador.io/iris-grpc created
    mapping.getambassador.io/iris-http created
    model.mlops.seldon.io/iris created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

    model.mlops.seldon.io/iris condition met



```python
!curl -v http://${INGRESS_IP}/v2/models/iris/infer -H "Content-Type: application/json"\
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
    < date: Sat, 16 Apr 2022 15:45:43 GMT
    < server: envoy
    < x-envoy-upstream-service-time: 792
    < seldon-route: iris_1
    < 
    * Connection #0 to host 172.21.255.1 left intact
    {"model_name":"iris_1","model_version":"1","id":"72ac79f5-b355-4be3-b8c5-2ebedaa39f60","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}


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

    host.getambassador.io "wildcard" deleted
    listener.getambassador.io "emissary-ingress-listener-8080" deleted
    mapping.getambassador.io "iris-grpc" deleted
    mapping.getambassador.io "iris-http" deleted
    model.mlops.seldon.io "iris" deleted


### Traffic Split Two Models

***Currently not working due to this [issue](https://github.com/emissary-ingress/emissary/issues/4062)***



```python
!kustomize build config/traffic-split
```

    apiVersion: getambassador.io/v3alpha1
    kind: Host
    metadata:
      name: wildcard
      namespace: seldon-mesh
    spec:
      hostname: '*'
      requestPolicy:
        insecure:
          action: Route
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Listener
    metadata:
      name: emissary-ingress-listener-8080
      namespace: seldon-mesh
    spec:
      hostBinding:
        namespace:
          from: ALL
      port: 8080
      protocol: HTTP
      securityModel: INSECURE
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: iris1-grpc
      namespace: seldon-mesh
    spec:
      add_request_headers:
        seldon-model:
          value: iris1
      grpc: true
      hostname: '*'
      prefix: /inference.GRPCInferenceService
      rewrite: ""
      service: seldon-mesh:80
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: iris1-http
      namespace: seldon-mesh
    spec:
      add_request_headers:
        seldon-model:
          value: iris1
      add_response_headers:
        seldon_model:
          value: iris1
      hostname: '*'
      prefix: /v2
      rewrite: ""
      service: seldon-mesh:80
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: iris2-grpc
      namespace: seldon-mesh
    spec:
      add_request_headers:
        seldon-model:
          value: iris2
      grpc: true
      hostname: '*'
      prefix: /inference.GRPCInferenceService
      rewrite: ""
      service: seldon-mesh:80
      weight: 50
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: iris2-http
      namespace: seldon-mesh
    spec:
      add_request_headers:
        seldon-model:
          value: iris2
      add_response_headers:
        seldon_model:
          value: iris2
      hostname: '*'
      prefix: /v2
      rewrite: ""
      service: seldon-mesh:80
      weight: 50
    ---
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



```python
!kustomize build config/traffic-split | kubectl apply -f -
```

    host.getambassador.io/wildcard created
    listener.getambassador.io/emissary-ingress-listener-8080 created
    mapping.getambassador.io/iris1-grpc created
    mapping.getambassador.io/iris1-http created
    mapping.getambassador.io/iris2-grpc created
    mapping.getambassador.io/iris2-http created
    model.mlops.seldon.io/iris1 created
    model.mlops.seldon.io/iris2 created



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
    < date: Sat, 16 Apr 2022 15:46:17 GMT
    < server: envoy
    < x-envoy-upstream-service-time: 920
    < seldon-route: iris2_1
    < seldon_model: iris2
    < 
    * Connection #0 to host 172.21.255.1 left intact
    {"model_name":"iris2_1","model_version":"1","id":"ed521c32-cd85-4cb8-90eb-7c896803f271","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}


```python
!grpcurl -d '{"model_name":"iris1","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
    -plaintext \
    -import-path ../../apis \
    -proto ../../apis/mlops/v2_dataplane/v2_dataplane.proto \
    ${INGRESS_IP}:80 inference.GRPCInferenceService/ModelInfer
```

    {
      "modelName": "iris2_1",
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

    host.getambassador.io "wildcard" deleted
    listener.getambassador.io "emissary-ingress-listener-8080" deleted
    mapping.getambassador.io "iris1-grpc" deleted
    mapping.getambassador.io "iris1-http" deleted
    mapping.getambassador.io "iris2-grpc" deleted
    mapping.getambassador.io "iris2-http" deleted
    model.mlops.seldon.io "iris1" deleted
    model.mlops.seldon.io "iris2" deleted

