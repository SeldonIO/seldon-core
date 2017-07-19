API_ENDPOINT:=$(shell cat API_ENDPOINT)
API_CLIENT_KEY=$(shell cat api-client-key.txt)
API_CLIENT_SECRET=$(shell cat api-client-secret.txt)

EP_PING=/ping
EP_VERSION=/version
EP_TOKEN=/oauth/token
EP_FEEDBACK=/api/v0.1/feedback
EP_PREDICTION=/api/v0.1/predictions

ACCESS_TOKEN=$(shell curl -s -H "Accept: application/json" ${API_CLIENT_KEY}:$(API_CLIENT_SECRET)@$(API_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials|jq -r '.access_token')

notarget:
	@echo need target

update_api_endpoint_for_localhost:
	@echo "localhost:8080" > API_ENDPOINT && cat API_ENDPOINT
update_api_endpoint_for_minikube:
	@MINIKUBE_IP=$$(minikube ip) && echo "$${MINIKUBE_IP}:30032" > API_ENDPOINT && cat API_ENDPOINT


get_token:
	curl -v -H "Accept: application/json" ${API_CLIENT_KEY}:$(API_CLIENT_SECRET)@$(API_ENDPOINT)$(EP_TOKEN) -d grant_type=client_credentials

do_get_feedback:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -s -i -H "Accept: application/json" -H "Authorization: Bearer $(ACCESS_TOKEN)" $(API_ENDPOINT)$(EP_FEEDBACK)


do_post_test:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -v -i -X POST -H "Content-Type:application/json" -H "Accept: application/json" -H "Authorization: Bearer $(ACCESS_TOKEN)" $(API_ENDPOINT)$(EP_PREDICTION) -d '{"request":{"features":["a"],"values":[[1.0]]}}'

do_exception_test:
	echo "------------------------------" && echo "ACCESS_TOKEN[$(ACCESS_TOKEN)]" && echo "------------------------------" \
		&& curl -v -i -X POST -H "Content-Type:application/json" -H "Accept: application/json" -H "Authorization: Bearer $(ACCESS_TOKEN)" $(API_ENDPOINT)$(EP_PREDICTION) -d '{"meta":{"default_format":true},"data":{"feature":["a","b"],"values":[[1.0,2.0],[3.0,4.0]]}}'



compile_proto:
	python -m grpc.tools.protoc -I ../../proto --python_out=python/ --grpc_python_out=python/ ../../proto/prediction.proto
	python -m grpc.tools.protoc -I ../../proto --python_out=python/ --grpc_python_out=python/ ../../proto/seldonengine.proto

test_engine_endpoint:
	curl -X POST -H "Content-Type:application/json" -H "Accept: application/json" 0.0.0.0:8081/api/v0.1/predictions -d '{"request":{"features":["a"],"values":[[1.0]]}}'



