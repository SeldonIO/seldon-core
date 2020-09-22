BASE=../../..

## REST

run_rest_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_rest.yaml --http_port 8000 


run_model:
	export PREDICTIVE_UNIT_SERVICE_PORT=9001 && export PREDICTIVE_UNIT_METRICS_SERVICE_PORT=6001 && seldon-core-microservice --service-type MODEL Model REST

run_transformer:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && seldon-core-microservice --service-type TRANSFORMER Transformer REST


curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"jsonData":"{\"data\":{\"ndarray\":[[1.0,2.0]]}}"}'

curl_rest_multipart:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -F jsonData=@input.json




