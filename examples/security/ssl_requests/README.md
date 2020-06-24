## SSL Enabled Requests

Install seldon with ingress provider.



```bash
%%bash
kubectl create ns cert-manager
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v0.15.1 \
  --set installCRDs=true
```

#### Deploying self signed certificate


```bash
%%bash
kubectl apply -f - << END
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: default
spec:
  selfSigned: {}
END
```

    issuer.cert-manager.io/selfsigned-issuer created



```python
!kubectl get issuer
```

    NAME                READY   AGE
    selfsigned-issuer   True    23s



```bash
%%bash 
kubectl apply -f - << END
apiVersion: cert-manager.io/v1alpha3
kind: Certificate
metadata:
  name: sklearn-default-cert
spec:
  dnsNames:
  - example.com
  issuerRef:
    name: selfsigned-issuer
  secretName: sklearn-default-cert
END
```

    certificate.cert-manager.io/sklearn-default-cert created



```python
!kubectl get certificate
```

    NAME                   READY   SECRET                 AGE
    sklearn-default-cert   True    sklearn-default-cert   10s



```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    ssl:
      certSecretName: sklearn-default-cert 
END
```

    seldondeployment.machinelearning.seldon.io/sklearn configured



```python
!kubectl get sdep
```

    NAME                   READY   SECRET                 AGE
    sklearn-default-cert   True    sklearn-default-cert   3m1s



```python
### Send a request from inside cluster
```


```bash
%%bash
kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -k -v "https://sklearn-default.default.svc.cluster.local:8000/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
```

    * SSLv3, TLS handshake, Client hello (1):
    * SSLv3, TLS handshake, Server hello (2):
    * SSLv3, TLS handshake, CERT (11):
    * SSLv3, TLS handshake, Server key exchange (12):
    * SSLv3, TLS handshake, Server finished (14):
    * SSLv3, TLS handshake, Client key exchange (16):
    * SSLv3, TLS change cipher, Client hello (1):
    * SSLv3, TLS handshake, Finished (20):
    * SSLv3, TLS change cipher, Client hello (1):
    * SSLv3, TLS handshake, Finished (20):
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"hello"}}



```python
!kubectl delete gateway -n istio-system seldon-gateway || true
```

    gateway.networking.istio.io "seldon-gateway" deleted



```bash
%%bash
kubectl apply -n istio-system -f - << END
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: seldon-gateway
spec:
  selector:
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
      mode: PASSTHROUGH
  - hosts:
    - '*'
    port:
      name: https2
      number: 8443
      protocol: HTTPS
    tls:
      mode: PASSTHROUGH
END
```

    gateway.networking.istio.io/seldon-gateway created



```bash
%%bash
kubectl apply -f - << END
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: sklearn-https
spec:
  gateways:
  - istio-system/seldon-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /seldon/default/sklearn/
    rewrite:
      uri: /
    route:
    - destination:
        host: sklearn-default
        port:
          number: 8000
        subset: default
  tls:
  - match:
    - port: 443
      sniHosts:
      - '*'
    route:
    - destination:
        host: sklearn-default
        port:
          number: 8000
        subset: default
END
```

    virtualservice.networking.istio.io/sklearn-https created



```bash
%%bash
curl -v -X POST -H 'Content-Type: application/json' \
    -d '{"data": {"ndarray": [[1,2,3,4]]}}' \
    http://localhost:80/seldon/default/sklearn/api/v1.0/predictions
```

    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{}}


    Note: Unnecessary use of -X or --request, POST is already inferred.
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                     Dload  Upload   Total   Spent    Left  Speed
      0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 127.0.0.1...
    * TCP_NODELAY set
    * Connected to localhost (127.0.0.1) port 80 (#0)
    > POST /seldon/default/sklearn/api/v1.0/predictions HTTP/1.1
    > Host: localhost
    > User-Agent: curl/7.58.0
    > Accept: */*
    > Content-Type: application/json
    > Content-Length: 34
    > 
    } [34 bytes data]
    * upload completely sent off: 34 out of 34 bytes
    < HTTP/1.1 200 OK
    < content-type: application/json
    < seldon-puid: 394bfaf4-7e46-481e-8f99-b5b4f29ad297
    < x-content-type-options: nosniff
    < date: Tue, 23 Jun 2020 11:32:16 GMT
    < content-length: 125
    < x-envoy-upstream-service-time: 4
    < server: envoy
    < 
    { [125 bytes data]
    100   159  100   125  100    34   8928   2428 --:--:-- --:--:-- --:--:-- 11357
    * Connection #0 to host localhost left intact



```bash
%%bash
cd ../../../executor/proto && \
grpcurl \
         -rpc-header seldon:sklearn -rpc-header namespace:default \
        -d '{"data": {"ndarray": []}}' \
         -proto prediction.proto  127.0.0.1:443 seldon.protos.Seldon/Predict
```


```python

```
