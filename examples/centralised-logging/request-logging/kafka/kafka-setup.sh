kubectl create ns kafka
curl -L https://github.com/strimzi/strimzi-kafka-operator/releases/download/0.11.4/strimzi-cluster-operator-0.11.4.yaml \
  | sed 's/namespace: .*/namespace: kafka/' \
  | kubectl -n kafka apply -f -

kubectl apply -f https://raw.githubusercontent.com/strimzi/strimzi-kafka-operator/0.11.4/examples/kafka/kafka-persistent-single.yaml -n kafka

kubectl apply -f https://github.com/knative/eventing/releases/download/v0.6.0/kafka.yaml

kubectl apply -f ./kafka-config.yaml

kubectl apply -f ./kafka-channel.yaml