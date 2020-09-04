SHELL=/bin/bash
KAFKA_BIN=../kafka_2.12-2.5.0/bin

KAFKA_PORT := $(shell kubectl get service my-cluster-kafka-external-bootstrap -n default -o=jsonpath='{.spec.ports[0].nodePort}{"\n"}')


# https://github.com/strimzi/strimzi-kafka-operator/tree/master/helm-charts/strimzi-kafka-operator
install:
	helm install my-release strimzi/strimzi-kafka-operator


create_cluster:
	kubectl create -f cluster.yaml

create_topic:
	kubectl create -f topic.yaml

start_consumer:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	${KAFKA_BIN}/kafka-console-consumer.sh --bootstrap-server 172.17.0.2:${KAFKA_PORT} --topic my-topic --from-beginning

start_producer:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	${KAFKA_BIN}/kafka-console-producer.sh --broker-list 172.17.0.2:${KAFKA_PORT} --topic my-topic
