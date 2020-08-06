VERSION := $(shell cat VERSION)

readme:
	jupyter nbconvert README.ipynb --to markdown

make train_model: install_dev
	(cd src && python train_model.py)

make test:
	(cd src && \
		pytest -s --verbose -W ignore --log-level=INFO 2>&1)

make install_dev:
	pip install -r src/requirements.txt

# INTEGRATION TESTS
install_integration_dev:
	pip install -r integration/requirements-dev.txt

test_integration:
	(cd integration && \
		pytest -s --verbose -W ignore --log-level=INFO 2>&1)

deploy_model:
	helm install charts/seldon-model-server --namespace seldon --name seldon-model-server
	kubectl rollout status -n seldon deploy/seldon-model-server-default-cbb4caa

delete_model:
	helm delete seldon-model-server
	helm del --purge seldon-model-server || echo "Model deleted"

install_helm:
	kubectl -n kube-system create sa tiller
	kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
	helm init --service-account tiller
	kubectl rollout status deploy/tiller-deploy -n kube-system

install_ambassador:
	helm repo add datawire https://www.getambassador.io
	helm repo update

	helm install ambassador \
		datawire/ambassador \
		-f integration/ambassador_values.yaml \
		--set image.repository=quay.io/datawire/ambassador \
		--set crds.keep=false \
		--set enableAES=false \
		--namespace seldon \
		--set replicaCount=1 \
		--wait

install_seldon:
	helm install seldon-core-operator --name seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --namespace seldon
	kubectl rollout status -n seldon deployment/seldon-controller-manager

create_namespaces:
	kubectl create namespace seldon || echo "Namespace seldon already exists"
	kubectl config set-context $$(kubectl config current-context) --namespace=seldon

kind_setup: install_helm install_ambassador install_seldon create_namespaces deploy_model

kind_create_cluster:
	kind create cluster --config integration/kind_config.yaml

kind_delete_cluster:
	kind delete cluster


