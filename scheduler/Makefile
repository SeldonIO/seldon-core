SCHEDULER_IMG ?= seldonio/seldon-scheduler:latest
AGENT_IMG ?= seldonio/seldon-agent:latest
RCLONE_IMG ?= seldonio/seldon-rclone:latest
ENVOY_IMG ?= seldonio/seldon-envoy:latest
ENVOY_LOCAL_IMG ?= seldonio/seldon-envoy-local:latest
KIND_NAME=ansible

GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

#####################################
# Build
#####################################

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


#####################################
# Build Docker
#####################################

docker-build-scheduler: copy-apis
	docker build -t ${SCHEDULER_IMG} -f Dockerfile.scheduler .

docker-push-scheduler: ## Push docker image with the manager.
	docker push ${SCHEDULER_IMG}

copy-apis:
	rm -rf apis-TEMP
	cp -r ../apis apis-TEMP

docker-build-agent: copy-apis
	docker build -t ${AGENT_IMG} -f Dockerfile.agent .

docker-push-agent: ## Push docker image with the manager.
	docker push ${AGENT_IMG}

docker-build-rclone:
	docker build -t ${RCLONE_IMG} -f Dockerfile.rclone .

docker-push-rclone: ## Push docker image with the manager.
	docker push ${RCLONE_IMG}

docker-build-envoy:
	docker build -t ${ENVOY_IMG} -f Dockerfile.envoy .

docker-build-envoy-local:
	docker build -t ${ENVOY_LOCAL_IMG} -f Dockerfile.envoy-local .

docker-push-envoy: ## Push docker image with the manager.
	docker push ${ENVOY_IMG}

docker-build-all: docker-build-agent docker-build-envoy docker-build-envoy-local docker-build-rclone docker-build-scheduler

#####################################
# Kind 
#####################################

kind-image-install-scheduler: docker-build-scheduler
	kind load -v 3 docker-image ${SCHEDULER_IMG} --name ${KIND_NAME}

kind-image-install-agent: docker-build-agent
	kind load -v 3 docker-image ${AGENT_IMG} --name ${KIND_NAME}

kind-image-install-envoy: docker-build-envoy 
	kind load -v 3 docker-image ${ENVOY_IMG} --name ${KIND_NAME}

kind-image-install-rclone: docker-build-rclone
	kind load -v 3 docker-image ${RCLONE_IMG} --name ${KIND_NAME}

kind-image-install-all: kind-image-install-scheduler kind-image-install-envoy kind-image-install-agent kind-image-install-rclone

#####################################
# Start with Docker
#####################################

start-scheduler:
	docker run --name scheduler -d --rm --network host -p 9004:9004 -p 9002:9002 -p 9005:9005 ${SCHEDULER_IMG}
  
start-agent-mlserver:
	docker run --name agent -d --rm --network host -v ${PWD}/config:/mnt/config -v ${PWD}/mnt:/mnt/agent -e SELDON_SERVER_CAPABILITIES='sklearn,xgboost' -e SELDON_OVERCOMMIT='false' -e SELDON_SERVER_HTTP_PORT='8080' -e SELDON_SERVER_GRPC_PORT='8081' -e SELDON_SCHEDULER_HOST='0.0.0.0' -e SELDON_SCHEDULER_PORT='9005' -e SELDON_SERVER_TYPE='mlserver' -e MEMORY_REQUEST='1000000' ${AGENT_IMG} /bin/agent --log-level debug --config-path /mnt/config

start-agent-triton:
	docker run --name agent -d --rm --network host -v ${PWD}/config:/mnt/config -v ${PWD}/mnt:/mnt/agent -e SELDON_SERVER_CAPABILITIES='tensorflow,onnx,pytorch' -e SELDON_OVERCOMMIT='false' -e SELDON_SERVER_HTTP_PORT='8080' -e SELDON_SERVER_GRPC_PORT='8081' -e SELDON_SCHEDULER_HOST='0.0.0.0' -e SELDON_SCHEDULER_PORT='9005' -e SELDON_SERVER_TYPE='triton' -e MEMORY_REQUEST='1000000' ${AGENT_IMG} /bin/agent --log-level debug --config-path /mnt/config

start-envoy:
	docker run --name envoy -d --rm --network host -p 9003:9003 -p 9000:9000 ${ENVOY_LOCAL_IMG}

start-rclone:
	docker run --name rclone -d --rm -v ${PWD}/mnt:/mnt/agent -p 5572:5572 seldonio/seldon-rclone:latest

start-mlserver:
	mkdir -p mnt/models
	rm -rf mnt/models/*
	docker run --name mlserver -d --rm -v ${PWD}/mnt/models:/mnt/models -p 8080:8080 -p 8081:8081 seldonio/mlserver:1.0.0.rc1 mlserver start /mnt/models

start-triton-debug:
	mkdir -p mnt/models
	rm -rf mnt/models/*
	docker run --name triton -d --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p8080:8080 -p8081:8081 -v ${PWD}/mnt/models:/mnt/models nvcr.io/nvidia/tritonserver:21.12-py3 /opt/tritonserver/bin/tritonserver --model-repository=/mnt/models --http-port=8080 --grpc-port=8081 --log-verbose=1 --model-control-mode=explicit

start-triton:
	mkdir -p mnt/models
	rm -rf mnt/models/*
	docker run --name triton -d --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p8080:8080 -p8081:8081 -v ${PWD}/mnt/models:/mnt/models nvcr.io/nvidia/tritonserver:21.12-py3 /opt/tritonserver/bin/tritonserver --model-repository=/mnt/models --http-port=8080 --grpc-port=8081 --model-control-mode=explicit


start-all-mlserver: start-scheduler start-agent-mlserver start-mlserver start-envoy start-rclone
stop-all-mlserver:
	docker rm -f scheduler
	docker rm -f agent
	docker rm -f mlserver
	docker rm -f envoy
	docker rm -f rclone
start-all-triton: start-scheduler start-agent-triton start-triton start-envoy start-rclone
stop-all-triton:
	docker rm -f scheduler
	docker rm -f agent
	docker rm -f triton
	docker rm -f envoy
	docker rm -f rclone


#####################################
# Start local binaries
#####################################

start-scheduler-local:
	./bin/scheduler

start-agent-local:
	./bin/agent --agent-folder ${PWD}/mnt --inference-http-port 8080 --inference-grpc-port 8081 \
		--server-type mlserver \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000000,"capabilities":["sklearn"],"overCommit":false}'

start-triton-agent-local:
	./bin/agent --agent-folder ${PWD}/mnt --inference-http-port 8080 --inference-grpc-port 8081 \
		--server-name triton \
		--server-type triton \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000000,"capabilities":["tensorflow","onnx","pytorch"],"overCommit":false}'

start-envoy-local:
	./hack/start-envoy-delta.sh


start-rclone-local:
	mkdir -p mnt/rclone
	rm -rf mnt/rclone/*
	rm -f config/rclone.config
	rclone rcd --config=config/rclone.config --rc-no-auth --verbose

start-proxy:
ifdef LOG_LEVEL
	./bin/proxy --level $(LOG_LEVEL)
else
	./bin/proxy
endif

#####################################
# K8S Testing
#####################################

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

deploy-servers:
	kustomize build k8s/server | kubectl create -f -

undeploy-servers:
	kustomize build k8s/server | kubectl delete -f -

#####################################
# Misc
#####################################

.PHONY: build-triton-protos
build-triton-protos:
	protoc \
		-I. \
		--go_opt=paths=source_relative \
		--go_out=. \
		./pkg/agent/repository/triton/config/model_config.proto
