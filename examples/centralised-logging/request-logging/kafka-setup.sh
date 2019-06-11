kubectl create ns kafka
curl -L https://github.com/strimzi/strimzi-kafka-operator/releases/download/0.11.4/strimzi-cluster-operator-0.11.4.yaml \
  | sed 's/namespace: .*/namespace: kafka/' \
  | kubectl -n kafka apply -f -

kubectl apply -f https://raw.githubusercontent.com/strimzi/strimzi-kafka-operator/0.11.4/examples/kafka/kafka-persistent-single.yaml -n kafka

kubectl apply -f ./kafka-config.yaml

kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/200-channelable-manipulator-clusterrole.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/200-controller-clusterrole.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/200-dispatcher-clusterrole.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/200-dispather-service.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/200-serviceaccount.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/200-webhook-clusterrole.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/201-clusterrolebinding.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/300-kafka-channel.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/400-webhook-service.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/500-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/500-dispatcher.yaml
kubectl apply -f https://raw.githubusercontent.com/knative/eventing/master/contrib/kafka/config/500-webhook.yaml

kubectl apply -f ./kafka-channel.yaml