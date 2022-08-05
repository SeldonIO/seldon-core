## SSL Enabled Requests

Install seldon with ingress provider.



```bash
%%bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
kubectl create ns cert-manager
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v1.6.1 \
  --set installCRDs=true
```

    "jetstack" has been added to your repositories
    Hang tight while we grab the latest from your chart repositories...
    ...Successfully got an update from the "strimzi" chart repository
    ...Successfully got an update from the "kube-eagle" chart repository
    ...Successfully got an update from the "jetstack" chart repository
    ...Successfully got an update from the "incubator" chart repository
    ...Successfully got an update from the "astronomer" chart repository
    ...Successfully got an update from the "datawire" chart repository
    ...Successfully got an update from the "bitnami" chart repository
    ...Successfully got an update from the "stable" chart repository
    Update Complete. ⎈ Happy Helming!⎈ 
    NAME: cert-manager
    LAST DEPLOYED: Tue Aug 25 12:45:43 2020
    NAMESPACE: cert-manager
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    cert-manager has been deployed successfully!
    
    In order to begin issuing certificates, you will need to set up a ClusterIssuer
    or Issuer resource (for example, by creating a 'letsencrypt-staging' issuer).
    
    More information on the different types of issuers and how to configure them
    can be found in our documentation:
    
    https://cert-manager.io/docs/configuration/
    
    For information on how to configure cert-manager to automatically provision
    Certificates for Ingress resources, take a look at the `ingress-shim`
    documentation:
    
    https://cert-manager.io/docs/usage/ingress/


    Error from server (AlreadyExists): namespaces "cert-manager" already exists

#### Creating issuer so we can deploy certificates


```bash
%%bash
kubectl apply -f - << END
apiVersion: cert-manager.io/v1
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
    selfsigned-issuer   True    5s


#### Deploying self signed certificate


```bash
%%bash 
kubectl apply -f - << END
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: sklearn-default-cert
spec:
  dnsNames:
  - "*"
  issuerRef:
    name: selfsigned-issuer
  secretName: sklearn-default-cert
END
```

    certificate.cert-manager.io/sklearn-default-cert configured



```python
!kubectl get certificate
```

    NAME                   READY   SECRET                 AGE
    sklearn-default-cert   True    sklearn-default-cert   10s


#### Confirm the certificate has been created


```python
!kubectl get secret sklearn-default-cert 
```

    NAME                   TYPE                DATA   AGE
    sklearn-default-cert   kubernetes.io/tls   3      33s


#### Create a NON-SSL seldon core model


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.14.1/sklearn/iris
      name: classifier
    name: default
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/sklearn configured


#### And an SSL seldon core model


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-ssl
spec:
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.14.1/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    ssl:
      certSecretName: sklearn-default-cert 
END
```

    seldondeployment.machinelearning.seldon.io/sklearn-ssl created



```python
!kubectl get sdep
```

    NAME      AGE
    sklearn   11s


#### Send requests to the NON-SSL model

First we'll try sending a non-ssl request (which should work):


```bash
%%bash
kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -v "http://sklearn-default.default.svc.cluster.local:8000/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"hello"}}


And then we'll send an SSL request (which should fail)


```bash
%%bash
kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -k -v "https://sklearn-default.default.svc.cluster.local:8000/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}' || echo "done"
```

    * SSLv3, TLS handshake, Client hello (1):
    * Unknown SSL protocol error in connection to sklearn-default.default.svc.cluster.local:8000 
    curl: (35) Unknown SSL protocol error in connection to sklearn-default.default.svc.cluster.local:8000 
    done


    pod default/curl terminated (Error)


#### Send requests to the SSL-ENABLED model

First we send a non-ssl request (which should FAIL)


```bash
%%bash
kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -v "http://sklearn-ssl-default.default.svc.cluster.local:8000/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}' || echo "done"
```

    
    
    
    
    
    
    
    * Empty reply from server
    curl: (52) Empty reply from server
    done


    pod default/curl terminated (Error)


Then we send an SSL request (which should be SUCCESSFUL)


```bash
%%bash
kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -k -v "https://sklearn-ssl-grpc-default.default.svc.cluster.local:8000/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}' || echo "done"
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
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    


#### Test Ambassador Configured with SSL
You can install Ambassador by following the instructions in [the documentation](https://docs.seldon.io/projects/seldon-core/en/latest/ingress/ambassador.html#option-1-ambassador-api-gateway).

The external API would still be reachable without SSL. 

This means that both models (the SSL and NON-SSL) will be reachable through port 80 in the ambassador gateway.

This is because the Ambassador gateway establishes an SSL communication with the SSL-ENABLED model, and establishes a non-ssl communication with the NON-SSL-ENABLED model.

##### Testing the NON-SSL model


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -k -v "http://ambassador.ambassador.svc.cluster.local/seldon/default/sklearn/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    


Testing the SSL ENABLED Model


```python
!kubectl run --quiet=true -it --rm curl --image=radial/busyboxplus:curl --restart=Never -- \
    curl -X POST -k -v "http://ambassador.ambassador.svc.cluster.local/seldon/default/sklearn-ssl/api/v1.0/predictions" \
        -H "Content-Type: application/json" -d '{"data": { "ndarray": [[1,2,3,4]]}, "meta": { "puid": "hello" }}'
```

    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    
    


### Creating a GRPC model


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-ssl-grpc
spec:
  transport: grpc
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.14.1/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    ssl:
      certSecretName: sklearn-default-cert 
END
```

    seldondeployment.machinelearning.seldon.io/sklearn-ssl-grpc created


### Test the service directly
We can send a request directly to the service by portforwarding the service with the following command:
    
```
kubectl port-forward svc/sklearn-ssl-grpc-default 8000:8000
```

And then running the following request using the GRPCURL library (which you can download in their github page):

And finally we can send a request via the ambassador port as above:


```bash
%%bash
cd ../../../executor/proto && \
grpcurl \
         -rpc-header seldon:sklearn -rpc-header namespace:default \
        -d '{"data": {"ndarray": [[1,2,3,4]]}}' \
        -insecure -proto prediction.proto  localhost:8000 seldon.protos.Seldon/Predict
```

    {
      "meta": {
        
      },
      "data": {
        "names": [
          "t:0",
          "t:1",
          "t:2"
        ],
        "ndarray": [
            [
                  0.0006985194531162841,
                  0.003668039039435755,
                  0.9956334415074478
                ]
          ]
      }
    }



```python

```
