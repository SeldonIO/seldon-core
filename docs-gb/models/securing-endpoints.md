# Securing model endpoints

In enterprise use cases, you may need to control who can access the endpoints for deployed models or pipelines. You can leverage existing authentication mechanisms in your cluster or environment, such as service mesh-level controls, or use cloud provider solutions like Apigee on GCP, Amazon API Gateway on AWS, or a provider-agnostic gateway like Gravitee. Seldon Core 2 integrates with various service meshes that support these requirements. Though Seldon Core 2 is service-mesh agnostic, the example on this page demonstrates how to set up authentication and authorization to secure a model endpoint using the Istio service mesh.

## Securing Endpoints with Istio

Service meshes offer a flexible way of defining authentication and authorization rules for your models. With Istio, for example, you can configure multiple layers of security within an Istio Gateway, such as a [TLS for HTTPS at the gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-tls-ingress-gateway-for-a-single-host) level, [mutual TLS (mTLS) for secure internal communication](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway), as well as [AuthorizationPolicies](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [RequestAuthentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) policies to enforce both authentication and authorization controls.

**Prerequisites**
* [Deploy a model](../kubernetes/service-meshes/istio.md)
* [Configure a gateway](../kubernetes/service-meshes/istio.md)
* [Create a virtual service to expose the REST and gRPC endpoints](../kubernetes/service-meshes/istio.md)
* Configure a OIDC provider to authenticate. Obtain the `issuer` url, `jwksUri`, and the `Access token` from the OIDC provider.
{% hint style="info" %}
**Note** There are many types of authorization policies that you can configure to enable access control on workloads in the mesh. 
{% endhint %}

In the following example, you can secure the endpoint such that any requests to the endpoint without the access token are denied.

To secure the endpoints of a model, you need to:
1. Create a `RequestAuthentication` resource named `ingress-jwt-auth` in the `istio-system namespace`. Replace `<OIDC_TOKEN_ISSUER>` and `<OIDC_TOKEN_ISSUER_JWKS>` with your OIDC provider’s specific issuer URL and JWKS (JSON Web Key Set) URI.
   
```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: ingress-jwt-auth
  namespace: istio-system  # This is the namespace where Istio Ingress Gateway usually resides
spec:
  selector:
    matchLabels:
      istio: istio-ingressgateway  # Apply to Istio Ingress Gateway pods
  jwtRules:
    - issuer: <OIDC_TOKEN_ISSUER>
      jwksUri: <OIDC_TOKEN_ISSUER_JWKS>
```
Create the resource using `kubectl apply -f ingress-jwt-auth.yaml`.

2. Create an authorization policy `deny-empty-jwt` in the namespace `istio-system`.
 
```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-empty-jwt
  namespace: istio-system
spec:
  action: DENY
  rules:
    - from:
        - source:
            notRequestPrincipals:
              - '*'  # Denies requests without a valid JWT principal
      to:
        - operation:
            paths:
              - /v2/*  # Applies to requests with this path pattern
  selector:
    matchLabels:
      app: istio-ingressgateway  # Applies to Istio Ingress Gateway pods
```
Create the resource using `kubectl apply -f deny-empty-jwt.yaml`.

3. To verify that the requests without an access token are denied send this request:
   ```bash
    curl -i http://$MESH_IP/v2/models/iris/infer \
    -H "Content-Type: application/json" \
    -H "seldon-model":iris \
    -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
    ``` 
  The output is similar to:
  ```bash
  HTTP/1.1 403 Forbidden
  content-length: 19
  content-type: text/plain  
  date: Fri, 25 Oct 2024 11:14:33 GMT
  server: istio-envoy
  connection: close
  Closing connection 0
  RBAC: access denied
  ```
  Now, send the same request with an access token:
  ```bash
  curl -i http://$MESH_IP/v2/models/iris/infer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "seldon-model":iris \
  -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
  ```
  The output is similar to:
  ```bash
  HTTP/1.1 200 OK
  ce-endpoint: iris_1
  ce-id: 2fb8a086-ee22-4285-9826-9d38111cbb9e
  ce-inferenceservicename: mlserver
  ce-modelid: iris_1
  ce-namespace: seldon-mesh
  ce-requestid: 2fb8a086-ee22-4285-9826-9d38111cbb9e
  ce-source: io.seldon.serving.deployment.mlserver.seldon-mesh
  ce-specversion: 0.3
  ce-type: io.seldon.serving.inference.response
  content-length: 213
  content-type: application/json
  date: Fri, 25 Oct 2024 11:44:49 GMT
  server: envoy
  x-request-id: csdo9cbc2nks73dtlk3g
  x-envoy-upstream-service-time: 9
  x-seldon-route: :iris_1:
  ```

