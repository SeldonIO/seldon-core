SCHEDULER_IMG ?= seldonio/seldon-scheduler:latest
AGENT_IMG ?= seldonio/seldon-agent:latest
RCLONE_IMG ?= seldonio/seldon-rclone:latest
ENVOY_IMG ?= seldonio/seldon-envoy:latest
ENVOY_LOCAL_IMG ?= seldonio/seldon-envoy-local:latest
MLSERVER_IMG ?= seldonio/mlserver:1.0.0.rc1
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:21.12-py3
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
# Start with Docker Compose
#####################################

DOCKER_COMPOSE_COMMON_IMAGES = \
	SCHEDULER_IMAGE_AND_TAG=${SCHEDULER_IMG} \
		AGENT_IMAGE_AND_TAG=${AGENT_IMG} \
		ENVOY_IMAGE_AND_TAG=${ENVOY_LOCAL_IMG} \
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

.PHONY: start-all-mlserver
start-all-mlserver: copy-apis
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		docker-compose -f all-base.yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver up -d

.PHONY: stop-all-mlserver
stop-all-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		docker-compose -f all-base.yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver down ${DOCKER_COMPOSE_REMOVE_VOLUMES}

.PHONY: start-all-triton
start-all-triton: copy-apis
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		TRITON_LOG_LEVEL=${DOCKER_COMPOSE_TRITON_LOG_LEVEL} \
		docker-compose -f all-base.yaml -f all-triton.yaml --env-file env.all -p scv2_triton up -d

.PHONY: stop-all-triton
stop-all-triton:
	${DOCKER_COMPOSE_TRITON_IMAGES} \
		docker-compose -f all-base.yaml -f all-triton.yaml --env-file env.all -p scv2_triton down ${DOCKER_COMPOSE_REMOVE_VOLUMES}

# Single services

.PHONY: start-mlserver
start-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		docker-compose -f all-base-yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver run server -d

.PHONY: stop-mlserver
stop-mlserver:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		docker-compose -f all-base.yaml -f all-mlserver.yaml --env-file env.all -p scv2_mlserver rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES}

.PHONY: start-triton
start-triton:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		docker-compose -f all-base-yaml -f all-triton.yaml --env-file env.all -p scv2_triton run server -d

.PHONY: stop-triton
stop-triton:
	${DOCKER_COMPOSE_MLSERVER_IMAGES} \
		docker-compose -f all-base.yaml -f all-triton.yaml --env-file env.all -p scv2_triton rm --stop --force ${DOCKER_COMPOSE_REMOVE_VOLUMES}


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
