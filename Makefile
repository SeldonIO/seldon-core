

.PHONY: deploy-local
deploy-local:
	make -C scheduler start-all

.PHONY: undeploy-local
undeploy-local:
	make -C scheduler stop-all

.PHONY: deploy-k8s
deploy-k8s:
	kubectl create ns seldon-mesh || echo "namespace seldon-mesh already existing"
	kubectl create -f k8s/seldon-v2-crds.yaml
	sleep 1
	kubectl wait --for condition=established --timeout=60s crd/servers.mlops.seldon.io
	kubectl wait --for condition=established --timeout=60s crd/serverconfigs.mlops.seldon.io
	kubectl create -f k8s/seldon-v2-components.yaml
	kubectl create -f k8s/seldon-v2-servers.yaml

.PHONY: undeploy-k8s
undeploy-k8s:
	kubectl delete -f k8s/seldon-v2-servers.yaml
	kubectl delete -f k8s/seldon-v2-components.yaml
	kubectl delete -f k8s/seldon-v2-crds.yaml


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

