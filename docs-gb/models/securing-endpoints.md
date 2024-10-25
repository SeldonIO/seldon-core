# Securing model endpoints

You can secure the endpoints of a model that you deployed in a Kubernetes cluster using a service mesh. You can configure multiple layers of security within an Istio Gateway. For instance, you can configure [TLS for HTTPS at the gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-tls-ingress-gateway-for-a-single-host) level, enable [mutual TLS (mTLS) to secure internal communication](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway), and apply [AuthorizationPolicies](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [RequestAuthentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) policies to enforce both authentication and authorization controls.

## Prerequaites
* [Deploy a model](/kubernetes/service-meshes/istio.md)
* [Configure a gateway](/kubernetes/service-meshes/istio.md)
* [Create a virtual service to expose the REST and gRPC endpoints](/kubernetes/service-meshes/istio.md)
* Configure a OIDC provider to authenticate. Obtain the `issuer` url, `jwksUri`, and the `Access token` from the OIDC provider.

In the following example, you can secure the endpoint such that any requests to the end point without the access token are denied.

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

2. Create Authetication policy `deny-empty-jwt` in the namespace `istio-system`.
   ```yaml
   apiVersion: security.istio.io/v1beta1
    kind: AuthorizationPolicy
    metadata:
      name: core-v2-ingress
      namespace: istio-system
    spec:
      action: DENY
      rules:
      - from:
        - source:
            notRequestPrincipals:
            - '*'
        to:
          - operation:
              paths:
                 - /v2/*    
      selector:
        matchLabels:
          app: istio-ingressgateway  # Applies to Istio Ingress Gateway pods
    ``` 
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

