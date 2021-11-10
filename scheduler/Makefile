SCHEDULER_IMG ?= seldonio/seldon-scheduler:latest
AGENT_IMG ?= seldonio/seldon-agent:latest
RCLONE_IMG ?= seldonio/seldon-rclone:latest
ENVOY_IMG ?= seldonio/seldon-envoy:latest
KIND_NAME=ansible

GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

build-scheduler: test
	go build -o bin/scheduler -v ./cmd/scheduler

build-agent: test
	go build -o bin/agent -v ./cmd/agent

build: build-scheduler build-agent

test:
	go test ./pkg/...

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

start-scheduler:
	./bin/scheduler

start-envoy:
	./hack/start-envoy-delta.sh

start-agent:
	./bin/agent --model-repository ${PWD}/mlrepo --inference-port 8080 \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000,"availableMemoryBytes":1000,"capabilities":["sklearn"],"overCommit":false}'

start-mlserver:
	cd mlrepo && mlserver start .

start-rclone:
	cd mlrepo && rclone rcd --rc-no-auth --verbose

# sends host header to allow routing
curl-model2:
	watch -n1 curl -H \"Host: model2\" 0.0.0.0:9000/model2

docker-build-scheduler:
	docker build -t ${SCHEDULER_IMG} -f Dockerfile.scheduler .

docker-push-scheduler: ## Push docker image with the manager.
	docker push ${SCHEDULER_IMG}

kind-image-install-scheduler: docker-build-scheduler
	kind load -v 3 docker-image ${SCHEDULER_IMG} --name ${KIND_NAME}

docker-build-agent:
	docker build -t ${AGENT_IMG} -f Dockerfile.agent .

docker-push-agent: ## Push docker image with the manager.
	docker push ${AGENT_IMG}

kind-image-install-agent: docker-build-agent
	kind load -v 3 docker-image ${AGENT_IMG} --name ${KIND_NAME}

docker-build-rclone:
	docker build -t ${RCLONE_IMG} -f Dockerfile.rclone .

docker-push-rclone: ## Push docker image with the manager.
	docker push ${RCLONE_IMG}

kind-image-install-rclone: docker-build-rclone
	kind load -v 3 docker-image ${RCLONE_IMG} --name ${KIND_NAME}

docker-build-envoy:
	docker build -t ${ENVOY_IMG} -f Dockerfile.envoy .

docker-push-envoy: ## Push docker image with the manager.
	docker push ${ENVOY_IMG}

kind-image-install-envoy: docker-build-envoy 
	kind load -v 3 docker-image ${ENVOY_IMG} --name ${KIND_NAME}

kind-image-install-all: kind-image-install-scheduler kind-image-install-envoy kind-image-install-agent kind-image-install-rclone

deploy:
	cd k8s/scheduler && kustomize edit set image scheduler=${SCHEDULER_IMG}
	cd k8s/envoy && kustomize edit set image scheduler=${ENVOY_IMG}
	cd k8s/mlserver && kustomize edit set image agent=${AGENT_IMG}
	cd k8s/mlserver && kustomize edit set image rclone=${RCLONE_IMG}
	kustomize build k8s/default | kubectl apply -f -

undeploy:
	kustomize build k8s/default | kubectl delete -f -

deploy-mlserver:
	kubectl apply namespace seldon-mesh || echo "namespace seldon-mesh exists"
	kustomize build k8s/testing/mlserver | kubectl create -f -

undeploy-mlserver:
	kustomize build k8s/testing/mlserver | kubectl delete -f -


# ip address will change. Make automatic?
curl-mlserver-test:
	curl -v -H "Host: model2" -H "Content-Type: application/json" 172.18.255.1/v2/models/classifier/infer -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'

build-protos:
	protoc -I=./apis/mlops/scheduler --go_out=./apis/mlops/scheduler --go_opt=paths=source_relative --go-grpc_out=./apis/mlops/scheduler --go-grpc_opt=paths=source_relative ./apis/mlops/scheduler/scheduler.proto
	protoc -I=./apis/mlops/agent -I=./apis/mlops/scheduler --go_out=./apis/mlops/agent --go_opt=paths=source_relative --go-grpc_out=./apis/mlops/agent --go-grpc_opt=paths=source_relative ./apis/mlops/agent/agent.proto
