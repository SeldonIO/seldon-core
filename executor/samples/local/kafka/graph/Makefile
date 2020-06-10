BASE=../../../..
KAFKA_BIN=../kafka_2.12-2.5.0/bin
KAFKA_PORT := $(shell kubectl get service my-cluster-kafka-external-bootstrap -n default -o=jsonpath='{.spec.ports[0].nodePort}{"\n"}')
VERSION := $(shell cat ../../../../../version.txt)
EXECUTOR_IMAGE=seldonio/seldon-core-executor-ubi8

docker_run_executor_rest:
	docker run -it --rm --net=host -v ${PWD}:/models ${EXECUTOR_IMAGE}:${VERSION}  --sdep seldon-model --namespace default --predictor example --file /models/model_single.yaml --port 8000 --server_type kafka --kafka_broker 172.17.0.2:${KAFKA_PORT} -transport rest --kafka_input_topic testin --kafka_output_topic testout 


run_executor_rest:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --port 8000 --server_type kafka --kafka_broker 172.17.0.2:${KAFKA_PORT} -transport rest --kafka_input_topic testin --kafka_output_topic testout 

run_executor_rest_graph:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --port 8000 --server_type kafka --kafka_broker 172.17.0.2:${KAFKA_PORT} -transport rest --kafka_input_topic testin --kafka_output_topic testout --kafka_full_graph


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
