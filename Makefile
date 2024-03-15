SELDON_NAMESPACE?=seldon-mesh

.PHONY: deploy-local
deploy-local: undeploy-local
	make -C scheduler pull-all start-all

.PHONY: undeploy-local
undeploy-local:
	make -C scheduler stop-all

# Start raw deploy
.PHONY: deploy-k8s
deploy-k8s:
	kubectl create ns ${SELDON_NAMESPACE} || echo "namespace already existing"
	kubectl create -f k8s/yaml/crds.yaml
	sleep 1
	kubectl wait --for condition=established --timeout=60s crd/servers.mlops.seldon.io
	kubectl wait --for condition=established --timeout=60s crd/serverconfigs.mlops.seldon.io
	kubectl create -f k8s/yaml/components.yaml -n ${SELDON_NAMESPACE}
	kubectl create -f k8s/yaml/servers.yaml -n ${SELDON_NAMESPACE}
# End raw deploy

# Start raw undeploy
.PHONY: undeploy-k8s
undeploy-k8s:
	kubectl delete --ignore-not-found=true -f k8s/yaml/servers.yaml --wait=true -n ${SELDON_NAMESPACE}
	kubectl delete --ignore-not-found=true -f k8s/yaml/components.yaml -n ${SELDON_NAMESPACE}
	kubectl delete --ignore-not-found=true -f k8s/yaml/crds.yaml
# End raw undeploy

#
# Dev
#


init-go-modules:
	go work init || echo "go modules already initialized"
	go work use operator
	go work use scheduler
	go work use apis/go
	go work use components/tls
	go work use hodometer
	go work use tests/integration


# use -W option for warnings as errors
docs_build_html:
	cd docs && \
		SPHINXOPTS=-W make html

docs_serve_html: docs_clean docs_build_html
	cd docs/build/html && \
		python -m http.server 8000

docs_clean:
	cd docs && make clean

docs_install_requirements:
	cd docs && pip install -r requirements-docs.txt

docs_dev_server: docs_clean
	make -C docs dev_server

# can be installed here: https://github.com/client9/misspell#install
docs_spellcheck:
	misspell -w docs/source/contents


#
# Release
#

# This must be a top-level target that sets version across all components
set-versions:
	make -C k8s CUSTOM_IMAGE_TAG=${NEW_VERSION} create
	make -C k8s NEW_VERSION=${NEW_VERSION} set-chart-version

# This must be a top-level target that prepares all artifacts for release
prep-artifacts:
	rm .release -rf && mkdir .release
	# Package helm charts
	helm package k8s/helm-charts/seldon-core-v2-crds -d .release
	helm package k8s/helm-charts/seldon-core-v2-setup -d .release
	helm package k8s/helm-charts/seldon-core-v2-runtime -d .release
	helm package k8s/helm-charts/seldon-core-v2-servers -d .release
	helm package k8s/helm-charts/seldon-core-v2-certs -d .release
	# Add yaml files
	cp k8s/yaml/components.yaml .release/
	cp k8s/yaml/crds.yaml .release/
	cp k8s/yaml/servers.yaml .release/
	cp k8s/yaml/runtime.yaml .release/
	cp k8s/yaml/certs.yaml .release/
	# Build CLI
	make -C operator build-seldon && mv operator/bin/seldon .release/seldon-linux-amd64

#
# License
#

update-copyright:
	./hack/boilerplate.sh


install-go-license-tools:
	pip install -U \
		'git+https://github.com/SeldonIO/kubeflow-testing#egg=go-license-tools&subdirectory=py/kubeflow/testing/go-license-tools'

.PHONY: update-3rd-party-licenses
update-3rd-party-licenses:
	make -C scheduler licenses
	make -C operator licenses
	make -C hodometer licenses
	make -C components/tls licenses
	make -C tests/integration licenses
	make -C scheduler/data-flow licenses
	{ (cat scheduler/licenses/license_info.csv operator/licenses/license_info.csv hodometer/licenses/license_info.csv components/tls/licenses/license_info.csv tests/integration/licenses/license_info.csv | cut -d, -f3) ; (cat scheduler/data-flow/licenses/dependency-license.json | jq .dependencies[].licenses[0].name) } | sed 's/\"//g' | sort | uniq -c > licenses/3rd-party-summary.txt
