CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio
SCHEDULER_IMG ?= ${DOCKERHUB_USERNAME}/seldon-scheduler:${CUSTOM_IMAGE_TAG}
AGENT_IMG ?= ${DOCKERHUB_USERNAME}/seldon-agent:${CUSTOM_IMAGE_TAG}
RCLONE_IMG ?= ${DOCKERHUB_USERNAME}/seldon-rclone:${CUSTOM_IMAGE_TAG}
MODELGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-modelgateway:${CUSTOM_IMAGE_TAG}
ENVOY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-envoy:${CUSTOM_IMAGE_TAG}
MLSERVER_IMG ?= seldonio/mlserver:1.0.0.rc1
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:21.12-py3
KIND_NAME=ansible

GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

#####################################
# Build
#####################################

build-scheduler: test
	go build -o bin/scheduler ./cmd/scheduler

build-proxy:
	go build -o bin/proxy ./cmd/proxy

build-agent: test
	go build -o bin/agent ./cmd/agent

build-modelgateway: test
	go build -o bin/modelgateway -v ./cmd/modelgateway

build-dataflow-producer:
	go build -o data-flow/scripts/bin/producer ./data-flow/scripts/producer.go ./data-flow/scripts/common.go

build-dataflow-consumer:
	go build -o data-flow/scripts/bin/consumer ./data-flow/scripts/consumer.go ./data-flow/scripts/common.go

build-dataflow-engine:
	cd data-flow; ./gradlew clean build -x test --no-daemon

build: build-scheduler build-agent build-proxy build-modelgateway build-dataflow-engine

lint: ## Run go linters against code.
	gofmt -w pkg
	golangci-lint run --fix

test-go:
	go test ./pkg/... -coverprofile cover.out

test-jvm:
	cd data-flow; ./gradlew clean test --no-daemon

test: test-go test-jvm


#####################################
# Build Docker
#####################################

docker-build-scheduler: copy-apis
	docker build -t ${SCHEDULER_IMG} -f Dockerfile.scheduler .

docker-push-scheduler: 
	docker push ${SCHEDULER_IMG}

copy-apis:
	rm -rf apis-TEMP
	cp -r ../apis apis-TEMP

docker-build-agent: copy-apis
	docker build -t ${AGENT_IMG} -f Dockerfile.agent .

docker-push-agent: 
	docker push ${AGENT_IMG}

docker-build-rclone:
	docker build -t ${RCLONE_IMG} -f Dockerfile.rclone .

docker-push-rclone: 
	docker push ${RCLONE_IMG}

docker-build-modelgateway: copy-apis
	docker build -t ${MODELGATEWAY_IMG} -f Dockerfile.modelgateway .

docker-push-modelgateway: 
	docker push ${MODELGATEWAY_IMG}

docker-build-envoy:
	docker build -t ${ENVOY_IMG} -f Dockerfile.envoy .

docker-push-envoy: 
	docker push ${ENVOY_IMG}

docker-build-all: docker-build-agent docker-build-envoy docker-build-rclone docker-build-scheduler docker-build-modelgateway

docker-push-all: docker-push-agent docker-push-envoy docker-push-rclone docker-push-scheduler docker-push-modelgateway

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

kind-image-install-modelgateway: docker-build-modelgateway
	kind load -v 3 docker-image ${MODELGATEWAY_IMG} --name ${KIND_NAME}

kind-image-install-all: kind-image-install-scheduler kind-image-install-envoy kind-image-install-agent kind-image-install-rclone kind-image-install-modelgateway

#####################################
# Start with Docker Compose
#####################################

DOCKER_COMPOSE_COMMON_IMAGES = \
	SCHEDULER_IMAGE_AND_TAG=${SCHEDULER_IMG} \
		AGENT_IMAGE_AND_TAG=${AGENT_IMG} \
		ENVOY_IMAGE_AND_TAG=${ENVOY_IMG} \
		RCLONE_IMAGE_AND_TAG=${RCLONE_IMG}

DOCKER_COMPOSE_MLSERVER_IMAGES = \
	${DOCKER_COMPOSE_COMMON_IMAGES} \
		SERVER_IMAGE_AND_TAG=${MLSERVER_IMG}

