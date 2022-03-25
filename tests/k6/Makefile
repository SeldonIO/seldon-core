CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio

IMG ?= ${DOCKERHUB_USERNAME}/seldon-k6:${CUSTOM_IMAGE_TAG}
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

docker-push: 
	docker push ${IMG}

docker-build:
	cd ../../ && docker build --build-arg GIT_COMMIT=${GIT_COMMIT} --build-arg GIT_BRANCH=${GIT_BRANCH}  -t ${IMG} -f tests/k6/Dockerfile.k6 .

docker-run:
	docker run ${IMG}

build-push: docker-build docker-push

deploy-envoy-test:
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/envoy | kubectl apply -f -

undeploy-envoy-test:
	kustomize build configs/k8s/overlays/envoy | kubectl delete -f -

deploy-rproxy-test:
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/rproxy | kubectl apply -f -

undeploy-rproxy-test:
	kustomize build configs/k8s/overlays/rproxy | kubectl delete -f -

deploy-server-test: 
	cd configs/k8s/base && kustomize edit set image k6=${IMG} 
	kustomize build configs/k8s/overlays/server | kubectl apply -f -

undeploy-server-test:
	kustomize build configs/k8s/overlays/server | kubectl delete -f -