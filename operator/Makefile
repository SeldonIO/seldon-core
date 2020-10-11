SHELL := /bin/bash
VERSION := $(shell cat ../version.txt)
# Image URL to use all building/pushing image targets
IMAGE_NAME_BASE=seldon-core-operator
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
#CRD_OPTIONS ?= ""

KIND_NAME ?= kind

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

.PHONY: lint
lint: licenses/dep.txt
	# Check if licenses have changed
	git \
		--no-pager diff \
		--exit-code \
		./licenses

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

# Install V1 CRDs into a cluster
# Note use of create to stop too long annotation being created. See https://github.com/kubernetes-sigs/kubebuilder/issues/1140
install_v1: manifests_v1
	kustomize build config/crd_v1 | kubectl create -f -

# Uninstall V1 CRDs into a cluster
uninstall_v1: 
	kustomize build config/crd_v1 | kubectl delete -f -

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

# Commented out alternative is looking ahead to issue that on Openshift our v1 CRD is too large
# to be installed. This may also affect operator-sdk community operators.
# See https://github.com/operator-framework/operator-registry/issues/385
# Solution may be to drop v1alpha2 and v1alpha3 versions to decrease size by 2/3
manifests_v1: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd_v1/bases crd:crdVersions=v1
	#$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./apis/machinelearning.seldon.io/v1" output:crd:artifacts:config=config/crd_v1/bases crd:crdVersions=v1

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

# password can be found at: https://connect.redhat.com/project/1366481/view
redhat-image-scan:
	docker pull ${IMG_REDHAT}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_operator} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-7f50cebe-122b-495a-a143-41426dfcb6c9/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-7f50cebe-122b-495a-a143-41426dfcb6c9/${IMG_VERSION_REDHAT}

kind-image-install: docker-build
	docker save ${IMG} > operator.tar
	kind load image-archive operator.tar --name ${KIND_NAME}

kind-image-install-redhat: docker-build-redhat
	docker save ${IMG_REDHAT} > operator.tar
	kind load image-archive operator.tar --name ${KIND_NAME}

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

install-dev:
	# Tool to generate license info
	pip install \
		'git+https://github.com/kubeflow/testing#egg=go-license-tools&subdirectory=py/kubeflow/testing/go-license-tools'

.PHONY: licenses/dep.txt
licenses/dep.txt:
	go list -m all | cut -d ' ' -f 1 > licenses/dep.txt

.PHONY: licenses
licenses: licenses/dep.txt
	# NOTE: You need to create a file in ~/.github_api_token with a GitHub token.
	get-github-repo \
		-o licenses/repo.txt \
		--manual-dep-repo-mapping licenses/dep_repo.manual.csv \
		licenses/dep.txt
	get-github-license-info -o licenses/license_info.csv licenses/repo.txt
	python -m 'patch_additional_license_info' \
		licenses/license_info.csv \
		licenses/additional_license_info.csv
	concatenate-license -o licenses/license.txt licenses/license_info.csv

generate-resources:
	rm -rf generated
	mkdir generated
	kustomize build config/default/ -o generated
	cp generated/apiextensions.k8s.io_v1beta1_customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml testing/machinelearning.seldon.io_seldondeployments.yaml

config/crd_v1/patches/graph_children.yaml:
	python hack/create_graph_openapi_schema.py hack/graph_patch.tmpl.yaml config/crd_v1/patches/graph_children.yaml


###################################
#
# Openshift / OperatorHub addtions
#
###################################

#
# Testing with operstor-sdk 1.0.1
#

.PHONY: bundle
recreate_bundle:
	rm -r bundle
	kustomize build config/manifests | operator-sdk --verbose generate bundle --version ${VERSION}
	python hack/csv_hack.py bundle/manifests/seldon-operator.clusterserviceversion.yaml

.PHONY: update_packagemanifests
update_packagemanifests:
	kustomize build config/manifests | operator-sdk --verbose generate packagemanifests --version ${VERSION}
	python hack/csv_hack.py packagemanifests/${VERSION}/seldon-operator.clusterserviceversion.yaml
	mv packagemanifests/${VERSION}/seldon-operator.clusterserviceversion.yaml packagemanifests/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml

#
# Tested with operator-courier 2.1.10
#
.PHONY: verify_package
verify_package:
	operator-courier verify packagemanifests/

# Commented out as bundle/opm_index seems the way to test on latest Openshift
#.PHONY: push_package
#push_package:
#	operator-courier push packagemanifests/ seldon seldon-operator ${VERSION} "$QUAY_TOKEN"

.PHONY: create_bundle_image
create_bundle_image:
	docker build . -f bundle.Dockerfile -t quay.io/seldon/seldon-operator:v${VERSION}

.PHONY: push_bundle_image
push_bundle_image:
	docker push quay.io/seldon/seldon-operator:v${VERSION}

.PHONY: validate_bundle_image
validate_bundle_image:
	operator-sdk bundle validate quay.io/seldon/seldon-operator:v${VERSION}

# PREV_VERSION needs manual update for now
PREV_VERSION=1.2.2
opm_index:
	opm index add -c docker --bundles quay.io/seldon/seldon-operator:v${VERSION},quay.io/seldon/seldon-operator:v${PREV_VERSION} --tag quay.io/seldon/test-catalog:latest


opm_push:
	docker push quay.io/seldon/test-catalog:latest

#
# Scorecard
#

scorecard:
	operator-sdk scorecard --kubeconfig ~/.kube/config quay.io/seldon/seldon-operator:v${VERSION}


#
# Certified Openshift Operator
#
# https://redhat-connect.gitbook.io/certified-operator-guide/appendix/bundle-maintenance-after-migration

PKG_CERT=packagemanifests-certified
create_certified_bundle: 
	rm -rf ${PKG_CERT}/${VERSION}
	mkdir -p ${PKG_CERT}/${VERSION}
	cp bundle/manifests/* ${PKG_CERT}/${VERSION}
	./hack/update-openshift-certified.sh ${VERSION}
	cd ${PKG_CERT} && opm alpha bundle generate -c alpha -p seldon-operator -d ./${VERSION}/ -u ./${VERSION}/
	mv ${PKG_CERT}/bundle.Dockerfile ${PKG_CERT}/bundle-${VERSION}.Dockerfile
	mv ${PKG_CERT}/${VERSION}/seldon-operator.clusterserviceversion.yaml ${PKG_CERT}/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
	mv ${PKG_CERT}/${VERSION}/manifests/seldon-operator.clusterserviceversion.yaml ${PKG_CERT}/${VERSION}/manifests/seldon-operator.v${VERSION}.clusterserviceversion.yaml
	cd ${PKG_CERT} && sed -i 's#seldon-operator.v.*#seldon-operator.v${VERSION}#' seldon-operator.package.yaml

build_certified_bundle:
	opm alpha bundle build -d packagemanifests-certified/${VERSION}/ --tag quay.io/seldon/seldon-operator-certified:v${VERSION} --package seldon-operator --channels alpha --overwrite -b docker


bundle_certified_push:
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_operator_bundle} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag quay.io/seldon/seldon-operator-certified:v${VERSION} scan.connect.redhat.com/ospid-78efc0cc-a4a5-4244-a31d-6820eae1ccb1/seldon-operator-certified:${VERSION}
	docker push scan.connect.redhat.com/ospid-78efc0cc-a4a5-4244-a31d-6820eae1ccb1/seldon-operator-certified:${VERSION}
