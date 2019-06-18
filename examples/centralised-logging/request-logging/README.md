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

And trigger:
```
kubectl apply -f ./trigger.yaml
```

## Running and Seeing logs

Follow the EFK minikube setup from [centralised logging guide](../README.md).

This time when you install the loadtester, requests should get filtered through the to seldon-request-logger and from there to elastic.

If you filter to pods beginning 'seldon-request-logger' or containing the attribute 'sdepName' then you should see request-response pairs for 'seldon-single-model'.

## Notes

We filter out the istio-proxy, istio-init and queue-proxy containers with configuration in the fluentd values file.

TODO: say how to turn on or off req logging via env vars

Knative eventing supports alternative brokering implementations. This uses the default in-memory implementation but can be swapped for e.g. kafka.
To swap for kafka see https://github.com/knative/eventing/tree/master/contrib/kafka/config

For running on a cloud-provider cluster the helm charts and yaml config would need to be changed. In some places replicas and resources have been reduced to fit into minikube. For a cloud-provider cluster the upstream config defaults are more appropriate.

To add custom fields for tracking requests (e.g. an order or customer id), add this to the 'meta.tags' section of the SeldonMessage and use the [predict_raw method in the model implementation](https://docs.seldon.io/projects/seldon-core/en/latest/python/python_component.html?highlight=predict_raw#low-level-methods)