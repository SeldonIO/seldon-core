SCHEDULER_IMG ?= seldonio/seldon-scheduler:latest
AGENT_IMG ?= seldonio/seldon-agent:latest
RCLONE_IMG ?= seldonio/seldon-rclone:latest
ENVOY_IMG ?= seldonio/seldon-envoy:latest
KIND_NAME=ansible

GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

build-scheduler: test
	go build -o bin/scheduler -v ./cmd/scheduler

build-proxy:
	go build -o bin/proxy -v ./cmd/proxy

build-agent: test
	go build -o bin/agent -v ./cmd/agent

build: build-scheduler build-agent build-proxy

lint: ## Run go linters against code.
	gofmt -w pkg
	golangci-lint run --fix

test:
	go test ./pkg/... -coverprofile cover.out

start-scheduler:
	./bin/scheduler

start-proxy:
ifdef LOG_LEVEL
	./bin/proxy --level $(LOG_LEVEL)
else
	./bin/proxy
endif

start-envoy:
	./hack/start-envoy-delta.sh

start-agent:
	./bin/agent --agent-folder ${PWD}/mnt --inference-port 8080 \
		--server-type mlserver \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000,"capabilities":["sklearn"],"overCommit":false}'

start-triton-agent:
	./bin/agent --agent-folder ${PWD}/mnt --inference-port 8080 \
		--server-name triton \
		--server-type triton \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000,"capabilities":["tensorflow","onnx","pytorch"],"overCommit":false}'

start-mlserver:
	mkdir -p mnt/models
	rm -rf mnt/models/*
	mlserver start ${PWD}/mnt/models

start-triton:
	mkdir -p mnt/models
	rm -rf mnt/models/*
	docker run --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p8080:8080 -p8081:8081 -v ${PWD}/mnt/models:/mnt/models nvcr.io/nvidia/tritonserver:21.12-py3 /opt/tritonserver/bin/tritonserver --model-repository=/mnt/models --http-port=8080 --grpc-port=8081 --log-verbose=1 --model-control-mode=explicit

start-rclone:
	mkdir -p mnt/rclone
	rm -rf mnt/rclone/*
	rm -f config/rclone.config
	rclone rcd --config=config/rclone.config --rc-no-auth --verbose

# sends host header to allow routing
curl-model2:
	watch -n1 curl -H \"Host: model2\" 0.0.0.0:9000/model2

copy-apis:
	rm -rf apis-TEMP
	cp -r ../apis apis-TEMP

docker-build-scheduler: copy-apis
	docker build -t ${SCHEDULER_IMG} -f Dockerfile.scheduler .

docker-push-scheduler: ## Push docker image with the manager.
	docker push ${SCHEDULER_IMG}

kind-image-install-scheduler: docker-build-scheduler
	kind load -v 3 docker-image ${SCHEDULER_IMG} --name ${KIND_NAME}

docker-build-agent: copy-apis
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

deploy-minio-secret:
	kubectl create namespace seldon-mesh || echo "seldon-mesh namespace exists"
	kustomize build k8s/auth | kubectl apply -f -

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

.PHONY: build-triton-protos
build-triton-protos:
	protoc \
		-I. \
		--go_opt=paths=source_relative \
		--go_out=. \
		./pkg/agent/repository/triton/config/model_config.proto
