# Centralised Logging Example

## Introduction

Here we will set up EFK (elasticsearch, fluentbit, kibana) as a stack to gather logs from SeldonDeployments and make them searchable.

This demo is aimed at minikube and based on https://www.katacoda.com/javajon/courses/kubernetes-observability/efk

Alternatives are available and if you are running in cloud then you can consider a managed service from your cloud provider.

## Setup

If helm is not already set up then it needs to be configured:

kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller

First we will install elasticsearch as the storage backend. This will require a [persistent volume](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolume) to be available from which elasticsearch can make a claim.

kubectl apply -f pv-elasticsearch-data.yaml
kubectl apply -f pv-elasticsearch-master.yaml

Note you would want to use a different type of volume in cloud - the above is for minikube.

Now install elasticsearch:

helm install stable/elasticsearch --name=elasticsearch --namespace=logs \
--set client.replicas=1 \
--set master.replicas=1 \
--set cluster.env.MINIMUM_MASTER_NODES=1 \
--set cluster.env.RECOVER_AFTER_MASTER_NODES=1 \
--set cluster.env.EXPECTED_MASTER_NODES=1 \
--set data.replicas=1 \
--set data.heapSize=300m \
--set master.persistence.storageClass=elasticsearch-master \
--set master.persistence.size=500M \
--set data.persistence.storageClass=elasticsearch-data \
--set data.persistence.size=500M

Next fluentbit as a collection agent:

helm install stable/fluent-bit --name=fluent-bit --namespace=logs --set backend.type=es --set backend.es.host=elasticsearch-client

And kibana UI:

helm install stable/kibana --name=kibana --namespace=logs --set env.ELASTICSEARCH_HOSTS=http://elasticsearch-client:9200 --set service.type=NodePort --set service.nodePort=31000

## Generating Logging

First we need seldon and a seldon deployment.

Install seldon operator:

helm install seldon-core-operator --name seldon-core --set usageMetrics.enabled=true --namespace seldon-system --repo https://storage.googleapis.com/seldon-charts

Now a model:

helm install seldon-single-model --name seldon-single-model --repo https://storage.googleapis.com/seldon-charts

And the loadtester:

kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust --overwrite
helm install seldon-core-loadtesting --name seldon-core-loadtesting --repo https://storage.googleapis.com/seldon-charts --set locust.host=http://seldon-single-model-seldon-single-model:8000 --set oauth.enabled=false --set oauth.key=oauth-key --set oauth.secret=oauth-secret --set locust.hatchRate=1 --set locust.clients=1 --set loadtest.sendFeedback=0 --set locust.minWait=0 --set locust.maxWait=0 --set replicaCount=1


## Inspecting Logging

echo $(minikube ip)":31000" to find kibana URL

When Kibana appears for the first time there will be a brief animation while it initializes.
On the Welcome page click Explore on my own.
From the left-hand menu select the top Discover item.
In the form field Index pattern enter kubernetes_cluster-*
It should read "Success!" and Click the > Next step button on the right.
In the next form select timestamp from the dropdown labeled Time Filter field name.
From the bottom-right of the form select Create index pattern.
In a moment a list of fields will appear.
Again, from the left-hand menu select the top Discover item.
The log list will appear.
Refine the list a bit by selecting log near the bottom the left-hand Selected fields list.
When you hover over or click on the word log, click the Add button to the right of the label.
You should be able to filter for e.g. kubernetes.labels.seldon-app or kubernetes.container_name