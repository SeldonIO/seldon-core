BASE=../../../..
KAFKA_BIN=../kafka_2.12-2.5.0/bin
KAFKA_PORT := $(shell kubectl get service my-cluster-kafka-external-bootstrap -n default -o=jsonpath='{.spec.ports[0].nodePort}{"\n"}')


run_executor_rest:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --http_port 8000 --server_type kafka --kafka_broker 172.17.0.2:${KAFKA_PORT} -transport rest --kafka_topic_in testin --kafka_topic_out testout 

run_executor_rest_graph:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --http_port 8000 --server_type kafka --kafka_broker 172.17.0.2:${KAFKA_PORT} -transport rest --kafka_topic_in testin --kafka_topic_out testout --kafka_graph_internal


run_proxy:
	${BASE}/kafka-proxy --model_name classifier --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --broker 172.17.0.2:${KAFKA_PORT}

run_dummy_rest_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_rest_local 

produce_request:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	echo '{"data":{"ndarray":[[1.0,2.0]]}}' | ${KAFKA_BIN}/kafka-console-producer.sh --broker-list 172.17.0.2:${KAFKA_PORT} --topic testin

consume_responses:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	${KAFKA_BIN}/kafka-console-consumer.sh --bootstrap-server 172.17.0.2:${KAFKA_PORT} --topic testout 
