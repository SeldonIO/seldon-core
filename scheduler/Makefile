CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio
SCHEDULER_IMG ?= ${DOCKERHUB_USERNAME}/seldon-scheduler:${CUSTOM_IMAGE_TAG}
AGENT_IMG ?= ${DOCKERHUB_USERNAME}/seldon-agent:${CUSTOM_IMAGE_TAG}
RCLONE_IMG ?= ${DOCKERHUB_USERNAME}/seldon-rclone:${CUSTOM_IMAGE_TAG}
MODELGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-modelgateway:${CUSTOM_IMAGE_TAG}
PIPELINEGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-pipelinegateway:${CUSTOM_IMAGE_TAG}
DATAFLOW_IMG ?= ${DOCKERHUB_USERNAME}/seldon-dataflow-engine:${CUSTOM_IMAGE_TAG}
ENVOY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-envoy:${CUSTOM_IMAGE_TAG}
MLSERVER_IMG ?= seldonio/mlserver:1.1.0.dev3
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:21.12-py3
KIND_NAME=ansible

GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

#####################################
# Build
#####################################

.PHONY: build-scheduler
build-scheduler: test-go
	go build -o bin/scheduler ./cmd/scheduler

.PHONY: build-proxy
build-proxy: test-go
	go build -o bin/proxy ./cmd/proxy

.PHONY: build-agent
build-agent: test-go
	go build -o bin/agent ./cmd/agent

.PHONY: build-modelgateway
build-modelgateway: test-go
	go build -o bin/modelgateway -v ./cmd/modelgateway

.PHONY: build-pipelinegateway
build-pipelinegateway: test-go
	go build -o bin/pipelinegateway -v ./cmd/pipelinegateway

.PHONY: build-dataflow-producer
build-dataflow-producer: test-jvm
	go build -o data-flow/scripts/bin/producer ./data-flow/scripts/producer.go ./data-flow/scripts/common.go

.PHONY: build-dataflow-consumer
build-dataflow-consumer:
	go build -o data-flow/scripts/bin/consumer ./data-flow/scripts/consumer.go ./data-flow/scripts/common.go

.PHONY: build-dataflow-engine
build-dataflow-engine:
	cd data-flow; ./gradlew clean build -x test --no-daemon

.PHONY: build-go
build-go: build-scheduler build-agent build-proxy build-modelgateway build-pipelinegateway

.PHONY: build-jvm
build-jvm: build-dataflow-engine

.PHONY: build
build: build-go build-jvm

.PHONY: lint
lint: ## Run go linters against code.
	gofmt -w pkg
	gofmt -w cmd
	golangci-lint run --fix

.PHONY: test-go
test-go:
	go test ./pkg/... -coverprofile cover.out

.PHONY: test-jvm
test-jvm:
	cd data-flow; ./gradlew clean test --no-daemon

.PHONY: test
test: test-go test-jvm


#####################################
# Build Docker
#####################################

.PHONY: docker-build-scheduler
docker-build-scheduler: copy-apis
	docker build -t ${SCHEDULER_IMG} -f Dockerfile.scheduler .

.PHONY: docker-push-scheduler
docker-push-scheduler: 
	docker push ${SCHEDULER_IMG}

.PHONY: copy-apis
copy-apis:
	rm -rf apis-TEMP
	cp -r ../apis apis-TEMP

.PHONY: docker-build-agent
docker-build-agent: copy-apis
	docker build -t ${AGENT_IMG} -f Dockerfile.agent .

.PHONY: docker-push-agent
docker-push-agent: 
	docker push ${AGENT_IMG}

.PHONY: docker-build-rclone
docker-build-rclone:
	docker build -t ${RCLONE_IMG} -f Dockerfile.rclone .

.PHONY: docker-push-rclone
docker-push-rclone: 
	docker push ${RCLONE_IMG}

.PHONY: docker-build-modelgateway
docker-build-modelgateway: copy-apis
	docker build -t ${MODELGATEWAY_IMG} -f Dockerfile.modelgateway .

.PHONY: docker-push-modelgateway
docker-push-modelgateway: 
	docker push ${MODELGATEWAY_IMG}

docker-build-pipelinegateway: copy-apis
	docker build -t ${PIPELINEGATEWAY_IMG} -f Dockerfile.pipelinegateway .

