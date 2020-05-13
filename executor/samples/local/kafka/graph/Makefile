BASE=../../../..
KAFKA_BIN=../kafka_2.12-2.5.0/bin

run_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --http_port 8000 --kafka_broker 172.17.0.2:31415 -transport kafka --kafka_topic_in testin --kafka_topic_out testout 


run_proxy:
	${BASE}/kafka-proxy --model_name classifier --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --broker 172.17.0.2:31415

run_dummy_rest_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_rest_local 

produce_request:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	echo '{"data":{"ndarray":[[1.0,2.0]]}}' | ${KAFKA_BIN}/kafka-console-producer.sh --broker-list 172.17.0.2:31415 --topic testin

consume_responses:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	${KAFKA_BIN}/kafka-console-consumer.sh --bootstrap-server 172.17.0.2:31415 --topic testout 
