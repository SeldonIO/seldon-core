CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio

IMG ?= ${DOCKERHUB_USERNAME}/seldon-k6:${CUSTOM_IMAGE_TAG}
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

CMD = kubectl describe pod ${POD} -n seldon-mesh | grep IP: -m 1 | cut -d' ' -f2- | xargs

POD = seldon-scheduler
SCHEDULER_ENDPOINT := $(shell ${CMD})

POD = seldon-envoy
ENVOY_ENDPOINT := $(shell ${CMD})

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
	cd ../../ && docker build --build-arg GIT_COMMIT=${GIT_COMMIT} --build-arg GIT_BRANCH=${GIT_BRANCH}  -t ${IMG} -f tests/k6/Dockerfile.k6 .

docker-run:
	docker run ${IMG}

build-push: docker-build docker-push

deploy-envoy-test:
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/envoy | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" ENVOY_ENDPOINT="${ENVOY_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-envoy-test:
	kustomize build configs/k8s/overlays/envoy | kubectl delete -f -

deploy-rproxy-test:
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/rproxy | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_TRITON_ENDPOINT}" envsubst | kubectl apply -f -

deploy-rproxy-mlserver-test:
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/rproxy | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_MLSERVER_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-rproxy-test:
	kustomize build configs/k8s/overlays/rproxy | kubectl delete -f -

deploy-server-test: 
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/server | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_TRITON_ENDPOINT}" envsubst | kubectl apply -f -

deploy-server-mlserver-test: 
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/server | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" RPROXY_ENDPOINT="${RPROXY_MLSERVER_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-server-test:
	kustomize build configs/k8s/overlays/server | kubectl delete -f -

deploy-kmodel-test: 
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/kmodel | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" PIPELINE_ENDPOINT="${PIPELINE_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-kmodel-test:
	kustomize build configs/k8s/overlays/kmodel | kubectl delete -f -

deploy-kpipeline-test: 
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/kpipeline | SCHEDULER_ENDPOINT="${SCHEDULER_ENDPOINT}" PIPELINE_ENDPOINT="${PIPELINE_ENDPOINT}" envsubst | kubectl apply -f -

undeploy-kpipeline-test:
	kustomize build configs/k8s/overlays/kpipeline | kubectl delete -f -

undeploy-all-test:
	kubectl get jobs -n seldon-mesh --no-headers=true | cut -d' ' -f1 | xargs kubectl delete -n seldon-mesh job

create-secret:
	#gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME} --display-name="SCV2 k6 tests"
	#gsutil iam ch serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com:objectAdmin gs://${GCS_BUCKET_NAME}/
	gcloud iam service-accounts keys create --iam-account "${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" k6-service-account.json
	kubectl create secret generic k6-sa-key --from-file k6-service-account.json -n seldon-mesh