docker-push-pipelinegateway: 
	docker push ${PIPELINEGATEWAY_IMG}

.PHONY: docker-build-envoy
docker-build-envoy:
	docker build -t ${ENVOY_IMG} -f Dockerfile.envoy .

.PHONY: docker-push-envoy
docker-push-envoy: 
	docker push ${ENVOY_IMG}

.PHONY: docker-build-dataflow
docker-build-dataflow: copy-apis
	docker build -t ${DATAFLOW_IMG} -f Dockerfile.dataflow .

.PHONY: docker-push-dataflow
docker-push-dataflow:
	docker push ${DATAFLOW_IMG}

.PHONY: docker-build-all
docker-build-all: docker-build-agent docker-build-envoy docker-build-rclone docker-build-scheduler docker-build-modelgateway docker-build-pipelinegateway docker-build-dataflow

.PHONY: docker-push-all
docker-push-all: docker-push-agent docker-push-envoy docker-push-rclone docker-push-scheduler docker-push-modelgateway docker-push-pipelinegateway docker-push-dataflow


#####################################
# Kind 
#####################################

.PHONY: kind-image-install-scheduler
kind-image-install-scheduler: docker-build-scheduler
	kind load -v 3 docker-image ${SCHEDULER_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-agent
kind-image-install-agent: docker-build-agent
	kind load -v 3 docker-image ${AGENT_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-envoy
kind-image-install-envoy: docker-build-envoy 
	kind load -v 3 docker-image ${ENVOY_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-rclone
kind-image-install-rclone: docker-build-rclone
	kind load -v 3 docker-image ${RCLONE_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-modelgateway
kind-image-install-modelgateway: docker-build-modelgateway
	kind load -v 3 docker-image ${MODELGATEWAY_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-pipelinegateway
kind-image-install-pipelinegateway: docker-build-pipelinegateway
	kind load -v 3 docker-image ${PIPELINEGATEWAY_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-dataflow
kind-image-install-dataflow: docker-build-dataflow
	kind load -v 3 docker-image ${DATAFLOW_IMG} --name ${KIND_NAME}

.PHONY: kind-image-install-all
kind-image-install-all: kind-image-install-scheduler kind-image-install-envoy kind-image-install-agent kind-image-install-rclone kind-image-install-modelgateway kind-image-install-pipelinegateway kind-image-install-dataflow

#####################################
# Start with Docker Compose
#####################################

.PHONY: create-model-folders
create-model-folders:
	mkdir -p ./mnt/mlserver/models
	mkdir -p ./mnt/mlserver/rclone
	mkdir -p ./mnt/triton/models
	mkdir -p ./mnt/triton/rclone

DOCKER_COMPOSE_COMMON_IMAGES = \
	SCHEDULER_IMAGE_AND_TAG=${SCHEDULER_IMG} \
		AGENT_IMAGE_AND_TAG=${AGENT_IMG} \
		ENVOY_IMAGE_AND_TAG=${ENVOY_IMG} \
		RCLONE_IMAGE_AND_TAG=${RCLONE_IMG} \
		SERVER_MLSERVER_IMAGE_AND_TAG=${MLSERVER_IMG} \
		TRITON_LOG_LEVEL=${DOCKER_COMPOSE_TRITON_LOG_LEVEL} \
		SERVER_TRITON_IMAGE_AND_TAG=${TRITON_IMG}

DOCKER_COMPOSE_TRITON_LOG_LEVEL ?= 0

DOCKER_COMPOSE_USE_EMPTY_VOLUMES ?= true

ifeq ($(DOCKER_COMPOSE_USE_EMPTY_VOLUMES),true)
	DOCKER_COMPOSE_REMOVE_VOLUMES = -v
else
	DOCKER_COMPOSE_REMOVE_VOLUMES =
endif

ifneq ($(LOCAL_MODEL_FOLDER),)
	DOCKER_COMPOSE_LOCAL_MODELS = -f model-volume.yaml
else
	DOCKER_COMPOSE_LOCAL_MODELS =
endif

DOCKER_COMPOSE_BASE_COMMAND = docker-compose --env-file env.all -f all-base.yaml 
DOCKER_COMPOSE_INTERNAL= -f all-internal.yaml
DOCKER_COMPOSE_HOST= -f all-host-network.yaml
DOCKER_COMPOSE_SERVICE_COMMAND = ${DOCKER_COMPOSE_COMMON_IMAGES} ${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_INTERNAL} ${DOCKER_COMPOSE_LOCAL_MODELS} -p scv2
DOCKER_COMPOSE_SERVICE_HOST_COMMAND = ${DOCKER_COMPOSE_COMMON_IMAGES} ${DOCKER_COMPOSE_BASE_COMMAND} ${DOCKER_COMPOSE_HOST} ${DOCKER_COMPOSE_LOCAL_MODELS} -p scv2

.PHONY: start-all
start-all: copy-apis
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d

.PHONY: stop-all
stop-all:
	${DOCKER_COMPOSE_SERVICE_COMMAND} down ${DOCKER_COMPOSE_REMOVE_VOLUMES}

.PHONY: start-all-host
start-all-host: copy-apis create-model-folders
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d

.PHONY: stop-all-host
stop-all-host: 
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} down ${DOCKER_COMPOSE_REMOVE_VOLUMES}


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
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d mlserver

.PHONY: stop-mlserver
stop-mlserver:
	${DOCKER_COMPOSE_SERVICE_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} mlserver

.PHONY: start-mlserver-host 
start-mlserver-host: create-model-folders
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d mlserver

.PHONY: stop-mlserver-host
stop-mlserver-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} mlserver

.PHONY: start-triton
start-triton:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d triton

.PHONY: stop-triton
stop-triton:
	${DOCKER_COMPOSE_SERVICE_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} triton

.PHONY: start-triton-host
start-triton-host: create-model-folders
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d triton

.PHONY: stop-triton-host
stop-triton-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} triton

.PHONY: start-agent-mlserver
start-agent-mlserver:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d agent-mlserver

.PHONY: stop-agent-mlserver
stop-agent-mlserver:
	${DOCKER_COMPOSE_SERVICE_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent-mlserver

.PHONY: start-agent-mlserver-host
start-agent-mlserver-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d agent-mlserver

.PHONY: stop-agent-mlserver-host
stop-agent-mlserver-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent-mlserver

.PHONY: start-agent-triton
start-agent-triton:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d agent-triton

.PHONY: stop-agent-triton
stop-agent-triton:
	${DOCKER_COMPOSE_SERVICE_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent-triton

.PHONY: start-agent-triton-host
start-agent-triton-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d agent-triton

.PHONY: stop-agent-triton-host
stop-agent-triton-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} \
		rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} agent-triton

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
start-rclone-mlserver:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d rclone-mlserver

.PHONY: stop-rclone-mlserver
stop-rclone-mlserver:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} rclone-mlserver

.PHONY: start-rclone-mlserver-host
start-rclone-mlserver-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d rclone-mlserver

.PHONY: stop-rclone-mlserver-host
stop-rclone-mlserver-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} rclone-mlserver

.PHONY: start-rclone
start-rclone-triton:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d rclone-triton

.PHONY: stop-rclone-triton
stop-rclone-triton:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} rclone-triton

