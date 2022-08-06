SELDON_NAMESPACE?=seldon-mesh

.PHONY: deploy-local
deploy-local:
	make -C scheduler pull-all start-all

.PHONY: undeploy-local
undeploy-local:
	make -C scheduler stop-all

# Start raw deploy
.PHONY: deploy-k8s
deploy-k8s:
	kubectl create ns ${SELDON_NAMESPACE} || echo "namespace already existing"
	kubectl create -f k8s/yaml/seldon-v2-crds.yaml
	sleep 1
	kubectl wait --for condition=established --timeout=60s crd/servers.mlops.seldon.io
	kubectl wait --for condition=established --timeout=60s crd/serverconfigs.mlops.seldon.io
	kubectl create -f k8s/yaml/seldon-v2-components.yaml -n ${SELDON_NAMESPACE}
	kubectl create -f k8s/yaml/seldon-v2-servers.yaml -n ${SELDON_NAMESPACE}
# End raw deploy

# Start raw undeploy
.PHONY: undeploy-k8s
undeploy-k8s:
	kubectl delete --ignore-not-found=true -f k8s/yaml/seldon-v2-servers.yaml --wait=true -n ${SELDON_NAMESPACE}
	kubectl delete --ignore-not-found=true -f k8s/yaml/seldon-v2-components.yaml -n ${SELDON_NAMESPACE} 
	kubectl delete --ignore-not-found=true -f k8s/yaml/seldon-v2-crds.yaml
# End raw undeploy

#
# Dev
#

# use -W option for warnings as errors
docs_build_html:
	cd docs && \
		SPHINXOPTS=-W make html

docs_serve_html: docs_clean docs_build_html
	cd docs/build/html && \
		python -m http.server 8000

docs_clean:
	cd docs && \
		make clean

docs_install_requirements:
	cd docs && \
		pip install -r requirements-docs.txt

docs_dev_server: docs_clean
	cd docs && \
		sphinx-autobuild --host 0.0.0.0 source build/html

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
	# Add yaml files
	cp k8s/yaml/seldon-v2-components.yaml .release/
	cp k8s/yaml/seldon-v2-crds.yaml .release/
	cp k8s/yaml/seldon-v2-servers.yaml .release/
	# Build CLI
	make -C operator build-seldon && mv operator/bin/seldon .release/seldon-linux-amd64
