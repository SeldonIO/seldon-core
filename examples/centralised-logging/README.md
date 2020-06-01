# Centralised and Request Logging Example

Centralized logging means pulling pod logs and bringing together in a single place - elasticsearch.

Request logging means also logging the http requests and responses in elasticsearch.

## Introduction

Here we will set up EFK (elasticsearch, fluentd/fluentbit, kibana) as a stack to gather logs from SeldonDeployments and make them searchable.

This demo is aimed at KIND or minikube but can also work with a cloud provider. Uses helm v3.

Either run through step-by-step or use full-kind-setup.sh.

## Setup Elastic - KIND

Start cluster

```
kind create cluster --config kind_config.yaml --image kindest/node:v1.15.6
```

Install elastic with KIND config:

```
kubectl create namespace seldon-logs
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
helm install elasticsearch elasticsearch --version 7.6.0 --namespace=seldon-logs -f elastic-kind.yaml --repo https://helm.elastic.co --set image=docker.elastic.co/elasticsearch/elasticsearch-oss
```

## Setup Elastic - Minikube

Start Minikube with flags as shown:

```
minikube start --cpus 6 --memory 10240 --disk-size=30g --kubernetes-version='1.15.0'
```

Install elasticsearch with minikube configuration:

```
kubectl create namespace seldon-logs
helm install elasticsearch elasticsearch --version 7.6.0 --namespace=seldon-logs -f elastic-minikube.yaml --repo https://helm.elastic.co --set image=docker.elastic.co/elasticsearch/elasticsearch-oss
```

## Fluentd and Kibana

Then fluentd as a collection agent (chosen in preference to fluentbit - see notes at end):

```
helm install fluentd fluentd-elasticsearch --version 8.0.0 --namespace=seldon-logs -f fluentd-values.yaml --repo https://kiwigrid.github.io
```

And kibana UI:

```
helm install kibana kibana --version 7.6.0 --namespace=seldon-logs --set service.type=NodePort --repo https://helm.elastic.co --set image=docker.elastic.co/kibana/kibana-oss
```



## Setting Up Model

First we need seldon and a seldon deployment.

Install seldon operator:

```
kubectl create namespace seldon-system

helm install seldon-core ../../helm-charts/seldon-core-operator/ --namespace seldon-system
```

Check that it now recognises the seldon CRD by running `kubectl get sdep`.

Now a model:

```
helm install seldon-single-model ../../helm-charts/seldon-single-model/ --set model.logger.enabled=true --set model.logger.url="http://default-broker.seldon-logs"
```

## Setting up Request Logging

The approach is:

1 Configure a seldon deployment to send the requests and responses of the HTTP traffic into a knative broker.
2 The broker sends these to a knative service for logging, called seldon-request-logger
3 seldon-request-logger processes and sends to elasticsearch

The seldon-request-logger enriches the raw message to optimise for searching.

Run `kubectl apply -f seldon-request-logger.yaml`


Create broker:

```
kubectl label namespace seldon-logs knative-eventing-injection=enabled
sleep 3
kubectl -n seldon-logs get broker default
```

The broker should show 'READY' as True.

Note that when we installed the seldon model earlier we told it to log to a broker in the seldon-logs namespace.

And trigger:

```
kubectl apply -f ./trigger.yaml
```

## Generating Logging

And the loadtester (first line is only needed for KIND):

```
kubectl label nodes kind-worker role=locust --overwrite
kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite

helm install seldon-core-loadtesting ../../helm-charts/seldon-core-loadtesting/ --set locust.host=http://seldon-single-model-default:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=1000 --set locust.maxWait=1000 --set replicaCount=1
```

## Inspecting Logging and Search for Requests

Access kibana with a port-forward to `localhost:5601`:
```
kubectl port-forward svc/kibana-kibana -n seldon-logs 5601:5601
```

When Kibana appears for the first time there will be a brief animation while it initializes.
On the Welcome page click Explore on my own.
From the top-left or from the `Visualize and Explore Data` panel select the `Discover` item.
In the form field Index pattern enter *
It should read "Success!" and Click the `> Next` step button on the right.
In the next form select timestamp from the dropdown labeled `Time Filter` field name.
From the bottom-right of the form select `Create index pattern`.
In a moment a list of fields will appear.
From the top-left or the home screen's `Visualize and Explore Data` panel, select the `Discover` item.
The log list will appear.
Refine the list a bit by selecting `log` near the bottom the left-hand Selected fields list.
When you hover over or click on the word `log`, click the `Add` button to the right of the label.
You can create a filter using the `Add Filter` button under `Search`. The field can be `kubernetes.labels.seldon-app` and the value can be an 'is' match on `seldon-single-model-default`.

To add mappings, go to `Management` at the bottom-left and then `Index Patterns`. Hit `Refresh` on the index created earlier. The number of fields should increase.

Now we can go back and add further filters if we want.

Adding a filter for `Ce-Inferenceservicename` exists will restrict to just request-response pairs.

![picture](./kibana-custom-search.png)


## Credits

Loosely based on https://www.katacoda.com/javajon/courses/kubernetes-observability/efk
Fluentd filtering based on https://blog.ptrk.io/tweaking-an-efk-stack-on-kubernetes/