.PHONY: start-rclone-triton-host
start-rclone-triton-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d rclone-triton

.PHONY: stop-rclone-triton-host
stop-rclone-triton-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} rclone-triton

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

.PHONY: start-pipelinegateway
start-pipelinegateway:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d pipelinegateway

.PHONY: stop-pipelinegateway
stop-pipelinegateway:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} pipelinegateway

.PHONY: start-pipelinegateway-host
start-pipelinegateway-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d pipelinegateway

.PHONY: stop-pipelinegateway-host
stop-pipelinegateway-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} pipelinegateway

.PHONY: start-kafka
start-kafka:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d kafka

.PHONY: stop-kafka
stop-kafka:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} kafka

.PHONY: start-kafka-host
start-kafka-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d kafka

.PHONY: stop-kafka-host
stop-kafka-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES} kafka

.PHONY: start-dataflow-engine
start-dataflow-engine:
	${DOCKER_COMPOSE_SERVICE_COMMAND} up -d dataflow

.PHONY: stop-dataflow-engine
stop-dataflow-engine:
	${DOCKER_COMPOSE_SERVICE_COMMAND} rm --stop --force dataflow

.PHONY: start-dataflow-engine-host
start-dataflow-engine-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} up -d dataflow

.PHONY: stop-dataflow-engine-host
stop-dataflow-engine-host:
	${DOCKER_COMPOSE_SERVICE_HOST_COMMAND} rm --stop --force dataflow


