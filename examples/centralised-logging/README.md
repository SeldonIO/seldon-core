# Centralised Logging Example

## Introduction

Here we will set up EFK (elasticsearch, fluentbit, kibana) as a stack to gather logs from SeldonDeployments and make them searchable.

This demo is aimed at minikube and loosely based on https://www.katacoda.com/javajon/courses/kubernetes-observability/efk

Alternatives are available and if you are running in cloud then you can consider a managed service from your cloud provider.

## Setup

If helm is not already set up then it needs to be configured:

```
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

First setup elasticsearch helm repo:

```
helm repo add elastic https://helm.elastic.co
```

Next install elasticsearch with minikube configuration:

```
helm install --name elasticsearch elastic/elasticsearch --version 7.1.0 --namespace=logs -f elastic-minikube.yaml
```

Then fluentbit as a collection agent:

```
helm install stable/fluent-bit --name=fluent-bit --namespace=logs --set backend.type=es --set backend.es.host=elasticsearch-master
```

And kibana UI:

```
helm install elastic/kibana --version 7.1.0 --name=kibana --namespace=logs --set service.type=NodePort
```

## Generating Logging

First we need seldon and a seldon deployment.

Install seldon operator:

```
helm install seldon-core-operator --name seldon-core --namespace seldon-system --repo https://storage.googleapis.com/seldon-charts
```

Now a model:

```
helm install seldon-single-model --name seldon-single-model --repo https://storage.googleapis.com/seldon-charts
```

And the loadtester:

```
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite
helm install seldon-core-loadtesting --name seldon-core-loadtesting --repo https://storage.googleapis.com/seldon-charts --set locust.host=http://seldon-single-model-seldon-single-model:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=0 --set locust.maxWait=0 --set replicaCount=1
```

## Inspecting Logging and Search for Requests

To find kibana URL

```
echo $(minikube ip)":"$(kubectl get svc kibana-kibana -n logs -o=jsonpath='{.spec.ports[?(@.port==5601)].nodePort}')
```

When Kibana appears for the first time there will be a brief animation while it initializes.
On the Welcome page click Explore on my own.
From the top-left or from the `Visualize and Explore Data` panel select the `Discover` item.
In the form field Index pattern enter kubernetes_cluster-*
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

To add mappings, go to `Management` at the bottom-left and then `Index Patterns`. Hit `Refresh` on the index created earlier. The number of fields should increase and `data.names` should be present.

Now we can go back and add a further filter for `data.names` with the operator `exists`. We can add further filters if we want, such as the presence of a feature name or the presence of a feature value.

![picture](./kibana-custom-search.png)

## Notes

All pods will be logged. To exclude pods see https://docs.fluentbit.io/manual/filter/kubernetes#request-to-exclude-logs

In the future we may need to find a way to transform requests so that data-points are searchable due to https://discuss.elastic.co/t/query-array-by-position/124765