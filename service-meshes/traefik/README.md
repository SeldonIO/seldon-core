## Traefik Examples

Assumes

 * You have installed Traefik as per their [docs](https://doc.traefik.io/traefik/getting-started/install-traefik/#use-the-helm-chart) into namespace traefik-v2
 
 Tested with traefik-10.19.4



```python
INGRESS_IP=!kubectl get svc traefik -n traefik-v2 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
INGRESS_IP=INGRESS_IP[0]
import os
os.environ['INGRESS_IP'] = INGRESS_IP
INGRESS_IP
```




    '172.21.255.1'



### Traefik Single Model Example


```python
!kustomize build config/single-model
```

    apiVersion: v1
    kind: Service
    metadata:
      name: myapps
      namespace: seldon-mesh
    spec:
      ports:
      - name: web
        port: 80
        protocol: TCP
      selector:
        app: traefik-ingress-lb
      type: LoadBalancer
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
    ---
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      entryPoints:
      - web
      routes:
      - kind: Rule
        match: PathPrefix(`/`)
        middlewares:
        - name: iris-header
        services:
        - name: seldon-mesh
          port: 80
          scheme: h2c
    ---
    apiVersion: traefik.containo.us/v1alpha1
    kind: Middleware
    metadata:
      name: iris-header
      namespace: seldon-mesh
    spec:
      headers:
        customRequestHeaders:
          seldon-model: iris



```python
!kustomize build config/single-model | kubectl apply -f -
```

    service/myapps created
    model.mlops.seldon.io/iris created
    ingressroute.traefik.containo.us/iris created
    middleware.traefik.containo.us/iris-header created



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
    < Content-Length: 196
    < Content-Type: application/json
    < Date: Sat, 16 Apr 2022 15:53:27 GMT
    < Seldon-Route: iris_1
    < Server: envoy
    < X-Envoy-Upstream-Service-Time: 895
    < 
    * Connection #0 to host 172.21.255.1 left intact
    {"model_name":"iris_1","model_version":"1","id":"0dccf477-78fa-4a11-92ff-4d7e4f1cdda8","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}


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

    service "myapps" deleted
    model.mlops.seldon.io "iris" deleted
    ingressroute.traefik.containo.us "iris" deleted
    middleware.traefik.containo.us "iris-header" deleted



```python

```
