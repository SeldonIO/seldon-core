# Centralised Logging Example

## Introduction

Here we will set up EFK (elasticsearch, fluentd/fluentbit, kibana) as a stack to gather logs from SeldonDeployments and make them searchable.

This demo is aimed at minikube.

Alternatives are available and if you are running in cloud then you can consider a managed service from your cloud provider.

If you just want to bootstrap a full logging and request tracking setup for minikube, run ./full-setup.sh. That includes the [request logging setup](./request-logging/README.md)

## Setup

If helm is not already set up then it needs to be configured:

```
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

Install elasticsearch with minikube configuration:

```
helm install --name elasticsearch elasticsearch --version 7.1.1 --namespace=logs -f elastic-minikube.yaml --repo https://helm.elastic.co
```

Then fluentd as a collection agent (chosen in preference to fluentbit - see notes at end):

```
helm install fluentd-elasticsearch --name fluentd --namespace=logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
```

And kibana UI:

```
helm install kibana --version 7.1.1 --name=kibana --namespace=logs --set service.type=NodePort --repo https://helm.elastic.co
```

## Generating Logging

First we need seldon and a seldon deployment.

Install seldon operator:

```
helm install --name seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system
```

Check that it now recognises the seldon CRD by running `kubectl get sdep`.

Now a model:

```
helm install --name seldon-single-model ../../helm-charts/seldon-single-model/ --set engine.env.LOG_MESSAGES_EXTERNALLY="false"
```

And the loadtester:

```
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite
helm install --name seldon-core-loadtesting ../../helm-charts/seldon-core-loadtesting/ --set locust.host=http://seldon-single-model-seldon-single-model:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=0 --set locust.maxWait=0 --set replicaCount=1
```

## Inspecting Logging and Search for Requests

To find kibana URL

```
echo $(minikube ip)":"$(kubectl get svc kibana-kibana -n logs -o=jsonpath='{.spec.ports[?(@.port==5601)].nodePort}')
```

When Kibana appears for the first time there will be a brief animation while it initializes.
On the Welcome page click Explore on my own.
From the top-left or from the `Visualize and Explore Data` panel select the `Discover` item.
In the form field Index pattern enter logstash-*
It should read "Success!" and Click the `> Next` step button on the right.
In the next form select timestamp from the dropdown labeled `Time Filter` field name.
From the bottom-right of the form select `Create index pattern`.
In a moment a list of fields will appear.
From the top-left or the home screen's `Visualize and Explore Data` panel, select the `Discover` item.
The log list will appear.
Refine the list a bit by selecting `log` near the bottom the left-hand Selected fields list.
When you hover over or click on the word `log`, click the `Add` button to the right of the label.
You can create a filter using the `Add Filter` button under `Search`. The field can be `kubernetes.labels.seldon-app` and the value can be an 'is' match on `seldon-single-model-seldon-single-model`.

The custom fields in the request bodies may not currently be in the index. If you hover over one in a request you may see `No cached mapping for this field`.

To add mappings, go to `Management` at the bottom-left and then `Index Patterns`. Hit `Refresh` on the index created earlier. The number of fields should increase and `request.data.names` should be present.

Now we can go back and add a further filter for `data.names` with the operator `exists`. We can add further filters if we want, such as the presence of a feature name or the presence of a feature value.

![picture](./kibana-custom-search.png)

## Notes

The fluentd setup is configured to ensure only labelled pods are logged and seldon pods are automatically labelled.

Fluentbit can be chosen instead. This could be installed with:

```
helm install stable/fluent-bit --name=fluent-bit --namespace=logs --set backend.type=es --set backend.es.host=elasticsearch-master
```

In that case pods would be logged. At the time of writing fluentbit only supports [excluding pods by label, not including](https://github.com/fluent/fluent-bit/issues/737).

Seldon can also be used to log full HTTP requests. See [request logging guide](./request-logging/README.md)

The elasticsearch backend is not available externally by default but can be exposed if needed for debugging with `kubectl patch svc elasticsearch-master -n logs -p '{"spec": {"type": "LoadBalancer"}}'`

## Credits

Loosely based on https://www.katacoda.com/javajon/courses/kubernetes-observability/efk
Fluentd filtering based on https://blog.ptrk.io/tweaking-an-efk-stack-on-kubernetes/