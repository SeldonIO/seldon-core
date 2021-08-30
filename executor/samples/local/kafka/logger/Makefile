BASE=../../../..
KAFKA_BIN=../kafka_2.13-2.8.0/bin
# Assume kafka cluster created by ansible playbooks in core woth loadbalncer exposed endpoints
KAFKA_IP := $(shell kubectl get service  seldon-kafka-plain-0 -n kafka -o=jsonpath='{.status.loadBalancer.ingress[0].ip}{"\n"}')
KAFKA_TOPIC=seldon

## REST

run_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model.yaml --http_port 8000 --grpc_port 9500 --debug --log_kafka_broker ${KAFKA_IP}:9092 --log_kafka_topic ${KAFKA_TOPIC}


run_dummy_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_local 

curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"ndarray":[[1.0,2.0]]}}'

curl_rest_big:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"names":["Age","Workclass","Education","Marital Status","Occupation","Relationship","Race","Sex","Capital Gain","Capital Loss","Hours per week","Country"],"ndarray":[[53,4,0,2,8,4,2,0,0,0,60,9]]}}'

grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:9500 seldon.protos.Seldon/Predict

consume_responses:
	${KAFKA_BIN}/kafka-console-consumer.sh --bootstrap-server ${KAFKA_IP}:9092 --topic ${KAFKA_TOPIC}


