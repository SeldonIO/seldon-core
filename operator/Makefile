SHELL := /bin/bash
VERSION := $(shell cat ../version.txt)
# Image URL to use all building/pushing image targets
IMAGE_NAME_BASE=seldon-core-operator
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}

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
test: install-ginkgo generate fmt vet manifests_all generate-resources
	ACK_GINKGO_RC=true \
	ACK_GINKGO_DEPRECATIONS=1.16.4 \
	$(GINKGO) -r -outputdir=. -cover -coverprofile=cover.out ./controllers ./utils ./apis

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests_all
	go run ./main.go --webhook-port=9000

install-cert-manager:
	kubectl create namespace cert-manager || echo "Namespace cert-manager-exists"
	kubectl label namespace cert-manager cert-manager.io/disable-validation=true || echo "namespace cert-manager-already labelled"
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml
	kubectl rollout status deployment.apps/cert-manager -n cert-manager
	kubectl rollout status deployment.apps/cert-manager-cainjector -n cert-manager
	kubectl rollout status deployment.apps/cert-manager-webhook -n cert-manager


manifests_all: manifests manifests_v1_small

# Install CRDs into a cluster
# Note use of create to stop too long annotation being created. See https://github.com/kubernetes-sigs/kubebuilder/issues/1140
install: manifests
	kustomize build config/crd | kubectl create -f -

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

# Commented out alternative is looking ahead to issue that on Openshift our v1 CRD is too large
# to be installed. This may also affect operator-sdk community operators.
# See https://github.com/operator-framework/operator-registry/issues/385
# Solution may be to drop v1alpha2 and v1alpha3 versions to decrease size by 2/3
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases crd:crdVersions=v1
	#$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./apis/machinelearning.seldon.io/v1" output:crd:artifacts:config=config/crd/bases crd:crdVersions=v1

manifests_v1_small: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./apis/machinelearning.seldon.io/v1" output:crd:artifacts:config=config/crd_v1_small/bases crd:crdVersions=v1


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

# Build the docker image without testing
docker-build-no-test: generate-resources
	docker build . -t ${IMG}

# Build the docker image for Redhat
docker-build-redhat: test generate-resources
	docker build -f Dockerfile.redhat -t ${IMG_REDHAT} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Push the docker image
docker-push-redhat:
	docker push ${IMG_REDHAT}

kind-image-install: docker-build-no-test
	kind load -v 3 docker-image ${IMG} --name ${KIND_NAME}

kind-image-install-redhat: docker-build-redhat
	docker save ${IMG_REDHAT} > operator.tar
	kind load image-archive operator.tar --name ${KIND_NAME}

# find or download controller-gn
# download controller-gen if necessary

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0 ;\
	}

CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: install-ginkgo
install-ginkgo:
ifeq (, $(shell command -v ginkgo))
	go install github.com/onsi/ginkgo/ginkgo@v1.16.4

GINKGO=$(GOBIN)/ginkgo
else
GINKGO=$(shell type -p ginkgo)
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
	cp generated/apiextensions.k8s.io_v1_customresourcedefinition_seldondeployments.machinelearning.seldon.io.yaml testing/machinelearning.seldon.io_seldondeployments.yaml

config/crd/patches/graph_children.yaml:
	python hack/create_graph_openapi_schema.py hack/graph_patch.tmpl.yaml config/crd/patches/graph_children.yaml


###################################
#
# Openshift community
#
###################################


.PHONY: bundle
recreate_bundle:
	rm -fr bundle
	kustomize build config/manifests | operator-sdk --verbose generate bundle --default-channel stable --version ${VERSION} --channels stable
	python hack/csv_hack.py bundle/manifests/seldon-operator.clusterserviceversion.yaml ${VERSION}
	echo '  com.redhat.openshift.versions: "v4.8-v4.9"' >> bundle/metadata/annotations.yaml
	echo 'LABEL com.redhat.openshift.versions=v4.8-v4.9' >> bundle.Dockerfile

create_bundle_image:
	docker build . -f bundle.Dockerfile -t quay.io/seldon/seldon-operator:v${VERSION}

push_bundle_image:
	docker push quay.io/seldon/seldon-operator:v${VERSION}

.PHONY: validate_bundle_image
validate_bundle_image:
	operator-sdk bundle validate quay.io/seldon/seldon-operator:v${VERSION}


# Only index up to 1.2.2 as we have removed "replaces" in 1.2.2 to stop chain there for testing
.PHONY: opm_index
opm_index:
	opm index add -c docker --bundles quay.io/seldon/seldon-operator:v${VERSION} --mode replaces --tag quay.io/seldon/test-catalog:latest


opm_push:
	docker push quay.io/seldon/test-catalog:latest


.PHONY: update_openshift
update_openshift: recreate_bundle create_bundle_image push_bundle_image validate_bundle_image opm_index opm_push


#
# Scorecard
#

