# Securing model endpoints

You can secure the endpoints of a model that you deployed in a Kubernetes cluster using a service mesh. You can configure multiple layers of security within an Istio Gateway. For instance, you can configure [TLS for HTTPS at the gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-tls-ingress-gateway-for-a-single-host) level, enable [mutual TLS (mTLS) to secure internal communication](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway), and apply [AuthorizationPolicies](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [RequestAuthentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) policies to enforce both authentication and authorization controls.

## Prerequaites
* [Deploy a model]
* [Configure a gateway]
* [Create a virtual service to expose the REST and gRPC endpoints]
* Configure a OIDC provider to authenticate `https://$MESH_IP/v2`. Obtain the `issuer` url, `jwksUri`, and the `Access token` from the OIDC provider.

In the following example, you can secure the endpoint such that any requests to the end point without the access token are denied.

To secure the enpoints of a model, you need to:
1. Create a RequestAuthentication resource `ingress-jwt-auth` in the namespace `istio-system`.
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
      - issuer: "https://fc4dba59-f6ea-4f05-9fd2-37ff194947ba.app.skycloak.io/realms/core2"
        jwksUri: "https://fc4dba59-f6ea-4f05-9fd2-37ff194947ba.app.skycloak.io/realms/core2/protocol/openid-connect/certs"
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
     curl -i http://$MESH_IP/models/iris/infer \
    -H "Content-Type: application/json" \
    -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
    ``` 
  The output is similar to:
  ```bash

  ```
  Now, send the same request with an access token:
  ```bash
  curl -i http://34.90.95.128/v2/models/iris/infer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
  ```
  The output is similar to:
  ```bash

  ```

