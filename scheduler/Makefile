SCHEDULER_IMG ?= seldonio/seldon-scheduler:latest
ENVOY_IMG ?= seldonio/seldon-envoy:latest
KIND_NAME=ansible

GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

#go build -o seldon-xds-server -mod=readonly -v -ldflags="$(GO_LDFLAGS)" $(MODULE)/cmd/server
build: ## Build and install the binary
	go build -o bin/scheduler -v ./cmd/server

start-local-svcs:
	docker run -d --rm --name=echo9100 -p 9100:8080 stevesloka/echo-server echo-server --echotext=Sample-Endpoint1!
	docker run -d --rm --name=echo9101 -p 9101:8080 stevesloka/echo-server echo-server --echotext=Sample-Endpoint2!
	docker run -d --rm --name=echo9102 -p 9102:8080 stevesloka/echo-server echo-server --echotext=Sample-Endpoint3!
	docker run -d --rm --name=echo9103 -p 9103:8080 stevesloka/echo-server echo-server --echotext=Sample-Endpoint4!
	docker run -d --rm --name=echo9104 -p 9104:8080 stevesloka/echo-server echo-server --echotext=Sample-Endpoint5!


stop-svcs:
	docker rm -f echo9100 
	docker rm -f echo9101
	docker rm -f echo9102 
	docker rm -f echo9103 
	docker rm -f echo9104

start-scheduler-from-file:
	./bin/scheduler -config ./config/seldon.yaml

start-scheduler-from-configmap:
	./bin/scheduler

start-envoy_imcremental:
	./hack/start-envoy-delta.sh

curl-model2:
	watch -n1 curl 0.0.0.0:9000/model2


docker-build-scheduler:
	docker build -t ${SCHEDULER_IMG} -f Dockerfile.scheduler .

docker-push-scheduler: ## Push docker image with the manager.
	docker push ${SCHEDULER_IMG}

kind-image-install-scheduler: docker-build-scheduler
	kind load -v 3 docker-image ${SCHEDULER_IMG} --name ${KIND_NAME}

docker-build-envoy:
	docker build -t ${ENVOY_IMG} -f Dockerfile.envoy .

docker-push-envoy: ## Push docker image with the manager.
	docker push ${ENVOY_IMG}

kind-image-install-envoy: docker-build-envoy 
	kind load -v 3 docker-image ${ENVOY_IMG} --name ${KIND_NAME}

deploy:
	cd k8s/scheduler && kustomize edit set image scheduler=${SCHEDULER_IMG}
	cd k8s/envoy && kustomize edit set image scheduler=${ENVOY_IMG}
	kustomize build k8s/default | kubectl create -f -

undeploy:
	kustomize build k8s/default | kubectl delete -f -

deploy-mlserver:
	kustomize build k8s/testing/mlserver | kubectl create -f -

undeploy-mlserver:
	kustomize build k8s/testing/mlserver | kubectl delete -f -


# ip address will change. Make automatic?
curl_mlserver_test:
	curl -v -H "Host: model2" -H "Content-Type: application/json" 172.18.255.6/v2/models/classifier/infer -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'

build_protos:
	protoc -I=./apis/mlops/scheduler --go_out=./apis/mlops/scheduler --go_opt=paths=source_relative --go-grpc_out=./apis/mlops/scheduler --go-grpc_opt=paths=source_relative ./apis/mlops/scheduler/scheduler.proto
	protoc -I=./apis/mlops/agent --go_out=./apis/mlops/agent --go_opt=paths=source_relative --go-grpc_out=./apis/mlops/agent --go-grpc_opt=paths=source_relative ./apis/mlops/agent/agent.proto
