SHELL=/bin/bash
KAFKA_BIN=../kafka_2.13-2.8.0/bin
KAFKA_IP := $(shell kubectl get service  my-cluster-kafka-plain-0 -n kafka -o=jsonpath='{.status.loadBalancer.ingress[0].ip}{"\n"}')


#https://strimzi.io/quickstarts/
install:
	kubectl create namespace kafka
	kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka

create_cluster:
	kubectl apply -f https://strimzi.io/examples/latest/kafka/kafka-persistent-single.yaml -n kafka
	kubectl wait kafka/my-cluster --for=condition=Ready --timeout=300s -n kafka

create_topic:
	kubectl create -f topic.yaml

start_consumer:
	${KAFKA_BIN}/kafka-console-consumer.sh --bootstrap-server ${KAFKA_IP}:9092 --topic my-topic --from-beginning

start_producer:
	${KAFKA_BIN}/kafka-console-producer.sh --broker-list ${KAFKA_IP}:9092 --topic my-topic
