BASE=../../../..
KAFKA_BIN=../kafka_2.13-2.8.0/bin
# Assume kafka cluster created by ansible playbooks in core woth loadbalncer exposed endpoints
KAFKA_IP := $(shell kubectl get service  seldon-kafka-plain-0 -n kafka -o=jsonpath='{.status.loadBalancer.ingress[0].ip}{"\n"}')
VERSION := $(shell cat ../../../../../version.txt)
EXECUTOR_IMAGE=seldonio/seldon-core-executor-ubi8

docker_run_executor_rest:
	docker run -it --rm --net=host -v ${PWD}:/models ${EXECUTOR_IMAGE}:${VERSION}  --sdep seldon-model --namespace default --predictor example --file /models/model_single.yaml --server_type kafka --kafka_broker ${KAFKA_IP}:9092 -transport rest --kafka_input_topic testin --kafka_output_topic testout 


run_executor_rest:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --server_type kafka --kafka_broker ${KAFKA_IP}:9092 --kafka_input_topic testin --kafka_output_topic testout 

run_executor_rest_graph:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --port 8000 --server_type kafka --kafka_broker ${KAFKA_IP}:9092 --kafka_input_topic testin --kafka_output_topic testout --kafka_full_graph


run_proxy:
	${BASE}/kafka-proxy --model_name classifier --sdep seldon-model --namespace default --predictor example --file ./model_single.yaml --broker ${KAFKA_IP}:9092

run_dummy_rest_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_local 

produce_request:
	echo '{"data":{"ndarray":[[1.0,2.0]]}}' | ${KAFKA_BIN}/kafka-console-producer.sh --broker-list ${KAFKA_IP}:9092 --topic testin

consume_responses:
	${KAFKA_BIN}/kafka-console-consumer.sh --bootstrap-server ${KAFKA_IP}:9092 --topic testout 