DOCKER_COMPOSE_TRITON_IMAGES = \
	${DOCKER_COMPOSE_COMMON_IMAGES} \
		SERVER_IMAGE_AND_TAG=${TRITON_IMG}

DOCKER_COMPOSE_TRITON_LOG_LEVEL ?= 0

DOCKER_COMPOSE_USE_EMPTY_VOLUMES ?= true

ifeq ($(DOCKER_COMPOSE_USE_EMPTY_VOLUMES),true)
	DOCKER_COMPOSE_REMOVE_VOLUMES = -v
else
	DOCKER_COMPOSE_REMOVE_VOLUMES =
endif

DOCKER_COMPOSE_BASE_COMMAND = docker-compose --env-file env.all -f all-base.yaml 
DOCKER_COMPOSE_INTERNAL= -f all-internal.yaml
DOCKER_COMPOSE_HOST= -f all-host-network.yaml
DOCKER_COMPOSE_SERVICE_COMMAND = ${DOCKER_COMPOSE_COMMON_IMAGES} ${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -p scv2
DOCKER_COMPOSE_MLSERVER_COMMAND = ${DOCKER_COMPOSE_MLSERVER_IMAGES} ${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -f all-mlserver.yaml -p scv2_mlserver
DOCKER_COMPOSE_TRITON_COMMAND = ${DOCKER_COMPOSE_TRITON_IMAGES} ${DOCKER_COMPOSE_TRITON_LOG_LEVEL} ${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -f all-triton.yaml -p scv2_triton
DOCKER_COMPOSE_SERVICE_HOST_COMMAND = ${DOCKER_COMPOSE_COMMON_IMAGES} ${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_HOST} -p scv2

.PHONY: start-all-mlserver
start-all-mlserver: copy-apis
	${DOCKER_COMPOSE_MLSERVER_COMMAND} up -d

.PHONY: stop-all-mlserver
stop-all-mlserver:
	${DOCKER_COMPOSE_MLSERVER_COMMAND} down ${DOCKER_COMPOSE_REMOVE_VOLUMES}

.PHONY: start-all-triton
start-all-triton: copy-apis
	${DOCKER_COMPOSE_TRITON_COMMAND} up -d

.PHONY: stop-all-triton
stop-all-triton:
	${DOCKER_COMPOSE_TRITON_COMMAND} down ${DOCKER_COMPOSE_REMOVE_VOLUMES}

.PHONY: start-all
start-all: start-all-mlserver

.PHONY: stop-all
stop-all: stop-all-mlserver

# Tooling

.PHONY: list-services-mlserver
list-services-mlserver:
	${DOCKER_COMPOSE_MLSERVER_COMMAND} ps

.PHONY: list-services-triton
list-services-triton:
	${DOCKER_COMPOSE_TRITON_COMMAND} ps

.PHONY: list-services
list-services:
	${DOCKER_COMPOSE_SERVICE_COMMAND} ps

# Single services

.PHONY: start-mlserver
start-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_INTERNAL} \
		-f all-mlserver.yaml -p scv2 up -d server

.PHONY: stop-mlserver
stop-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_INTERNAL} \
		-f all-mlserver.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} server

.PHONY: start-mlserver-host
start-mlserver-host:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_HOST} \
		-f all-mlserver.yaml -p scv2 up -d server

.PHONY: stop-mlserver-host
stop-mlserver-host:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_HOST} \
		-f all-mlserver.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} server

.PHONY: start-triton
start-triton:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		TRITON_LOG_LEVEL=${DOCKER_COMPOSE_TRITON_LOG_LEVEL} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_INTERNAL} \
		-f all-triton.yaml -p scv2 up -d server

.PHONY: stop-triton
stop-triton:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_INTERNAL} \
		-f all-triton.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} server

.PHONY: start-triton-host
start-triton-host:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		TRITON_LOG_LEVEL=${DOCKER_COMPOSE_TRITON_LOG_LEVEL} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_HOST} \
		-f all-triton.yaml -p scv2 up -d server

.PHONY: stop-triton-host
stop-triton-host:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} \
		${DOCKER_COMPOSE_HOST} \
		-f all-triton.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} server

.PHONY: start-agent-mlserver
start-agent-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -f all-mlserver.yaml -p scv2 up -d agent

.PHONY: stop-agent-mlserver
stop-agent-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -f all-mlserver.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent

