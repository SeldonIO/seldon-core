SDEP_NAME=image-classifier

readme:
	jupyter nbconvert README.ipynb --to markdown

train_model: install_dev
	cd src && python train_model.py

test:
	cd src && \
		pytest -s --verbose -W ignore --log-level=INFO 2>&1

install_dev:
	pip install -r src/requirements.txt

# INTEGRATION TESTS
install_integration_dev:
	pip install -r integration/requirements-dev.txt

test_integration:
	(cd integration && \
		pytest -s --verbose -W ignore --log-level=INFO 2>&1)

deploy_model:
	kubectl apply -f charts/SeldonDeployment-${SDEP_NAME}.json -n seldon
	sleep 1
	kubectl rollout status -n seldon deployment/${SDEP_NAME}-default-cbb4caa

delete_model:
	kubectl delete sdep ${SDEP_NAME}

helm_setup:
	helm repo add stable https://charts.helm.sh/stable
	helm repo add seldonio https://storage.googleapis.com/seldon-charts/
	helm repo add datawire https://www.getambassador.io
	helm repo update

install_ambassador:
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
	helm install seldon-core-operator \
		seldonio/seldon-core-operator \
		--namespace seldon \
		--version 1.0.0

create_namespaces:
	kubectl create namespace seldon || echo "Namespace seldon already exists"
	kubectl config set-context $$(kubectl config current-context) --namespace=seldon

kind_setup: helm_setup create_namespaces install_ambassador install_seldon

kind_create_cluster:
	kind create cluster --config integration/kind_config.yaml

kind_delete_cluster:
	kind delete cluster


