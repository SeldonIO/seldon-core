# Request Logging

## Approach

Before setting up request logging, first look at [centralised logging](../README.md).

Request logging does not necessarily need the EFK stack used for centralised logging. However, this example will use elasticsearch.

The approach is:

1 Configure a seldon deployment to send the requests and responses of the HTTP traffic into a knative broker.
2 The broker sends these to a knative service for logging, called seldon-request-logger
3 seldon-request-logger processes and sends to elasticsearch

The seldon-request-logger enriches the raw message to optimise for searching.

The seldon-request-logger implementation is replaceable and the type of the message emitted by the SeldonDeployment can be adjusted to go a different logger.

## Setup

Local options are minikube or KIND. Create minikube cluster with knative recommendations for resource - https://knative.dev/docs/install/knative-with-minikube/. For KIND cluster use:
```
kind create cluster --config ../kind_config.yaml --image kindest/node:v1.15.6
```

For KIND or minikube run knative-setup-minikube.sh. 

Otherwise (for non-local) follow the [knative installation](https://knative.dev/docs/install/) for your cloud provider.

Run `kubectl apply -f seldon-request-logger.yaml`


Create broker:

```
kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker default
```

Note that SeldonDeployments configured to log requests will look for a broker in their namespace unless told otherwise. So this is assuming the SeldonDeployment will be in the default namespace.

And trigger:
```
kubectl apply -f ./trigger.yaml
```

## Running and Seeing logs

Follow the EFK minikube setup from [centralised logging guide](../README.md) except the step to deploy the model, which instead should be:
```
helm install seldon-single-model ../../../helm-charts/seldon-single-model/ --set model.logger="http://default-broker"
```

(If you've already installed then you can first remove with `helm delete seldon-single-model --purge` or do an upgrade instead of an install.)

This time when you install the loadtester, requests should get filtered through the to seldon-request-logger and from there to elastic.

If you filter to pods beginning 'seldon-request-logger' or containing the attribute 'sdepName' then you should see request-response pairs for 'seldon-single-model'.

## Notes

We filter out the istio-proxy, istio-init and queue-proxy containers with configuration in the fluentd values file.

Request logging is disabled by default in seldon deployments but can be enabled by engine environment variables. These are made available in the helm charts or can be set in the svcOrchSpec section of the SeldonDeployment.

Knative eventing supports alternative brokering implementations. This uses the default in-memory implementation but can be swapped for e.g. kafka.
To swap for kafka see https://github.com/knative/eventing/tree/master/contrib/kafka/config

For running on a cloud-provider cluster the helm charts and yaml config would need to be changed. In some places replicas and resources have been reduced to fit into minikube. For a cloud-provider cluster the upstream config defaults are more appropriate.

To add custom fields for tracking requests (e.g. an order or customer id), add this to the 'meta.tags' section of the SeldonMessage and use the [predict_raw method in the model implementation](https://docs.seldon.io/projects/seldon-core/en/latest/python/python_component.html?highlight=predict_raw#low-level-methods)

## Debugging

If hitting problems be sure to check all pods are up and scheduled and cluster has enough resources - `kubectl get pods --all-namespaces`

Check that a broker is available in the default namespace with 'kubectl get broker -n default'

You can check whether messages are going through the broker by checking the logs of the default-broker-filter pod. Each message should cause a log line there. If you don't see that pod at all then the broker is not up.

To manually send messages create an interactive session with a busybox pod:

`kubectl run curl --image=radial/busyboxplus:curl -i --tty --rm`

Then from that session you can run a curl direct to the broker (or any other k8s service) e.g.:

```
curl -v "http://default-broker.default.svc.cluster.local/" \
  -X POST \
  -H "X-B3-Flags: 1" \
  -H 'CE-SpecVersion: 0.2' \
  -H "CE-Type: io.seldon.serving.inference.request" \
  -H "CE-Time: 2018-04-05T03:56:24Z" \
  -H "CE-ID: 45a8b444-3213-4758-be3f-540bf93f85ff" \
  -H "CE-Source: dev.knative.example" \
  -H "Seldon-Puid: 1a" \
  -H 'Content-Type: application/json' \
  -d '{"data": {"names": ["f0", "f1"], "ndarray": [0.77, 0.63]}}'

  curl -v "http://default-broker.default.svc.cluster.local/" \
  -X POST \
  -H "X-B3-Flags: 1" \
  -H 'CE-SpecVersion: 0.2' \
  -H "CE-Type: io.seldon.serving.inference.response" \
  -H "CE-Time: 2018-04-05T03:56:24Z" \
  -H "CE-ID: 45a8b444-3213-4758-be3f-540bf93f85fg" \
  -H "CE-Source: dev.knative.example" \
  -H "Seldon-Puid: 1a" \
  -H 'Content-Type: application/json' \
  -d '{"data": {"names": ["proba"], "ndarray": [0.09826376903346358]}}'

```