scorecard:
	operator-sdk scorecard --kubeconfig ~/.kube/config quay.io/seldon/seldon-operator:v${VERSION}


#
# Community and Upstream Operators
#

# Change to local checkout
# This should be paths to forks of
# - https://github.com/k8s-operatorhub/community-operators
# - https://github.com/redhat-openshift-ecosystem/community-operators-prod
COMMUNITY_OPERATORS_FOLDER=~/work/red-hat/community-operators
UPSTREAM_OPERATORS_FOLDER=~/work/red-hat/community-operators-prod

update_community:
	cp -r bundle/. ${COMMUNITY_OPERATORS_FOLDER}/operators/seldon-operator/${VERSION}

# Presently fails
test_community:
	cd ${COMMUNITY_OPERATORS_FOLDER} && export OP_TEST_DEBUG=3 &&  bash <(curl -sL https://raw.githubusercontent.com/redhat-openshift-ecosystem/community-operators-pipeline/ci/latest/ci/scripts/opp.sh) kiwi,lemon,orange operators/seldon-operator/${VERSION}

update_upstream:
	cp -r bundle/. ${UPSTREAM_OPERATORS_FOLDER}/operators/seldon-operator/${VERSION}


###################################
#
# Openshift Certified
#
###################################

create_certified_bundle:
	rm -rf bundle-certified/manifests
	cp -r bundle/manifests bundle-certified
	mv bundle-certified/manifests/seldon-operator.clusterserviceversion.yaml bundle-certified/manifests/seldon-operator-certified.clusterserviceversion.yaml
	./hack/update-openshift-certified.sh ${VERSION}


# password can be found at: https://connect.redhat.com/project/1366481/view
redhat-image-scan:
	docker pull ${IMG_REDHAT}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_operator} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-7f50cebe-122b-495a-a143-41426dfcb6c9/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-7f50cebe-122b-495a-a143-41426dfcb6c9/${IMG_VERSION_REDHAT}



CERTIFIED_OPERATORS_FOLDER=~/work/red-hat/certified-operators

#need to rename package in annotations.yaml to seldon-operator-certified
update_certified:
	cp -r bundle-certified/. ${CERTIFIED_OPERATORS_FOLDER}/operators/seldon-operator-certified/${VERSION}


GIT_REPO_URL=git@github.com:rafalskolasinski/certified-operators.git
BUNDLE_PATH=operators/seldon-operator-certified/${VERSION}
OPENSHIFT_PIPELINES_FOLDER=~/work/red-hat/operator-pipelines
run_certified_pipeline_basic:
	cd ${OPENSHIFT_PIPELINES_FOLDER} && tkn pipeline start operator-ci-pipeline \
		--param git_repo_url=${GIT_REPO_URL} \
		--param git_branch=${VERSION} \
		--param bundle_path=${BUNDLE_PATH} \
		--param env=stage \
		--workspace name=pipeline,volumeClaimTemplateFile=templates/workspace-template.yml \
		--workspace name=ssh-dir,secret=github-ssh-credentials \
		--showlog

# Run to pin images - presently will fail at later step but a new branch <version>-pinned will
# have been created which can be used in next step or fo rmanually creating PR
run_certified_pipeline_pinning:
	cd ${OPENSHIFT_PIPELINES_FOLDER} && tkn pipeline start operator-ci-pipeline \
		--param git_repo_url=${GIT_REPO_URL} \
		--param git_branch=${VERSION} \
		--param bundle_path=${BUNDLE_PATH} \
		--param env=stage \
		--param pin_digests=true \
		--param git_username=${GIT_USERNAME} \
		--param git_email=${GIT_EMAIL} \
		--param registry=registry.connect.redhat.com \
		--param image_namespace=seldonio \
		--workspace name=pipeline,volumeClaimTemplateFile=templates/workspace-template.yml \
		--workspace name=ssh-dir,secret=github-ssh-credentials \
		--workspace name=registry-credentials,secret=registry-dockerconfig-secret \
		--showlog

# Does not work at present
run_certified_pipeline_submit:
	cd ${OPENSHIFT_PIPELINES_FOLDER} && tkn pipeline start operator-ci-pipeline \
		--param git_repo_url=${GIT_REPO_URL} \
		--param git_branch=${VERSION}-pinned \
		--param bundle_path=${BUNDLE_PATH} \
		--param env=stage \
		--param pin_digests=false \
		--param git_username=${GIT_USERNAME} \
		--param git_email=${GIT_EMAIL} \
		--param upstream_repo_name=redhat-openshift-ecosystem/certified-operators \
		--param submit=true \
		--param image_namespace=seldonio \
		--workspace name=pipeline,volumeClaimTemplateFile=templates/workspace-template.yml \
		--workspace name=ssh-dir,secret=github-ssh-credentials \
		--showlog


# Run last to prepare for future release
PREV_VERSION=1.12.0
update_config:
	sed -i s#${PREV_VERSION}#${VERSION}# config/manifests/bases/seldon-operator.clusterserviceversion.yaml
