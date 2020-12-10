BASE=../../../..
KAFKA_BIN=../kafka_2.12-2.5.0/bin

run_proxy:
	${BASE}/kafka-proxy --model_name test --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --broker 172.17.0.2:31415

run_dummy_rest_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_rest_local 

produce_request:
	cd .. && $(MAKE) kafka_2.12-2.5.0
	echo '{"data":{"ndarray":[[1.0,2.0]]}}' | ${KAFKA_BIN}/kafka-console-producer.sh --broker-list 172.17.0.2:31415 --topic test.example.seldon-model.default 

