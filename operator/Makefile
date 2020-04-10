VERSION := $(shell cat ../version.txt)
# Image URL to use all building/pushing image targets
IMAGE_NAME_BASE=seldon-core-operator
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY:show_image
show_image:
	echo ${IMG}

all: manager

# Run tests
test: generate fmt vet manifests generate-resources
	ginkgo -r -outputdir=. -cover -coverprofile=cover.out ./controllers ./utils ./apis  

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go --webhook-port=9000

install-cert-manager:
	kubectl create namespace cert-manager || echo "Namespace cert-manager-exists"
	kubectl label namespace cert-manager cert-manager.io/disable-validation=true || echo "namespace cert-manager-already labelled"
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.12.0/cert-manager.yaml
	kubectl rollout status deployment.apps/cert-manager -n cert-manager
	kubectl rollout status deployment.apps/cert-manager-cainjector -n cert-manager
	kubectl rollout status deployment.apps/cert-manager-webhook -n cert-manager

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Install CRDs into a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

undeploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl delete -f -

undeploy-namespaced1: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/namespaced1 | kubectl delete -f -

undeploy-namespaced2: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/namespaced2 | kubectl delete -f -

undeploy-controllerid: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/controllerid | kubectl delete -f -

undeploy-lite: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/lite | kubectl delete -f -

deploy-local: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/local | kubectl apply -f -

deploy-namespaced1: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/namespaced1 | kubectl apply -f -

deploy-namespaced2: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/namespaced2 | kubectl apply -f -

deploy-controllerid: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/controllerid | kubectl apply -f -

deploy-cert: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/cert | kubectl apply -f -

deploy-lite: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/lite | kubectl apply -f -


# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Generate Clientset
create-client: test
	./hack/update-codegen.sh	

# Build the docker image
docker-build: test generate-resources
	docker build . -t ${IMG}

# Build the docker image for Redhat
docker-build-redhat: test generate-resources
	docker build . -f Dockerfile.redhat -t ${IMG_REDHAT}


# Push the docker image
docker-push:
	docker push ${IMG}

# Push the docker image
docker-push-redhat:
	docker push ${IMG_REDHAT}

# password can be found at: https://connect.redhat.com/project/1074021/view
redhat-image-scan:
	docker login -u unused scan.connect.redhat.com
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-82f819a9-f40d-4d59-b67c-136491f22cfb/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-82f819a9-f40d-4d59-b67c-136491f22cfb/${IMG_VERSION_REDHAT}

kind-image-install: docker-build
	docker save ${IMG} > operator.tar
	kind load image-archive operator.tar 

kind-image-install-redhat: docker-build-redhat
	docker save ${IMG_REDHAT} > operator.tar
	kind load image-archive operator.tar 

# find or download controller-gen
# download controller-gen if necessary

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}

CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

WEBHOOK_DIR=/tmp/k8s-webhook-server/serving-certs

tls-extract:
	mkdir -p ${WEBHOOK_DIR}
	kubectl get secrets -n seldon-system seldon-webhook-server-cert  -o 'go-template={{index .data "tls.key"}}' | base64 -d > ${WEBHOOK_DIR}/tls.key
	kubectl get secrets -n seldon-system seldon-webhook-server-cert  -o 'go-template={{index .data "tls.crt"}}' | base64 -d > ${WEBHOOK_DIR}/tls.crt


.PHONY: self-signed-cert
self-signed-cert:
	mkdir -p self-signed-cert
	./generate-keys.sh self-signed-cert


clean-cert:
	rm -r self-signed-cert


licenses/dep.txt:
	go list -m all | cut -d ' ' -f 1 > licenses/dep.txt

generate-resources:
	rm -rf generated
	mkdir generated
	kustomize build config/default/ -o generated
	cp generated/apiextensions.k8s.io_v1beta1_customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml testing/machinelearning.seldon.io_seldondeployments.yaml
