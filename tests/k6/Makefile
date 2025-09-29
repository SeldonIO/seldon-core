CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio
NAMESPACE ?= seldon-mesh
K6_VERSION ?= 1.1.0
XK6_VERSION ?= 1.0.1

IMG ?= ${DOCKERHUB_USERNAME}/seldon-k6:${CUSTOM_IMAGE_TAG}
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

SELDON_CLUSTER=seldon

CMD = kubectl describe pod ${POD} -n ${NAMESPACE} | grep IP: -m 1 | cut -d' ' -f2- | xargs
CMD_SVC = kubectl describe svc seldon-mesh -n ${NAMESPACE} | grep IP: -m 1 | cut -d' ' -f2- | xargs

POD = seldon-scheduler
SCHEDULER_ENDPOINT := $(shell ${CMD})

ENVOY_ENDPOINT := $(shell ${CMD_SVC})

POD = mlserver-0
RPROXY_MLSERVER_ENDPOINT := $(shell ${CMD})

POD = triton-0
RPROXY_TRITON_ENDPOINT := $(shell ${CMD})

POD = seldon-pipelinegateway
PIPELINE_ENDPOINT := $(shell ${CMD})

SERVICE_ACCOUNT_NAME ?= scv2-k6-tests
GCS_BUCKET_NAME ?= seldon-tmp
PROJECT_ID ?= seldon-pub

docker-push:
	docker push ${IMG}

docker-build:
	tar -czvf ../../k6.tar.gz .
	cd ../../ && docker build \
		--build-arg GIT_COMMIT=${GIT_COMMIT} \
		--build-arg GIT_BRANCH=${GIT_BRANCH} \
		--build-arg K6_SEMVER=${K6_VERSION} \
		--build-arg XK6_VERSION=${XK6_VERSION} \
		-t ${IMG} \
		-f tests/k6/Dockerfile.k6 .
	rm ../../k6.tar.gz

docker-build-and-push-prod:
	tar -czvf ../../k6.tar.gz .
	cd ../../ && docker buildx build \
		--provenance=true \
		--build-arg GIT_COMMIT=${GIT_COMMIT} \
		--build-arg GIT_BRANCH=${GIT_BRANCH} \
		--build-arg K6_SEMVER=${K6_VERSION} \
		--build-arg XK6_VERSION=${XK6_VERSION} \
		-t ${IMG} \
		--attest type=sbom,generator=docker/scout-sbom-indexer:latest \
		--push \
		-f tests/k6/Dockerfile.k6 .
	rm ../../k6.tar.gz

docker-run:
	docker run ${IMG}

build-push: docker-build docker-push

kind-push-image:
	kind load docker-image ${IMG} --name ${SELDON_CLUSTER}

.PHONY: set-img-and-namespace
set-img-and-namespace:
	cd configs/k8s/base && kustomize edit set image k6=${IMG} && kustomize edit set namespace ${NAMESPACE}

.PHONY: deploy-1-model-pipeline-test
deploy-1-model-pipeline-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/1_model_pipeline | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-1-model-pipeline-test:
	kustomize build configs/k8s/overlays/pipeline/1_model_pipeline | kubectl delete -f -

.PHONY: deploy-1-model-pipeline-all_svcs_scaled-test
deploy-1-model-pipeline-all_svcs_scaled-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/1_model_pipeline_all_svcs_scaled | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-1-model-pipeline-all_svcs_scaled-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/1_model_pipeline_all_svcs_scaled | kubectl delete -f -

.PHONY: deploy-5-model-double-pipeline-test
deploy-5-model-double-pipeline-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/5_model_double_pipeline | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-5-model-double-pipeline-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/5_model_double_pipeline | kubectl delete -f -

.PHONY: deploy-parallel-pipeline-test
deploy-parallel-pipeline-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/parallel_pipeline | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-parallel-pipeline-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/parallel_pipeline | kubectl delete -f -

.PHONY: deploy-scale-all-up-down-during-test
deploy-scale-all-up-down-during-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_all_up_down_during_test | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-scale-all-up-down-during-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_all_up_down_during_test | kubectl delete -f -

.PHONY: deploy-scale-dataflow-engine-during-test
deploy-scale-dataflow-engine-during-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_dataflow_engine_during_test | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-scale-dataflow-engine-during-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_dataflow_engine_during_test | kubectl delete -f -

.PHONY: deploy-scale-model-gw-engine-during-test
deploy-scale-model-gw-engine-during-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_model_gw_during_test | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-scale-model-gw-engine-during-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_model_gw_during_test | kubectl delete -f -

.PHONY: deploy-scale-pipeline-gw-engine-during-test
deploy-scale-pipeline-gw-engine-during-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_pipeline_gw_during_test | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-scale-pipeline-gw-engine-during-test:
	kustomize build configs/k8s/overlays/pipeline/scalability/scale_pipeline_gw_during_test | kubectl delete -f -

.PHONY: deploy-envoy-test
deploy-envoy-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/envoy | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-envoy-test:
	kustomize build configs/k8s/overlays/envoy | kubectl delete -f -

.PHONY: deploy-rproxy-test
deploy-rproxy-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/rproxy | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_TRITON_ENDPOINT}" envsubst | kubectl apply -f -

.PHONY: deploy-rproxy-mlserver-test
deploy-rproxy-mlserver-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/rproxy | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_MLSERVER_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-rproxy-test:
	kustomize build configs/k8s/overlays/rproxy | kubectl delete -f -

.PHONY: deploy-server-test
deploy-server-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/server | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_TRITON_ENDPOINT}" envsubst | kubectl apply -f -

.PHONY: deploy-server-mlserver-test
deploy-server-mlserver-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/server | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_MLSERVER_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-server-test:
	kustomize build configs/k8s/overlays/server | kubectl delete -f -

.PHONY: deploy-kmodel-test
deploy-kmodel-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/kmodel | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" PIPELINE_ENDPOINT="${PIPELINE_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-kmodel-test:
	kustomize build configs/k8s/overlays/kmodel | kubectl delete -f -

.PHONY: deploy-kpipeline-test
deploy-kpipeline-test: set-img-and-namespace
	kustomize build configs/k8s/overlays/kpipeline | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" PIPELINE_ENDPOINT="${PIPELINE_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-kpipeline-test:
	kustomize build configs/k8s/overlays/kpipeline | kubectl delete -f -

undeploy-all-test:
	kubectl get jobs -n ${NAMESPACE} --no-headers=true | cut -d' ' -f1 | xargs kubectl delete -n ${NAMESPACE} job

create-secret:
	#gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME} --display-name="SCV2 k6 tests"
	#gsutil iam ch serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com:objectAdmin gs://${GCS_BUCKET_NAME}/
	gcloud iam service-accounts keys create --iam-account "${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" k6-service-account.json
	kubectl create secret generic k6-sa-key --from-file k6-service-account.json -n ${NAMESPACE}

xk6-install:
	# Install xk6
	go install go.k6.io/xk6/cmd/xk6@v${XK6_VERSION}
	xk6 build v${K6_VERSION} --with github.com/grafana/xk6-kubernetes
