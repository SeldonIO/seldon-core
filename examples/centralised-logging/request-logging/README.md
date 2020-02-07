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
(Or for KIND only there is a full-kind-setup.sh script that runs steps.)

For KIND or minikube run:

```
./install_istio.sh
./install_knative.sh
```

Otherwise (for non-local) follow the [knative installation](https://knative.dev/docs/install/) for your cloud provider.

Ensure you run `kubectl get pods --all-namespaces` and wait for all knative-eventing pods to come up.

Run `kubectl apply -f seldon-request-logger.yaml`


Create broker:

```
kubectl label namespace default knative-eventing-injection=enabled
sleep 3
kubectl -n default get broker default
```

The broker should show 'READY' as True.

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

(If you've already installed then you can first remove with `helm delete seldon-single-model` or do an upgrade instead of an install.)

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

You can check whether messages are going through the broker by checking the logs of the default-broker-filter pod and/or checking whether seldon-request-logger pods get started. If you don't see a broker pod at all then the broker is not up.

If too many requests are going through for elastic to handle with its allocated memory you might hit https://github.com/elastic/elasticsearch/issues/31197

To manually send messages create an interactive session with a busybox pod:

`kubectl run curl --image=radial/busyboxplus:curl -i --tty --rm`

Then from that session you can run a curl direct to the broker (or any other k8s service) e.g. the below - you may want to set the CE-Time to your current time:

```
curl -v "http://default-broker.default.svc.cluster.local/" \
  -X POST \
  -H "X-B3-Flags: 1" \
  -H 'CE-SpecVersion: 1.0' \
  -H "CE-Type: io.seldon.serving.inference.request" \
  -H "CE-Time: 2020-02-06T16:41:24Z" \
  -H "CE-ID: 45a8b444-3213-4758-be3f-540bf93f85ff" \
  -H "CE-Source: dev.knative.example" \
  -H "Ce-Inferenceservicename: seldon-model" \
  -H "Ce-Modelid: classifier" \
  -H "Ce-Namespace: default" \
  -H "Ce-Predictor: example" \
  -H "Ce-Requestid: 11226d35-9679-40fe-a21b-d22bb55fbe0e" \
  -H 'Content-Type: application/json' \
  -d '{"data": {"names": ["f0", "f1"], "ndarray": [0.77, 0.63]}}'

  curl -v "http://default-broker.default.svc.cluster.local/" \
  -X POST \
  -H "X-B3-Flags: 1" \
  -H 'CE-SpecVersion: 1.0' \
  -H "CE-Type: io.seldon.serving.inference.response" \
  -H "CE-Time: 2020-02-06T16:41:24Z" \
  -H "CE-ID: 45a8b444-3213-4758-be3f-540bf93f85fg" \
  -H "CE-Source: dev.knative.example" \
  -H "Ce-Inferenceservicename: seldon-model" \
  -H "Ce-Modelid: classifier" \
  -H "Ce-Namespace: default" \
  -H "Ce-Predictor: example" \
  -H "Ce-Requestid: 11226d35-9679-40fe-a21b-d22bb55fbe0e" \
  -H 'Content-Type: application/json' \
  -d '{"data": {"names": ["proba"], "ndarray": [0.09826376903346358]}}'

```
