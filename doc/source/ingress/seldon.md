# Seldon OAuth Gateway

Seldon provides an example OAuth gateway you can use. We recommend however that you utilize for production gateway solutions such as Ambassador or istio.

## Install Seldon OAuth Gateway

You can install the Seldon OAuth gateway using Helm:

```bash
helm install helm-charts/seldon-core-oauth-gateway --name seldon-gateway --repo https://storage.googleapis.com/seldon-charts
```

## Provide OAuth Credentials

Provide OAuth credentials for your deployments when creating them. You should add `oauth_key` and `oauth_secret` values to your resource. For example:

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model
spec:
  name: test-deployment
  oauth_key: oauth-key
  oauth_secret: oauth-secret
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              memory: 1Mi
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1
```

## Serve Requests

To serve requests from your running deployment via the gateway follow the instructions [here](../workflow/serving.html#api-oauth-gateway)