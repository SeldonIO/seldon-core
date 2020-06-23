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

    seldondeployment.machinelearning.seldon.io/sklearn created



```python
!kubectl get sdep
```

    NAME                   READY   SECRET                 AGE
    sklearn-default-cert   True    sklearn-default-cert   3m1s



```bash
%%bash
curl -v -X POST -H 'Content-Type: application/json' \
    -d '{"data": {"ndarray": [[1,2,3,4]]}}' \
    https://localhost:80/seldon/default/sklearn/api/v1.0/predictions
```


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