#####################################
# Start local binaries
#####################################

.PHONY: start-scheduler-local
start-scheduler-local:
	./bin/scheduler

.PHONY: start-agent-local-mlserver
start-agent-local-mlserver:
	./bin/agent --agent-folder ${PWD}/mnt/mlserver --inference-http-port 8080 --inference-grpc-port 8081 --scheduler-host "0.0.0.0" --scheduler-port 9005 \
		--server-type mlserver \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000000,"capabilities":["alibi-detect","lightgbm","python","sklearn","xgboost"],"overCommitPercentage":20}'


.PHONY: start-agent-local-triton
start-agent-local-triton:
	./bin/agent --agent-folder ${PWD}/mnt/triton --inference-http-port 8082 --inference-grpc-port 8083 --scheduler-host "0.0.0.0" --scheduler-port 9005 --rclone-port 5573 \
		--server-name triton \
		--server-type triton \
		--log-level debug \
		--config-path ${PWD}/config \
		--replica-config '{"inferenceSvc":"0.0.0.0","inferenceHttpPort":8080,"inferenceGrpcPort":8081,"memoryBytes":1000000,"capabilities":["dali","fil","onnx","openvino","pytorch","tensorflow","tensorrt","triton-python"],"overCommitPercentage":20}'

.PHONY: start-envoy-local
start-envoy-local:
	./hack/start-envoy-delta.sh


.PHONY: start-rclone-local
start-rclone-local: create-model-folders
	rm -rf mnt/mlserver/rclone/*
	rm -rf mnt/triton/rclone/*
	rm -f config/rclone.config
	rclone rcd --config=config/rclone.config --rc-no-auth --verbose

.PHONY: start-proxy-local
start-proxy-local:
ifdef LOG_LEVEL
	./bin/proxy --level $(LOG_LEVEL)
else
	./bin/proxy
endif

.PHONY: start-modelgateway-local
start-modelgateway-local:
	./bin/modelgateway --config-path ${PWD}/config --log-level debug

.PHONY: start-pipelinegateway-local
start-pipelinegateway-local:
	./bin/pipelinegateway --config-path ${PWD}/config/kafka-host.json --log-level debug

data-flow/opentelemetry-javaagent.jar:
	cd data-flow && wget https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar

.PHONY: start-dataflow-engine-local
start-dataflow-engine-local: data-flow/opentelemetry-javaagent.jar
	SELDON_UPSTREAM_HOST=localhost \
        SELDON_UPSTREAM_PORT=9008 \
        SELDON_KAFKA_BOOTSTRAP_SERVERS=localhost:9092 \
        SELDON_CORES_COUNT=4 \
        java -javaagent:./data-flow/opentelemetry-javaagent.jar \
		-Dotel.resource.attributes=service.name=seldon-dataflow \
		-jar ./data-flow/build/libs/dataflow-1.0-SNAPSHOT.jar


#####################################
# K8S Testing
#####################################

.PHONY: deploy
deploy:
	cd k8s/scheduler && kustomize edit set image scheduler=${SCHEDULER_IMG}
	cd k8s/envoy && kustomize edit set image envoy=${ENVOY_IMG}
	cd k8s/mlserver && kustomize edit set image agent=${AGENT_IMG}
	cd k8s/mlserver && kustomize edit set image rclone=${RCLONE_IMG}
	cd k8s/modelgateway && kustomize edit set image modelgateway=${MODELGATEWAY_IMG}
	cd k8s/pipelinegateway && kustomize edit set image pipelinegateway=${PIPELINEGATEWAY_IMG}
	cd k8s/dataflow-engine && kustomize edit set image dataflow-engine=${DATAFLOW_IMG}
	kustomize build k8s/default | kubectl apply -f -

.PHONY: build-push-deploy
build-push-deploy: build docker-build-all docker-push-all deploy

.PHONY: deploy-minio-secret
deploy-minio-secret:
	kubectl create namespace seldon-mesh || echo "seldon-mesh namespace exists"
	kustomize build k8s/auth | kubectl apply -f -

.PHONY: undeploy
undeploy:
	kustomize build k8s/default | kubectl delete -f -

.PHONY: deploy-servers
deploy-servers:
	kustomize build k8s/server | kubectl create -f -

.PHONY: undeploy-servers
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
