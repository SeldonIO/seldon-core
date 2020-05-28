BASE=../../..

## REST

run_rest_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_rest.yaml --port 8000 


run_dummy_rest_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_rest_local 

run_dummy_logsink:
	docker run -it -p 2222:80 --rm -t mendhak/http-https-echo

curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"ndarray":[[1.0,2.0]]}}'

curl_rest_big:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"names":["Age","Workclass","Education","Marital Status","Occupation","Relationship","Race","Sex","Capital Gain","Capital Loss","Hours per week","Country"],"ndarray":[[53,4,0,2,8,4,2,0,0,0,60,9]]}}'



## GRPC

run_grpc_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_grpc.yaml --port 8000

run_dummy_grpc_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_grpc_local 

grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:8000 seldon.protos.Seldon/Predict


