# Request Logging

## Approach

Before setting up request logging, first look at [centralised logging](../README.md).

Request logging does not necessarily need the EFK stack used for centralised logging. However, this example will use it.

The approach is:

1 Configure a seldon deployment to send the request-response pairs of the HTTP traffic into a knative broker.
2 The broker sends this a knative service for logging, called seldon-request-logger
3 Fluentd picks up the logged message and feeds to elasticsearch

The seldon-request-logger enriches the raw message to optimise for searching.

The seldon-request-logger implementation is replaceable and the type of the message emitted by the SeldonDeployment can be adjusted (by an env var) to go a different logger.

## Setup

Create minikube cluster with knative recommendations for resource - https://knative.dev/v0.3-docs/install/knative-with-minikube/

Run knative-setup.sh

Run `kubectl apply -f seldon-message-logger.yaml`


Create broker:

```
kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker default
```

Create service:

```
kubectl apply -f knative-example-service.yaml
```

And trigger:
```
kubectl apply -f ./trigger.yaml
```

## Running and Seeing logs

Follow the EFK minikube setup from [centralised logging guide](../README.md).

This time when you install the loadtester, requests should get filtered through the to seldon-request-logger and from there to elastic.

If you filter to pods beginning 'seldon-request-logger' or containing the attribute 'sdepName' then you should see request-response pairs for 'seldon-single-model'.

## Notes

TODO: should we filter out the istio-proxy, istio-init and queue-proxy containers? They log much more than the user-container. Fluentd filtering might be:

```
<filter kubernetes.**>
    @type grep
    <exclude>
        key $.kubernetes.container_name
        pattern istio-proxy
    </exclude>
</filter>
```

TODO: say how to turn on or off via env vars

TODO: note on kafka etc