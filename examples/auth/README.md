# Authentication and Authorization for Seldon Core Requests

## Introduction

This is an example of setting up auth for seldon core model deployments in an istio enabled kubernetes cluster. Here we discuss the following topics:
  - Authentication at Seldon Deployment Component Level
  - Authorization based on user id token claims
  - Authentication at the Ingress Level

## Authentication Demo

### Prerequisites
- Kubernetes Cluster (~1.14)
- Istio 1.5
- Helm v3

### Demo setup

1. Install and setup Istio 1.5 with istioctl as mentioned in the [docs](https://istio.io/docs/setup/getting-started/). For this demo we have used the demo profile as shown.
```
istioctl manifest apply --set profile=demo
```

2. Create namespace `foo` and setup an istio gateway

```
kubectl create namespace foo

kubectl label namespace foo istio-injection=enabled

kubectl apply -f seldon-gateway.yaml
```

3. Create a `seldon-system` namespace and install Seldon Core using Helm. Also add gateway location and enable istio by setting the helm values as shown.
(Using a custom operator image because of https://github.com/istio/istio/issues/22246)

```
kubectl create namespace seldon-system
helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --namespace seldon-system \
    --set istio.enabled=true \
    --set istio.gateway="seldon-gateway.foo.svc.cluster.local" \
    --set image.repository="sachinmv31/seldon-core-operator" \
    --set image.tag="1.1.1-SNAPSHOT"
```


4. Deploy an iris model by applying the seldon manifest shown below,

```
kubectl apply -f iris.yaml
```

5. Make a prediction via the ingress gateway created.
```sh
export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -X POST http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json'     -d '{ "data": { "ndarray": [[1,2,3,4]] } }'
```
- Response
```
{"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{}}
```

## Authentication at Seldon Deployment Level

1. Setup request Authentication as follow:
```
kubectl apply -f - <<EOF
apiVersion: "security.istio.io/v1beta1"
kind: "RequestAuthentication"
metadata:
  name: seldon-auth-example
  namespace: foo
spec:
  selector:
    matchLabels:
      seldon-deployment-id : iris-model
  jwtRules:
  - issuer: "testing@secure.istio.io"
    jwksUri: "https://raw.githubusercontent.com/istio/istio/release-1.5/security/tools/jwt/samples/jwks.json"
    outputPayloadToHeader: "Seldon-Core-User"
EOF
```

2. Verify thant invalid token requests get blocked and returns a `401 Unauthorized` status code
```sh
curl -X POST http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json'  -H "Authorization: Bearer invalidToken"   -d '{ "data": { "ndarray": [[1,2,3,4]] } }' -o /dev/null -s -w "%{http_code}"
```
- Response
```
401 Unauthorized
```

3. Setup an istio authorization policy
```
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: seldon-auth-example
  namespace: foo
spec:
  selector:
    matchLabels:
      seldon-deployment-id : iris-model
  action: ALLOW
  rules:
  - from:
    - source:
       requestPrincipals: ["testing@secure.istio.io/testing@secure.istio.io"]
EOF
```

4. Verify that requests without a token get blocked and returns a `401 Unauthorized` status code
```sh
curl -X POST http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json' -d '{ "data": { "ndarray": [[1,2,3,4]] } }' -o /dev/null -s -w "%{http_code}"
```
- Response
```
403 Forbidden
```

5. Fetch a token and verify that requests with valid token passes through
```sh
TOKEN=$(curl https://raw.githubusercontent.com/istio/istio/release-1.5/security/tools/jwt/samples/demo.jwt -s) && echo $TOKEN | cut -d '.' -f2 - | base64 --decode -
```
- Response
```json
{"exp":4685989700,"foo":"bar","iat":1532389700,"iss":"testing@secure.istio.io","sub":"testing@secure.istio.io"}
```

```sh
curl -X POST http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN"  -d '{ "data": { "ndarray": [[1,2,3,4]] } }' -o /dev/null -s -w "%{http_code}"
```
- Response
```
200 OK
```

6. Also we can deny all invalid requests (without a decodable token) by adding the following rule to the authorization policy

```yaml
  action: DENY
  rules:
  - from:
    - source:
       notRequestPrincipals: ["*"]
```

7. This can be extended to Seldon deployment component level by selecting specific components by matching labels. Further the seldon core executor/engine can base64 decode the user claims from the header `Seldon-Core-User` as configured in the RequestAuthentication with as outputPayloadToHeader.


## Authorization based on user id token claims

1. Setup and authorization policy to enable only users in group `group1` to make predictions

```
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: seldon-auth-example
  namespace: foo
spec:
  selector:
    matchLabels:
      seldon-deployment-id : iris-model
  action: ALLOW
  rules:
  - from:
    - source:
       requestPrincipals: ["testing@secure.istio.io/testing@secure.istio.io"]
    when:
    - key: request.auth.claims[groups]
      values: ["group1"]
EOF
```

2. Verify that the user not in the group valid token
```sh
curl -X POST http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN"  -d '{ "data": { "ndarray": [[1,2,3,4]] } }' -o /dev/null -s -w "%{http_code}"
```
- Response
```
403 Forbidden
```


3. Fetch a new user token which has `group1` in the group claims and verify that the request is allowed
```sh
TOKEN_WITH_GROUP=$(curl https://raw.githubusercontent.com/istio/istio/release-1.5/security/tools/jwt/samples/groups-scope.jwt -s) && echo $TOKEN_WITH_GROUP | cut -d '.' -f2 - | base64 --decode -
```
- Response
```json
{"exp":3537391104,"groups":["group1","group2"],"iat":1537391104,"iss":"testing@secure.istio.io","scope":["scope1","scope2"],"sub":"testing@secure.istio.io"}
```

```sh
curl -X POST http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN_WITH_GROUP"  -d '{ "data": { "ndarray": [[1,2,3,4]] } }' -o /dev/null -s -w "%{http_code}"
```
- Response
```
200 OK
```

4. Cleanup

```
kubectl -n foo delete authorizationpolicies.security.istio.io  seldon-auth-example
kubectl -n foo delete requestauthentication seldon-auth-example
```

## Authentication at the Ingress Level

Similarly, you can also setup RequestAuthentication and AuthorizationPolicy at the ingress level by changing the selector


1. Setup request Authentication as follow:
```
kubectl apply -f - <<EOF
apiVersion: "security.istio.io/v1beta1"
kind: "RequestAuthentication"
metadata:
  name: seldon-auth-example
  namespace: istio-system
spec:
  selector:
    matchLabels:
       istio: ingressgateway
  jwtRules:
  - issuer: "testing@secure.istio.io"
    jwksUri: "https://raw.githubusercontent.com/istio/istio/release-1.5/security/tools/jwt/samples/jwks.json"
    outputPayloadToHeader: "Seldon-Core-User"
EOF
```

2. Here is an example of an Authorization policy that denies DELETE methods to the `/seldon` path


```
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: seldon-auth-example
  namespace: istio-system
spec:
  selector:
    matchLabels:
       istio: ingressgateway
  action: DENY
  rules:
  - to:
    - operation:
        methods: ["DELETE"]
        paths: ["/seldon"]     
EOF
```

3. Verify that the method is not allowed forthe request
```sh
curl -X DELETE http://$INGRESS_HOST/seldon/foo/iris-model/api/v1.0/predictions     -H 'Content-Type: application/json' -d '{ "data": { "ndarray": [[1,2,3,4]] } }' -o /dev/null -s -w "%{http_code}"
```
- Response
```
405 Method Not Allowed
```


4. Cleanup

```
kubectl -n istio-system delete authorizationpolicies.security.istio.io  seldon-auth-example
kubectl -n istio-system delete requestauthentication seldon-auth-example

kubectl delete namespace foo
```

## Relevant Links

- https://istio.io/docs/reference/config/security/request_authentication/

- https://istio.io/docs/reference/config/security/authorization-policy/

- https://istio.io/docs/tasks/security/authorization/authz-jwt/
