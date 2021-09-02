BASE=../../..

## REST

run_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model.yaml --http_port 8000 --grpc_port 9500 --debug


run_dummy_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_local 

run_dummy_logsink:
	docker run -it -p 2222:80 --rm -t mendhak/http-https-echo

curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"ndarray":[[1.0,2.0]]}}'

curl_rest_big:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"names":["Age","Workclass","Education","Marital Status","Occupation","Relationship","Race","Sex","Capital Gain","Capital Loss","Hours per week","Country"],"ndarray":[[53,4,0,2,8,4,2,0,0,0,60,9]]}}'

grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:9500 seldon.protos.Seldon/Predict