.PHONY: start-agent-mlserver-host
start-agent-mlserver-host:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_HOST} -f all-mlserver.yaml -p scv2 up -d agent

.PHONY: stop-agent-mlserver-host
stop-agent-mlserver-host:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_HOST} -f all-mlserver.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent

.PHONY: start-agent-triton
start-agent-triton:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -f all-triton.yaml -p scv2 up -d agent

.PHONY: stop-agent-triton
stop-agent-triton:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} -f all-triton.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent

.PHONY: start-agent-triton-host
start-agent-triton-host:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_HOST} -f all-triton.yaml -p scv2 up -d agent

.PHONY: stop-agent-triton-host
stop-agent-triton-host:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_HOST} -f all-triton.yaml -p scv2 rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent

.PHONY: start-scheduler
start-scheduler:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d scheduler

.PHONY: stop-scheduler
stop-scheduler:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force scheduler

.PHONY: start-scheduler-host
start-scheduler-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d scheduler

.PHONY: stop-scheduler-host
stop-scheduler-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force scheduler

.PHONY: start-envoy
start-envoy:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d envoy

.PHONY: stop-envoy
stop-envoy:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force envoy

.PHONY: start-envoy-host
start-envoy-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d envoy

.PHONY: stop-envoy-host
stop-envoy-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force envoy

.PHONY: start-rclone
start-rclone:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d rclone

.PHONY: stop-rclone
stop-rclone:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} rclone

.PHONY: start-rclone-host
start-rclone-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d rclone

.PHONY: stop-rclone-host
stop-rclone-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} rclone

.PHONY: start-modelgateway
start-modelgateway:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d modelgateway

.PHONY: stop-modelgateway
stop-modelgateway:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} modelgateway

.PHONY: start-modelgateway-host
start-modelgateway-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d modelgateway

.PHONY: stop-modelgateway-host
stop-modelgateway-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} modelgateway

.PHONY: start-kafka
start-kafka:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d kafka

.PHONY: stop-kafka
stop-kafka:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} kafka

.PHONY: start-kafka-host
start-kafka-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up kafka

.PHONY: stop-kafka-host
stop-kafka-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} kafka


.PHONY: stop-all
stop-all: stop-scheduler stop-envoy stop-rclone stop-agent-mlserver stop-agent-triton stop-mlserver stop-triton stop-modelgateway stop-kafka


#####################################
# Start local binaries
#####################################

start-scheduler-local:
	./bin/scheduler

start-agent-local:
	./bin/agent --agent-folder ${PWD}/mnt --inference-http-port 8080 --inference-grpc-port 8081 --scheduler-host "0.0.0.0" --scheduler-port 9005 \
		--server-type mlserver \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000000,"capabilities":["sklearn"],"overCommitPercentage":20}'


start-triton-agent-local:
	./bin/agent --agent-folder ${PWD}/mnt --inference-http-port 8080 --inference-grpc-port 8081 --scheduler-host "0.0.0.0" --scheduler-port 9005 \
		--server-name triton \
		--server-type triton \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000000,"capabilities":["tensorflow","onnx","pytorch"],"overCommitPercentage":20}'

start-envoy-local:
	./hack/start-envoy-delta.sh


start-rclone-local:
	mkdir -p mnt/rclone
	rm -rf mnt/rclone/*
	rm -f config/rclone.config
	rclone rcd --config=config/rclone.config --rc-no-auth --verbose

start-proxy-local:
ifdef LOG_LEVEL
	./bin/proxy --level $(LOG_LEVEL)
else
	./bin/proxy
endif

start-modelgateway-local:
	./bin/modelgateway --config-path ${PWD}/config --log-level debug

#####################################
# K8S Testing
#####################################

deploy:
	cd k8s/scheduler && kustomize edit set image scheduler=${SCHEDULER_IMG}
	cd k8s/envoy && kustomize edit set image envoy=${ENVOY_IMG}
	cd k8s/mlserver && kustomize edit set image agent=${AGENT_IMG}
	cd k8s/mlserver && kustomize edit set image rclone=${RCLONE_IMG}
	cd k8s/modelgateway && kustomize edit set image modelgateway=${MODELGATEWAY_IMG}
	kustomize build k8s/default | kubectl apply -f -

build-push-deploy: build docker-build-all docker-push-all deploy

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
